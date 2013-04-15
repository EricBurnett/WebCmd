// Platform-specific utility methods, for the windows platform.
// +build windows

package platform

import (
	"os/exec"
	"syscall"
)

// Configures the given Command to produce a hidden window, if possible. Must
// be called before the Command is executed.
func Hide(c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	} else {
		c.SysProcAttr.HideWindow = true
	}
}
