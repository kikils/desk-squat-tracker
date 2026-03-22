package python

import (
	"os/exec"
	"syscall"
)

func faceChildPrepare(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

func faceChildKill(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	if pid := cmd.Process.Pid; pid > 0 {
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	} else {
		_ = cmd.Process.Kill()
	}
}
