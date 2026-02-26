//go:build !((darwin && arm64) || (linux && amd64))

package nomic

// modelGZ is empty on unsupported platforms — Supported() returns false.
var modelGZ []byte
