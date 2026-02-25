package cli

import (
	"fmt"
	"testing"
)

func TestSilentError(t *testing.T) {
	t.Parallel()

	err := NewSilentError(fmt.Errorf("test error"))
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %q", err.Error())
	}
	if !IsSilentError(err) {
		t.Error("expected IsSilentError to return true")
	}
}

func TestIsSilentError_RegularError(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("regular error")
	if IsSilentError(err) {
		t.Error("expected IsSilentError to return false for regular error")
	}
}
