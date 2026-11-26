//go:build !windows

package ffmpeg

import "os/exec"

// ConfigureCmd is a no-op on non-Windows platforms.
func ConfigureCmd(cmd *exec.Cmd) {}
