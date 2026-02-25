package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rekal-dev/cli/cmd/rekal/cli/codec"
	"github.com/rekal-dev/cli/cmd/rekal/cli/db"
	"github.com/rekal-dev/cli/cmd/rekal/cli/session"
	"github.com/spf13/cobra"
)

func newCheckpointCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "checkpoint",
		Short: "Capture the current session after a commit",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true

			gitRoot, err := EnsureGitRoot()
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return NewSilentError(err)
			}
			if err := EnsureInitDone(gitRoot); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return NewSilentError(err)
			}

			return runCheckpoint(cmd, gitRoot)
		},
	}
}

func runCheckpoint(cmd *cobra.Command, gitRoot string) error {
	// Find session directory for this repo.
	sessionDir := session.FindSessionDir(gitRoot)
	if sessionDir == "" {
		return nil
	}

	files, err := session.FindSessionFiles(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("find session files: %w", err)
	}
	if len(files) == 0 {
		return nil
	}

	// Open data DB.
	dataDB, err := db.OpenData(gitRoot)
	if err != nil {
		return fmt.Errorf("open data DB: %w", err)
	}
	defer dataDB.Close()

	email := gitConfigValue("user.email")
	entropy := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	newID := func() string {
		return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
	}

	// Load existing wire format files from orphan branch.
	branch := rekalBranchName()
	bodyData := gitShowFile(gitRoot, branch, "rekal.body")
	dictData := gitShowFile(gitRoot, branch, "dict.bin")

	// Parse existing dict and body, or create new ones.
	dict := codec.NewDict()
	if len(dictData) > 0 {
		loaded, err := codec.LoadDict(dictData)
		if err == nil {
			dict = loaded
		}
	}
	body := bodyData
	if len(body) == 0 {
		body = codec.NewBody()
	}

	enc, err := codec.NewEncoder()
	if err != nil {
		return fmt.Errorf("create encoder: %w", err)
	}
	defer enc.Close()

	var sessionIDs []string
	var sessionRefs []uint64
	var inserted int

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		if len(data) == 0 {
			continue
		}

		hash := sha256Hex(data)

		exists, err := db.SessionExistsByHash(dataDB, hash)
		if err != nil {
			return fmt.Errorf("dedup check: %w", err)
		}
		if exists {
			continue
		}

		payload, err := session.ParseTranscript(data)
		if err != nil {
			continue
		}

		if len(payload.Turns) == 0 && len(payload.ToolCalls) == 0 {
			continue
		}

		sessionID := newID()
		capturedAt := time.Now().UTC()

		// Insert session into DuckDB.
		if err := db.InsertSession(
			dataDB, sessionID, "", hash,
			payload.ActorType, payload.AgentID, email, payload.Branch, capturedAt.Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("insert session: %w", err)
		}

		// Insert turns into DuckDB.
		for i, t := range payload.Turns {
			ts := ""
			if !t.Timestamp.IsZero() {
				ts = t.Timestamp.UTC().Format(time.RFC3339)
			}
			if err := db.InsertTurn(dataDB, newID(), sessionID, i, t.Role, t.Content, ts); err != nil {
				return fmt.Errorf("insert turn: %w", err)
			}
		}

		// Insert tool calls into DuckDB.
		for i, tc := range payload.ToolCalls {
			if err := db.InsertToolCall(dataDB, newID(), sessionID, i, tc.Tool, tc.Path, tc.CmdPrefix); err != nil {
				return fmt.Errorf("insert tool_call: %w", err)
			}
		}

		// Encode session frame for wire format.
		sessRef := dict.LookupOrAdd(codec.NSSessions, sessionID)
		emailRef := dict.LookupOrAdd(codec.NSEmails, email)
		branchRef := uint64(0)
		if payload.Branch != "" {
			branchRef = dict.LookupOrAdd(codec.NSBranches, payload.Branch)
		}

		actorType := codec.ActorHuman
		agentIDRef := uint64(0)
		if payload.ActorType == "agent" {
			actorType = codec.ActorAgent
			if payload.AgentID != "" {
				agentIDRef = dict.LookupOrAdd(codec.NSEmails, payload.AgentID)
			}
		}

		sf := &codec.SessionFrame{
			SessionRef: sessRef,
			CapturedAt: capturedAt,
			EmailRef:   emailRef,
			ActorType:  actorType,
			AgentIDRef: agentIDRef,
		}

		// Build turn records with delta timestamps.
		for i, t := range payload.Turns {
			role := codec.RoleHuman
			if t.Role == "assistant" {
				role = codec.RoleAssistant
			}
			var tsDelta uint64
			if i > 0 && !t.Timestamp.IsZero() && !payload.Turns[i-1].Timestamp.IsZero() {
				delta := t.Timestamp.Sub(payload.Turns[i-1].Timestamp)
				if delta > 0 {
					tsDelta = uint64(delta.Seconds())
				}
			}
			sf.Turns = append(sf.Turns, codec.TurnRecord{
				Role:      role,
				TsDelta:   tsDelta,
				BranchRef: branchRef,
				Text:      t.Content,
			})
		}

		// Build tool call records.
		for _, tc := range payload.ToolCalls {
			toolCode := codec.ToolCode(tc.Tool)
			tcr := codec.ToolCallRecord{
				Tool: toolCode,
			}
			if tc.Path == "" {
				tcr.PathFlag = codec.PathNull
			} else {
				pathRef := dict.LookupOrAdd(codec.NSPaths, tc.Path)
				tcr.PathFlag = codec.PathDictRef
				tcr.PathRef = pathRef
			}
			tcr.CmdPrefix = tc.CmdPrefix
			sf.ToolCalls = append(sf.ToolCalls, tcr)
		}

		frame := enc.EncodeSessionFrame(sf)
		body = codec.AppendFrame(body, frame)

		sessionIDs = append(sessionIDs, sessionID)
		sessionRefs = append(sessionRefs, sessRef)
		inserted++
	}

	if inserted == 0 {
		return nil
	}

	// Get git state for checkpoint.
	gitSHA := gitHeadSHA(gitRoot)
	gitBranch := gitCurrentBranch(gitRoot)
	filesTouched := gitFilesChanged(gitRoot)

	// Build checkpoint frame.
	cpBranchRef := dict.LookupOrAdd(codec.NSBranches, gitBranch)
	cpEmailRef := dict.LookupOrAdd(codec.NSEmails, email)

	var fileRecords []codec.FileTouchedRecord
	for _, ft := range filesTouched {
		parts := strings.SplitN(ft, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		changeType := byte('M')
		switch parts[0] {
		case "A":
			changeType = codec.ChangeAdded
		case "M":
			changeType = codec.ChangeModified
		case "D":
			changeType = codec.ChangeDeleted
		case "R":
			changeType = codec.ChangeRenamed
		default:
			if len(parts[0]) > 0 {
				changeType = parts[0][0]
			}
		}
		pathRef := dict.LookupOrAdd(codec.NSPaths, parts[1])
		fileRecords = append(fileRecords, codec.FileTouchedRecord{
			PathRef:    pathRef,
			ChangeType: changeType,
		})
	}

	now := time.Now().UTC()
	cf := &codec.CheckpointFrame{
		GitSHA:      gitSHA,
		BranchRef:   cpBranchRef,
		EmailRef:    cpEmailRef,
		Timestamp:   now,
		ActorType:   codec.ActorHuman,
		SessionRefs: sessionRefs,
		Files:       fileRecords,
	}
	body = codec.AppendFrame(body, enc.EncodeCheckpointFrame(cf))

	// Count existing frames for meta.
	existingFrames, _ := codec.ScanFrames(body)
	nFrames := uint32(len(existingFrames))

	// Encode meta frame.
	mf := &codec.MetaFrame{
		FormatVersion: 0x01,
		EmailRef:      cpEmailRef,
		CheckpointSHA: strings.Repeat("0", 40), // placeholder, updated after commit
		Timestamp:     now,
		NSessions:     uint32(dict.Len(codec.NSSessions)),
		NCheckpoints:  0,           // will be set after we know the count
		NFrames:       nFrames + 1, // +1 for this meta frame
		NDictEntries:  uint32(dict.TotalEntries()),
	}
	body = codec.AppendFrame(body, enc.EncodeMetaFrame(mf))

	// Commit wire format files to orphan branch.
	commitSHA, err := commitWireFormat(gitRoot, body, dict.Encode())
	if err != nil {
		return fmt.Errorf("commit to rekal branch: %w", err)
	}

	// Insert checkpoint and junction rows into DuckDB.
	if err := db.InsertCheckpoint(dataDB, commitSHA, gitSHA, gitBranch, email, now.Format(time.RFC3339), "human", ""); err != nil {
		return fmt.Errorf("insert checkpoint: %w", err)
	}
	for _, ft := range filesTouched {
		parts := strings.SplitN(ft, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		if err := db.InsertFileTouched(dataDB, newID(), commitSHA, parts[1], parts[0]); err != nil {
			return fmt.Errorf("insert file_touched: %w", err)
		}
	}
	for _, sid := range sessionIDs {
		if err := db.InsertCheckpointSession(dataDB, commitSHA, sid); err != nil {
			return fmt.Errorf("insert checkpoint_session: %w", err)
		}
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "rekal: %d session(s) captured\n", inserted)
	return nil
}

// commitWireFormat commits rekal.body and dict.bin to the orphan branch.
// Returns the new commit SHA.
func commitWireFormat(gitRoot string, bodyData, dictData []byte) (string, error) {
	branch := rekalBranchName()

	// Get the current tip of the orphan branch.
	parentOut, err := exec.Command("git", "-C", gitRoot, "rev-parse", branch).Output()
	if err != nil {
		return "", fmt.Errorf("resolve branch %s: %w", branch, err)
	}
	parent := strings.TrimSpace(string(parentOut))

	bodyHash, err := gitHashObject(gitRoot, bodyData)
	if err != nil {
		return "", fmt.Errorf("hash rekal.body: %w", err)
	}
	dictHash, err := gitHashObject(gitRoot, dictData)
	if err != nil {
		return "", fmt.Errorf("hash dict.bin: %w", err)
	}

	treeEntry := fmt.Sprintf("100644 blob %s\tdict.bin\n100644 blob %s\trekal.body\n", dictHash, bodyHash)
	mktreeCmd := exec.Command("git", "-C", gitRoot, "mktree")
	mktreeCmd.Stdin = strings.NewReader(treeEntry)
	treeOut, err := mktreeCmd.Output()
	if err != nil {
		return "", fmt.Errorf("mktree: %w", err)
	}
	treeHash := strings.TrimSpace(string(treeOut))

	commitOut, err := exec.Command("git", "-C", gitRoot,
		"commit-tree", treeHash, "-p", parent, "-m", "rekal: checkpoint",
	).Output()
	if err != nil {
		return "", fmt.Errorf("commit-tree: %w", err)
	}
	commitSHA := strings.TrimSpace(string(commitOut))

	if err := exec.Command("git", "-C", gitRoot, "update-ref", "refs/heads/"+branch, commitSHA).Run(); err != nil {
		return "", fmt.Errorf("update-ref: %w", err)
	}

	return commitSHA, nil
}

func gitHeadSHA(gitRoot string) string {
	out, err := exec.Command("git", "-C", gitRoot, "rev-parse", "HEAD").Output()
	if err != nil {
		return strings.Repeat("0", 40)
	}
	return strings.TrimSpace(string(out))
}

func gitCurrentBranch(gitRoot string) string {
	out, err := exec.Command("git", "-C", gitRoot, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func gitFilesChanged(gitRoot string) []string {
	out, err := exec.Command("git", "-C", gitRoot, "diff", "--name-status", "HEAD~1", "HEAD").Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

// gitShowFile reads a file from a git ref. Returns nil if not found.
func gitShowFile(gitRoot, ref, path string) []byte {
	out, err := exec.Command("git", "-C", gitRoot, "show", ref+":"+path).Output()
	if err != nil {
		return nil
	}
	return out
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
