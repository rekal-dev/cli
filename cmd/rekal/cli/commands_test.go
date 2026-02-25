package cli

import (
	"strings"
	"testing"
)

func TestStubCommands_RequirePreconditions(t *testing.T) {
	// These commands require git repo + init; test that they fail outside a git repo.
	commands := []string{"checkpoint", "push", "index", "log", "sync"}

	for _, name := range commands {
		name := name
		t.Run(name+"_not_git_repo", func(t *testing.T) {
			dir := t.TempDir()
			_, stderr, err := executeCmd(t, dir, name)
			if err == nil {
				t.Fatalf("%s should fail outside git repo", name)
			}
			if !strings.Contains(stderr, "not a git repository") {
				t.Errorf("%s: expected git repo error, got: %q", name, stderr)
			}
		})
	}
}

func TestStubCommands_RequireInit(t *testing.T) {
	commands := []string{"checkpoint", "push", "index", "log", "sync"}

	for _, name := range commands {
		name := name
		t.Run(name+"_not_initialized", func(t *testing.T) {
			dir := testGitRepo(t)
			_, stderr, err := executeCmd(t, dir, name)
			if err == nil {
				t.Fatalf("%s should fail without init", name)
			}
			if !strings.Contains(stderr, "rekal not initialized") {
				t.Errorf("%s: expected init error, got: %q", name, stderr)
			}
		})
	}
}

func TestStubCommands_NotYetImplemented(t *testing.T) {
	commands := []string{"checkpoint", "push", "index", "log", "sync"}

	for _, name := range commands {
		name := name
		t.Run(name, func(t *testing.T) {
			dir := testGitRepo(t)

			// Init the repo first.
			_, _, err := executeCmd(t, dir, "init")
			if err != nil {
				t.Fatalf("init: %v", err)
			}

			_, stderr, err := executeCmd(t, dir, name)
			if err != nil {
				t.Fatalf("%s should succeed (stub): %v", name, err)
			}
			if !strings.Contains(stderr, "not yet implemented") {
				t.Errorf("%s: expected 'not yet implemented', got: %q", name, stderr)
			}
		})
	}
}

func TestQuery_RequiresArg(t *testing.T) {
	dir := testGitRepo(t)
	_, _, err := executeCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	_, _, err = executeCmd(t, dir, "query")
	if err == nil {
		t.Fatal("query without SQL arg should fail")
	}
}

func TestQuery_NotYetImplemented(t *testing.T) {
	dir := testGitRepo(t)
	_, _, err := executeCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	_, stderr, err := executeCmd(t, dir, "query", "SELECT 1")
	if err != nil {
		t.Fatalf("query should succeed (stub): %v", err)
	}
	if !strings.Contains(stderr, "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got: %q", stderr)
	}
}

func TestRecall_NotYetImplemented(t *testing.T) {
	dir := testGitRepo(t)
	_, _, err := executeCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	_, stderr, err := executeCmd(t, dir, "--file", "foo", "JWT")
	if err != nil {
		t.Fatalf("recall should succeed (stub): %v", err)
	}
	if !strings.Contains(stderr, "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got: %q", stderr)
	}
}

func TestRecall_NoArgsShowsHelp(t *testing.T) {
	dir := testGitRepo(t)

	stdout, _, err := executeCmd(t, dir)
	if err != nil {
		t.Fatalf("root with no args: %v", err)
	}
	if !strings.Contains(stdout, "Rekal") {
		t.Errorf("expected help output, got: %q", stdout)
	}
}
