package nomic

import "errors"

// ErrNotSupported is returned when nomic embeddings are not available
// on the current platform.
var ErrNotSupported = errors.New("nomic embeddings not supported on this platform")
