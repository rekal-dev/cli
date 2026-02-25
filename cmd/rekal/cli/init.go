package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rekal-dev/cli/cmd/rekal/cli/db"
	"github.com/spf13/cobra"
)

const rekalHookMarker = "# managed by rekal"

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Rekal in the current git repository",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true

			gitRoot, err := EnsureGitRoot()
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return NewSilentError(err)
			}

			rekalDir := RekalDir(gitRoot)

			// Re-run = clean + reinit.
			if _, err := os.Stat(rekalDir); err == nil {
				if err := runClean(gitRoot); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), err)
					return NewSilentError(err)
				}
			}

			// Create .rekal/ directory.
			if err := os.MkdirAll(rekalDir, 0o755); err != nil {
				return fmt.Errorf("create .rekal/: %w", err)
			}

			// Create data DB with schema.
			dataDB, err := db.OpenData(gitRoot)
			if err != nil {
				return fmt.Errorf("create data DB: %w", err)
			}
			if err := db.InitDataSchema(dataDB); err != nil {
				dataDB.Close()
				return fmt.Errorf("init data schema: %w", err)
			}
			dataDB.Close()

			// Create index DB with schema.
			indexDB, err := db.OpenIndex(gitRoot)
			if err != nil {
				return fmt.Errorf("create index DB: %w", err)
			}
			if err := db.InitIndexSchema(indexDB); err != nil {
				indexDB.Close()
				return fmt.Errorf("init index schema: %w", err)
			}
			indexDB.Close()

			// Ensure .rekal/ is in .gitignore.
			if err := ensureGitignore(gitRoot); err != nil {
				return fmt.Errorf("update .gitignore: %w", err)
			}

			// Install hook stubs.
			if err := installHooks(gitRoot); err != nil {
				return fmt.Errorf("install hooks: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Rekal initialized.")
			return nil
		},
	}
}

func ensureGitignore(gitRoot string) error {
	gitignorePath := filepath.Join(gitRoot, ".gitignore")
	entry := ".rekal/"

	data, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(data)
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == entry {
			return nil // already present
		}
	}

	// Append .rekal/ to .gitignore.
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add newline before entry if file doesn't end with one.
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}
	_, err = f.WriteString(entry + "\n")
	return err
}

func installHooks(gitRoot string) error {
	hooksDir := filepath.Join(gitRoot, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return err
	}

	postCommit := filepath.Join(hooksDir, "post-commit")
	if err := writeHook(postCommit, "#!/bin/sh\n"+rekalHookMarker+"\nrekal checkpoint\n"); err != nil {
		return fmt.Errorf("post-commit hook: %w", err)
	}

	prePush := filepath.Join(hooksDir, "pre-push")
	if err := writeHook(prePush, "#!/bin/sh\n"+rekalHookMarker+"\nrekal push\n"); err != nil {
		return fmt.Errorf("pre-push hook: %w", err)
	}

	return nil
}

func writeHook(path, content string) error {
	// If a hook already exists and is not ours, leave it alone.
	existing, err := os.ReadFile(path)
	if err == nil && !strings.Contains(string(existing), rekalHookMarker) {
		return nil // not our hook; do not overwrite
	}
	return os.WriteFile(path, []byte(content), 0o755)
}
