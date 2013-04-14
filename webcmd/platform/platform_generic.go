// +build !windows

package platform

import (
    "os/exec"
)

func Hide(*exec.Cmd) {
    // No generic version.
}