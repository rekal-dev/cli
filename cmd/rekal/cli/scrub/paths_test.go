package scrub

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func hashedUser(username string) string {
	h := sha256.Sum256([]byte(username))
	return fmt.Sprintf("user_%x", h[:4])
}

func TestAnonymizeMacPath(t *testing.T) {
	t.Parallel()
	a := newAnonymizer("frank")
	input := "/Users/frank/projects/rekal/main.go"
	got := a.anonymizePath(input)
	want := "/Users/" + hashedUser("frank") + "/projects/rekal/main.go"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAnonymizeLinuxPath(t *testing.T) {
	t.Parallel()
	a := newAnonymizer("alice")
	input := "/home/alice/src/project/file.py"
	got := a.anonymizePath(input)
	want := "/home/" + hashedUser("alice") + "/src/project/file.py"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAnonymizeHyphenEncodedMac(t *testing.T) {
	t.Parallel()
	a := newAnonymizer("frank")
	input := "-Users-frank-projects-rekal"
	got := a.anonymizeText(input)
	want := "-Users-" + hashedUser("frank") + "-projects-rekal"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAnonymizeHyphenEncodedLinux(t *testing.T) {
	t.Parallel()
	a := newAnonymizer("bob")
	input := "-home-bob-code-app"
	got := a.anonymizeText(input)
	want := "-home-" + hashedUser("bob") + "-code-app"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAnonymizeTextMultipleOccurrences(t *testing.T) {
	t.Parallel()
	a := newAnonymizer("frank")
	input := "Reading /Users/frank/a.go and /Users/frank/b.go"
	got := a.anonymizeText(input)
	if strings.Contains(got, "/Users/frank/") {
		t.Errorf("username not fully anonymized: %s", got)
	}
	hashed := hashedUser("frank")
	if count := strings.Count(got, hashed); count != 2 {
		t.Errorf("expected 2 occurrences of hashed user, got %d: %s", count, got)
	}
}

func TestAnonymizeNilAnonymizer(t *testing.T) {
	t.Parallel()
	a := newAnonymizer("")
	if a != nil {
		t.Error("expected nil anonymizer for empty username")
	}
}

func TestAnonymizeNoMatchLeaveUnchanged(t *testing.T) {
	t.Parallel()
	a := newAnonymizer("frank")
	input := "no paths here, just text"
	got := a.anonymizeText(input)
	if got != input {
		t.Errorf("text should be unchanged: got %q", got)
	}
}

func TestHashedUserDeterministic(t *testing.T) {
	t.Parallel()
	h1 := hashedUser("frank")
	h2 := hashedUser("frank")
	if h1 != h2 {
		t.Errorf("hashed user should be deterministic: %s != %s", h1, h2)
	}
}

func TestHashedUserDifferentForDifferentUsers(t *testing.T) {
	t.Parallel()
	h1 := hashedUser("frank")
	h2 := hashedUser("alice")
	if h1 == h2 {
		t.Errorf("different users should have different hashes: %s == %s", h1, h2)
	}
}
