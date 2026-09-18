//go:build windows

package runner

import (
	"os/exec"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type ProcessTree struct {
	cmd *exec.Cmd
	job windows.Handle
}

func NewProcessTree(cmd *exec.Cmd) *ProcessTree {
	return &ProcessTree{cmd: cmd}
}

func (pt *ProcessTree) Start() error {
	err := pt.cmd.Start()
	if err != nil {
		return err
	}

	job, err := windows.CreateJobObject(nil, nil)
	if err == nil {
		info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
			BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
				LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
			},
		}
		_, err = windows.SetInformationJobObject(
			job,
			windows.JobObjectExtendedLimitInformation,
			uintptr(unsafe.Pointer(&info)),
			uint32(unsafe.Sizeof(info)),
		)
		if err == nil {
			h, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(pt.cmd.Process.Pid))
			if err == nil {
				windows.AssignProcessToJobObject(job, h)
				windows.CloseHandle(h)
				pt.job = job
			} else {
				windows.CloseHandle(job)
			}
		} else {
			windows.CloseHandle(job)
		}
	}
	return nil
}

func (pt *ProcessTree) Wait() error {
	err := pt.cmd.Wait()
	if pt.job != 0 {
		windows.CloseHandle(pt.job)
		pt.job = 0
	}
	return err
}

func (pt *ProcessTree) Kill(gracePeriod time.Duration) {
	// Windows doesn't easily support graceful SIGTERM for console apps.
	// We directly terminate the entire process tree.
	if pt.job != 0 {
		windows.TerminateJobObject(pt.job, 1)
		windows.CloseHandle(pt.job)
		pt.job = 0
	} else if pt.cmd.Process != nil {
		pt.cmd.Process.Kill() // Fallback
	}
}
