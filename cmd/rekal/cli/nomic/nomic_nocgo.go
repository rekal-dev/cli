//go:build !((darwin && arm64) || (linux && amd64))

package nomic

const (
	// ModelName identifies the embedding model for the session_embeddings table.
	ModelName = "nomic-v1.5"
	// EmbedDim is the output dimensionality of nomic-embed-text-v1.5.
	EmbedDim = 768
)

// Supported reports whether nomic embeddings are available on this platform.
func Supported() bool {
	return false
}

// Embedder is a stub on unsupported platforms.
type Embedder struct{}

// NewEmbedder always returns ErrNotSupported on unsupported platforms.
func NewEmbedder() (*Embedder, error) {
	return nil, ErrNotSupported
}

// Close is a no-op on unsupported platforms.
func (e *Embedder) Close() {}

// EmbedDocument returns ErrNotSupported.
func (e *Embedder) EmbedDocument(_ string) ([]float64, error) {
	return nil, ErrNotSupported
}

// EmbedQuery returns ErrNotSupported.
func (e *Embedder) EmbedQuery(_ string) ([]float64, error) {
	return nil, ErrNotSupported
}

// EmbedSessions returns ErrNotSupported.
func (e *Embedder) EmbedSessions(_ map[string]string) (map[string][]float64, error) {
	return nil, ErrNotSupported
}
