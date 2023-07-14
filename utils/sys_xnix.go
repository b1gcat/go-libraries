//go:build linux || darwin

package utils

import (
	"os/exec"
	"syscall"
)

func sysProcAttr(r *exec.Cmd) {
	r.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
