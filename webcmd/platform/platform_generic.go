// Platform-specific utility methods, for unknown platforms.
// +build !windows

package platform

import (
	"os/exec"
)

// Configures the given Command to produce a hidden window, if possible. Must
// be called before the Command is executed.
func Hide(*exec.Cmd) {
	// No generic version.
}
