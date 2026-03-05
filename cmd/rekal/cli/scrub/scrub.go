package scrub

import (
	"github.com/rekal-dev/rekal-cli/cmd/rekal/cli/session"
)

// Scrub applies secret redaction and path anonymization to a SessionPayload
// in place. Call this after session.ParseTranscript and before DB insertion.
func Scrub(payload *session.SessionPayload) {
	if payload == nil {
		return
	}

	for i := range payload.Turns {
		payload.Turns[i].Content = RedactText(payload.Turns[i].Content)
		payload.Turns[i].Content = AnonymizeText(payload.Turns[i].Content)
	}

	for i := range payload.ToolCalls {
		payload.ToolCalls[i].Path = AnonymizePath(payload.ToolCalls[i].Path)
		payload.ToolCalls[i].CmdPrefix = RedactText(payload.ToolCalls[i].CmdPrefix)
		payload.ToolCalls[i].CmdPrefix = AnonymizeText(payload.ToolCalls[i].CmdPrefix)
	}
}
