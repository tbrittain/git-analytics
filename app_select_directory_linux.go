//go:build linux

package main

import (
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type linuxDirectoryChooser struct {
	program string
	args    []string
}

const selectDirectoryDialogTitle = "Select Git Repository"

var (
	linuxLookPath   = exec.LookPath
	linuxRunCommand = runLinuxDirectoryChooser
	linuxStat       = os.Stat
	linuxUserHome   = os.UserHomeDir
)

// SelectDirectory opens a Linux-native folder picker and returns the selected path.
// Returns an empty string if the user cancels, no supported chooser is installed,
// or an error occurs.
func (a *App) SelectDirectory() string {
	return selectDirectoryLinux()
}

func selectDirectoryLinux() string {
	startDir := linuxDefaultStartDir()
	choosers := []linuxDirectoryChooser{
		{
			program: "kdialog",
			args:    []string{"--getexistingdirectory", startDir, "--title", selectDirectoryDialogTitle},
		},
		{
			program: "zenity",
			args:    []string{"--file-selection", "--directory", "--title=" + selectDirectoryDialogTitle},
		},
		{
			program: "yad",
			args:    []string{"--file-selection", "--directory", "--title=" + selectDirectoryDialogTitle},
		},
	}

	for _, chooser := range choosers {
		if _, err := linuxLookPath(chooser.program); err != nil {
			continue
		}

		return linuxRunChooser(chooser)
	}

	return ""
}

func linuxDefaultStartDir() string {
	home, err := linuxUserHome()
	if err != nil || strings.TrimSpace(home) == "" {
		return "."
	}
	return home
}

func linuxIsExistingDir(path string) bool {
	info, err := linuxStat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func runLinuxDirectoryChooser(program string, args ...string) (string, error) {
	cmd := exec.Command(program, args...)
	output, err := cmd.Output()
	return string(output), err
}

func linuxRunChooser(chooser linuxDirectoryChooser) string {
	out, err := linuxRunCommand(chooser.program, chooser.args...)
	if err != nil {
		// Includes user cancel (non-zero exit). Do not cascade to another dialog.
		return ""
	}

	path := linuxNormalizeSelectedPath(out)
	if path == "" || !linuxIsExistingDir(path) {
		return ""
	}

	return path
}

func linuxNormalizeSelectedPath(out string) string {
	path := strings.TrimSpace(out)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "file://") {
		return path
	}

	parsed, err := url.Parse(path)
	if err != nil || parsed.Scheme != "file" {
		return ""
	}

	unescaped, err := url.PathUnescape(parsed.Path)
	if err != nil {
		return ""
	}
	return unescaped
}
