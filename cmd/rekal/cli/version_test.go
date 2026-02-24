package cli

import (
	"testing"
)

func TestVersionNonEmpty(t *testing.T) {
	t.Parallel()
	if Version == "" {
		t.Error("Version should not be empty")
	}
}
