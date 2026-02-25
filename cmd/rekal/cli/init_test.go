package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testGitRepo creates a temp dir with git init and returns the path.
func testGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	return dir
}

// executeCmd runs the root command with given args from the given directory,
// capturing stdout and stderr.
func executeCmd(t *testing.T, dir string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewRootCmd()
	cmd.SetArgs(args)

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	oldDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldDir) }()

	execErr := cmd.Execute()
	return outBuf.String(), errBuf.String(), execErr
}

func TestInit_CreatesRekalDir(t *testing.T) {
	dir := testGitRepo(t)

	stdout, _, err := executeCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if !strings.Contains(stdout, "Rekal initialized.") {
		t.Errorf("expected success message, got: %q", stdout)
	}

	// Verify .rekal/ exists with databases.
	rekalDir := filepath.Join(dir, ".rekal")
	if _, err := os.Stat(rekalDir); err != nil {
		t.Error(".rekal/ should exist after init")
	}
	if _, err := os.Stat(filepath.Join(rekalDir, "data.db")); err != nil {
		t.Error("data.db should exist after init")
	}
	if _, err := os.Stat(filepath.Join(rekalDir, "index.db")); err != nil {
		t.Error("index.db should exist after init")
	}
}

func TestInit_GitignoreEntry(t *testing.T) {
	dir := testGitRepo(t)

	_, _, err := executeCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(data), ".rekal/") {
		t.Error(".gitignore should contain .rekal/")
	}
}

func TestInit_InstallsHooks(t *testing.T) {
	dir := testGitRepo(t)

	_, _, err := executeCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	postCommit := filepath.Join(dir, ".git", "hooks", "post-commit")
	data, err := os.ReadFile(postCommit)
	if err != nil {
		t.Fatalf("read post-commit: %v", err)
	}
	if !strings.Contains(string(data), rekalHookMarker) {
		t.Error("post-commit should contain rekal marker")
	}
	if !strings.Contains(string(data), "rekal checkpoint") {
		t.Error("post-commit should call rekal checkpoint")
	}
}

func TestInit_Reinit(t *testing.T) {
	dir := testGitRepo(t)

	// First init.
	_, _, err := executeCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("first init: %v", err)
	}

	// Re-init should succeed (clean + reinit).
	stdout, _, err := executeCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("reinit: %v", err)
	}
	if !strings.Contains(stdout, "Rekal initialized.") {
		t.Errorf("reinit should print success message")
	}
}

func TestInit_NotGitRepo(t *testing.T) {
	dir := t.TempDir()

	_, stderr, err := executeCmd(t, dir, "init")
	if err == nil {
		t.Fatal("init outside git repo should fail")
	}
	if !strings.Contains(stderr, "not a git repository") {
		t.Errorf("expected git repo error, got: %q", stderr)
	}
}

func TestClean_RemovesRekalDir(t *testing.T) {
	dir := testGitRepo(t)

	// Init first.
	_, _, err := executeCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Clean.
	stdout, _, err := executeCmd(t, dir, "clean")
	if err != nil {
		t.Fatalf("clean: %v", err)
	}
	if !strings.Contains(stdout, "Rekal cleaned.") {
		t.Errorf("expected clean message, got: %q", stdout)
	}

	rekalDir := filepath.Join(dir, ".rekal")
	if _, err := os.Stat(rekalDir); !os.IsNotExist(err) {
		t.Error(".rekal/ should not exist after clean")
	}
}

func TestClean_RemovesHooks(t *testing.T) {
	dir := testGitRepo(t)

	_, _, err := executeCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	_, _, err = executeCmd(t, dir, "clean")
	if err != nil {
		t.Fatalf("clean: %v", err)
	}

	postCommit := filepath.Join(dir, ".git", "hooks", "post-commit")
	if _, err := os.Stat(postCommit); !os.IsNotExist(err) {
		t.Error("post-commit hook should be removed after clean")
	}
}

func TestClean_Idempotent(t *testing.T) {
	dir := testGitRepo(t)

	// Clean without init — should succeed (idempotent).
	stdout, _, err := executeCmd(t, dir, "clean")
	if err != nil {
		t.Fatalf("clean (no init): %v", err)
	}
	if !strings.Contains(stdout, "Rekal cleaned.") {
		t.Errorf("expected clean message")
	}
}

func TestClean_NotGitRepo(t *testing.T) {
	dir := t.TempDir()

	_, stderr, err := executeCmd(t, dir, "clean")
	if err == nil {
		t.Fatal("clean outside git repo should fail")
	}
	if !strings.Contains(stderr, "not a git repository") {
		t.Errorf("expected git repo error, got: %q", stderr)
	}
}
