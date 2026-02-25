package cli

import "errors"

// SilentError wraps an error that has already been printed to the user.
// main.go checks for this type and skips printing when found.
type SilentError struct {
	err error
}

func (e *SilentError) Error() string {
	return e.err.Error()
}

func (e *SilentError) Unwrap() error {
	return e.err
}

// NewSilentError wraps err so main.go knows not to print it again.
func NewSilentError(err error) error {
	return &SilentError{err: err}
}

// IsSilentError reports whether err (or any error in its chain) is a SilentError.
func IsSilentError(err error) bool {
	var se *SilentError
	return errors.As(err, &se)
}
