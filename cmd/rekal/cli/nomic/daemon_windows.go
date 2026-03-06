//go:build windows

package nomic

import "os/exec"

// setSysProcAttr is a no-op on Windows (Unix sockets not supported).
func setSysProcAttr(_ *exec.Cmd) {}
