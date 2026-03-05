package session

import (
	"testing"
)

const codexFixtureJSONL = `{"type":"session_meta","session_id":"codex-001","timestamp":"2025-06-01T10:00:00Z","payload":{"cwd":"/tmp/repo","git":{"branch":"feature/auth"}}}
{"type":"event_msg","session_id":"codex-001","timestamp":"2025-06-01T10:00:01Z","payload":{"type":"user_message","message":"Add JWT authentication"}}
{"type":"event_msg","session_id":"codex-001","timestamp":"2025-06-01T10:00:05Z","payload":{"type":"agent_message","message":"I'll add JWT auth to the project."}}
{"type":"response_item","session_id":"codex-001","timestamp":"2025-06-01T10:00:06Z","payload":{"type":"function_call","name":"write_file","arguments":"{\"file_path\":\"src/auth.ts\",\"content\":\"export function verify() {}\"}"}}
{"type":"response_item","session_id":"codex-001","timestamp":"2025-06-01T10:00:07Z","payload":{"type":"function_call","name":"shell","arguments":"{\"command\":\"npm test && echo done\"}"}}
`

func TestCodexAdapter_Parse(t *testing.T) {
	t.Parallel()

	adapter := &CodexAdapter{}

	// Write fixture to temp file.
	tmpFile := t.TempDir() + "/session.jsonl"
	if err := writeTestFile(tmpFile, codexFixtureJSONL); err != nil {
		t.Fatal(err)
	}

	payload, err := adapter.Parse(SessionRef{Path: tmpFile})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if payload.Source != "codex" {
		t.Errorf("Source = %q, want codex", payload.Source)
	}
	if payload.SessionID != "codex-001" {
		t.Errorf("SessionID = %q, want codex-001", payload.SessionID)
	}
	if payload.Branch != "feature/auth" {
		t.Errorf("Branch = %q, want feature/auth", payload.Branch)
	}

	// 2 turns: user + agent message.
	if len(payload.Turns) != 2 {
		t.Fatalf("len(Turns) = %d, want 2", len(payload.Turns))
	}
	if payload.Turns[0].Role != "human" || payload.Turns[0].Content != "Add JWT authentication" {
		t.Errorf("Turns[0] = %+v", payload.Turns[0])
	}
	if payload.Turns[1].Role != "assistant" || payload.Turns[1].Content != "I'll add JWT auth to the project." {
		t.Errorf("Turns[1] = %+v", payload.Turns[1])
	}

	// 2 tool calls: write_file + shell.
	if len(payload.ToolCalls) != 2 {
		t.Fatalf("len(ToolCalls) = %d, want 2", len(payload.ToolCalls))
	}
	if payload.ToolCalls[0].Tool != "write_file" || payload.ToolCalls[0].Path != "src/auth.ts" {
		t.Errorf("ToolCalls[0] = %+v", payload.ToolCalls[0])
	}
	if payload.ToolCalls[1].Tool != "shell" || payload.ToolCalls[1].CmdPrefix != "npm test && echo done" {
		t.Errorf("ToolCalls[1] = %+v", payload.ToolCalls[1])
	}
}

func TestCodexAdapter_Parse_Empty(t *testing.T) {
	t.Parallel()

	adapter := &CodexAdapter{}
	tmpFile := t.TempDir() + "/empty.jsonl"
	if err := writeTestFile(tmpFile, ""); err != nil {
		t.Fatal(err)
	}

	payload, err := adapter.Parse(SessionRef{Path: tmpFile})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if payload != nil && (len(payload.Turns) != 0 || len(payload.ToolCalls) != 0) {
		t.Errorf("expected empty payload for empty file")
	}
}

func TestCodexSessionMatchesRepo(t *testing.T) {
	t.Parallel()

	tmpFile := t.TempDir() + "/session.jsonl"
	if err := writeTestFile(tmpFile, codexFixtureJSONL); err != nil {
		t.Fatal(err)
	}

	if !codexSessionMatchesRepo(tmpFile, "/tmp/repo") {
		t.Error("expected session to match /tmp/repo")
	}
	if codexSessionMatchesRepo(tmpFile, "/other/repo") {
		t.Error("expected session to NOT match /other/repo")
	}
}
