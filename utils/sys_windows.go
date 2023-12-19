//go:build windows

package utils

import (
	"context"
	"os/exec"
	"strings"
	"syscall"
)

func RunCmd(ctx context.Context, cmd ...string) (int, []byte, error) {
	c := exec.CommandContext(ctx, "cmd", "/c", strings.Join(cmd, " "))

	sysProcAttr(c)

	output, err := c.CombinedOutput()
	if err != nil {
		return -1, nil, err
	}
	return c.Process.Pid, output, nil
}

func sysProcAttr(r *exec.Cmd) {
	r.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		//CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
