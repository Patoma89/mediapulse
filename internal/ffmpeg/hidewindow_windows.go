//go:build windows

package ffmpeg

import (
	"os/exec"
	"syscall"
)

// ConfigureCmd hides console windows for spawned helper processes on Windows.
func ConfigureCmd(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
