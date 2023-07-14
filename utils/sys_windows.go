//go:build windows

package utils

import (
	"os/exec"
	"syscall"
)

func sysProcAttr(r *exec.Cmd) {
	r.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
