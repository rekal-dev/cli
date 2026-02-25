package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// EnsureGitRoot resolves and returns the git repository root.
// Returns an error if the current directory is not inside a git repository.
func EnsureGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository; run from a git repo")
	}
	return strings.TrimSpace(string(out)), nil
}

// EnsureInitDone checks that Rekal has been initialized in the given git root.
// It verifies that .rekal/ exists and contains the expected database files.
func EnsureInitDone(gitRoot string) error {
	rekalDir := filepath.Join(gitRoot, ".rekal")
	info, err := os.Stat(rekalDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("rekal not initialized; run 'rekal init' in a git repository")
	}
	dataDB := filepath.Join(rekalDir, "data.db")
	if _, err := os.Stat(dataDB); err != nil {
		return fmt.Errorf("rekal not initialized; run 'rekal init' in a git repository")
	}
	return nil
}

// RekalDir returns the path to .rekal/ within the given git root.
func RekalDir(gitRoot string) string {
	return filepath.Join(gitRoot, ".rekal")
}
