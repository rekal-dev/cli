package session

import (
	"testing"
)

const geminiFixtureJSON = `{
	"sessionId": "gemini-001",
	"startTime": "2025-06-01T10:00:00Z",
	"messages": [
		{
			"type": "user",
			"content": "Add a login page"
		},
		{
			"type": "gemini",
			"content": "I'll create a login page for you.",
			"toolCalls": [
				{
					"name": "write_file",
					"args": {"file_path": "src/login.tsx", "content": "<Login />"}
				},
				{
					"name": "run_command",
					"args": {"command": "npm run build"}
				}
			]
		},
		{
			"type": "user",
			"content": [{"text": "Looks good, now add tests"}]
		}
	]
}`

func TestGeminiAdapter_Parse(t *testing.T) {
	t.Parallel()

	adapter := &GeminiAdapter{}

	tmpFile := t.TempDir() + "/session-001.json"
	if err := writeTestFile(tmpFile, geminiFixtureJSON); err != nil {
		t.Fatal(err)
	}

	payload, err := adapter.Parse(SessionRef{Path: tmpFile})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if payload.Source != "gemini" {
		t.Errorf("Source = %q, want gemini", payload.Source)
	}
	if payload.SessionID != "gemini-001" {
		t.Errorf("SessionID = %q, want gemini-001", payload.SessionID)
	}

	// 3 turns: user + gemini + user (array content).
	if len(payload.Turns) != 3 {
		t.Fatalf("len(Turns) = %d, want 3", len(payload.Turns))
	}
	if payload.Turns[0].Role != "human" || payload.Turns[0].Content != "Add a login page" {
		t.Errorf("Turns[0] = %+v", payload.Turns[0])
	}
	if payload.Turns[1].Role != "assistant" || payload.Turns[1].Content != "I'll create a login page for you." {
		t.Errorf("Turns[1] = %+v", payload.Turns[1])
	}
	if payload.Turns[2].Role != "human" || payload.Turns[2].Content != "Looks good, now add tests" {
		t.Errorf("Turns[2] = %+v", payload.Turns[2])
	}

	// 2 tool calls from the gemini message.
	if len(payload.ToolCalls) != 2 {
		t.Fatalf("len(ToolCalls) = %d, want 2", len(payload.ToolCalls))
	}
	if payload.ToolCalls[0].Tool != "write_file" || payload.ToolCalls[0].Path != "src/login.tsx" {
		t.Errorf("ToolCalls[0] = %+v", payload.ToolCalls[0])
	}
	if payload.ToolCalls[1].Tool != "run_command" || payload.ToolCalls[1].CmdPrefix != "npm run build" {
		t.Errorf("ToolCalls[1] = %+v", payload.ToolCalls[1])
	}
}

func TestGeminiProjectHash(t *testing.T) {
	t.Parallel()

	// Just verify it returns a 64-char hex string.
	hash := geminiProjectHash("/Users/frank/projects/rekal")
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Same input should produce same hash.
	hash2 := geminiProjectHash("/Users/frank/projects/rekal")
	if hash != hash2 {
		t.Errorf("hash not deterministic: %q != %q", hash, hash2)
	}

	// Different input should produce different hash.
	hash3 := geminiProjectHash("/Users/frank/projects/other")
	if hash == hash3 {
		t.Error("different paths produced same hash")
	}
}
