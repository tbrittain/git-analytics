//go:build !windows

package git

import "os/exec"

// hideWindow is a no-op on non-Windows platforms.
func hideWindow(_ *exec.Cmd) {}
