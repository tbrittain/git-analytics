package git

import (
	"os/exec"
	"syscall"
)

// hideWindow sets process creation flags to prevent a visible console window
// from flashing on screen when running git subprocesses.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// https://learn.microsoft.com/en-us/windows/win32/procthread/process-creation-flags
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
