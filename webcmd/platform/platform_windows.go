// +build windows

package platform

import (
    "os/exec"
    "syscall"
)

func Hide(c *exec.Cmd) {
    if c.SysProcAttr == nil {
        c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
    } else {
        c.SysProcAttr.HideWindow = true
    }
}