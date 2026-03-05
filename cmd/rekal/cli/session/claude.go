package session

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ClaudeAdapter discovers and parses Claude Code sessions from JSONL files.
type ClaudeAdapter struct{}

func (a *ClaudeAdapter) Name() string { return "claude" }

func (a *ClaudeAdapter) Discover(repoPath string) ([]SessionRef, error) {
	sessionDir := FindSessionDir(repoPath)
	if sessionDir == "" {
		return nil, nil
	}

	files, err := FindSessionFiles(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	refs := make([]SessionRef, len(files))
	for i, f := range files {
		refs[i] = SessionRef{Path: f}
	}
	return refs, nil
}

func (a *ClaudeAdapter) Parse(ref SessionRef) (*SessionPayload, error) {
	data, err := os.ReadFile(ref.Path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	payload, err := ParseTranscript(data)
	if err != nil {
		return nil, err
	}
	payload.Source = "claude"
	payload.CapturedAt = time.Now().UTC()
	return payload, nil
}

// FindSessionDir returns the Claude Code session directory for the given repo path.
// Returns ~/.claude/projects/<sanitized-repo-path>/.
func FindSessionDir(repoPath string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	sanitized := SanitizeRepoPath(repoPath)
	return filepath.Join(home, ".claude", "projects", sanitized)
}

// FindSessionFiles lists all .jsonl session files in the given directory.
func FindSessionFiles(sessionDir string) ([]string, error) {
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".jsonl") {
			files = append(files, filepath.Join(sessionDir, e.Name()))
		}
	}
	return files, nil
}
