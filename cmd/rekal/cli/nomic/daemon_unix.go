//go:build !windows

package nomic

import (
	"os/exec"
	"syscall"
)

// setSysProcAttr detaches the daemon process from the parent session.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
