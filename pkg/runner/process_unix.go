//go:build !windows

package runner

import (
	"os/exec"
	"syscall"
	"time"
)

type ProcessTree struct {
	cmd *exec.Cmd
}

func NewProcessTree(cmd *exec.Cmd) *ProcessTree {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return &ProcessTree{cmd: cmd}
}

func (pt *ProcessTree) Start() error {
	return pt.cmd.Start()
}

func (pt *ProcessTree) Wait() error {
	return pt.cmd.Wait()
}

func (pt *ProcessTree) Kill(gracePeriod time.Duration) {
	if pt.cmd.Process == nil {
		return
	}
	pid := pt.cmd.Process.Pid

	// Send SIGTERM to the entire process group
	_ = syscall.Kill(-pid, syscall.SIGTERM)

	// Wait for grace period
	time.Sleep(gracePeriod)

	// Send SIGKILL to the process group to ensure everything is dead
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}
