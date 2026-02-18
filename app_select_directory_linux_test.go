//go:build linux

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelectDirectoryLinux_PrefersKdialog(t *testing.T) {
	restore := saveLinuxSelectDirectoryDeps()
	defer restore()

	tmp := t.TempDir()
	var calls []string

	linuxLookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}
	linuxRunCommand = func(program string, _ ...string) (string, error) {
		calls = append(calls, program)
		if program == "kdialog" {
			return tmp + "\n", nil
		}
		return "", nil
	}

	got := selectDirectoryLinux()
	if got != tmp {
		t.Fatalf("expected %q, got %q", tmp, got)
	}
	if len(calls) != 1 || calls[0] != "kdialog" {
		t.Fatalf("expected only kdialog call, got %v", calls)
	}
}

func TestSelectDirectoryLinux_FallsBackWhenToolMissing(t *testing.T) {
	restore := saveLinuxSelectDirectoryDeps()
	defer restore()

	tmp := t.TempDir()
	var calls []string

	linuxLookPath = func(file string) (string, error) {
		if file == "kdialog" {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + file, nil
	}
	linuxRunCommand = func(program string, _ ...string) (string, error) {
		calls = append(calls, program)
		if program == "zenity" {
			return tmp + "\n", nil
		}
		return "", nil
	}

	got := selectDirectoryLinux()
	if got != tmp {
		t.Fatalf("expected %q, got %q", tmp, got)
	}
	if len(calls) != 1 || calls[0] != "zenity" {
		t.Fatalf("expected only zenity call, got %v", calls)
	}
}

func TestSelectDirectoryLinux_StopsAfterCancelOrError(t *testing.T) {
	restore := saveLinuxSelectDirectoryDeps()
	defer restore()

	var calls []string

	linuxLookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}
	linuxRunCommand = func(program string, _ ...string) (string, error) {
		calls = append(calls, program)
		if program == "kdialog" {
			return "", errors.New("exit status 1")
		}
		return "", nil
	}

	got := selectDirectoryLinux()
	if got != "" {
		t.Fatalf("expected empty path, got %q", got)
	}
	if len(calls) != 1 || calls[0] != "kdialog" {
		t.Fatalf("expected only kdialog call, got %v", calls)
	}
}

func TestSelectDirectoryLinux_RequiresExistingDirectory(t *testing.T) {
	restore := saveLinuxSelectDirectoryDeps()
	defer restore()

	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "not-a-directory.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o600); err != nil {
		t.Fatalf("creating temp file: %v", err)
	}

	linuxLookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}
	linuxRunCommand = func(program string, _ ...string) (string, error) {
		if program == "kdialog" {
			return filePath + "\n", nil
		}
		return "", nil
	}

	got := selectDirectoryLinux()
	if got != "" {
		t.Fatalf("expected empty path, got %q", got)
	}
}

func TestSelectDirectoryLinux_ReturnsEmptyWhenNoChooserFound(t *testing.T) {
	restore := saveLinuxSelectDirectoryDeps()
	defer restore()

	linuxLookPath = func(file string) (string, error) {
		return "", errors.New("not found: " + file)
	}
	linuxRunCommand = func(program string, _ ...string) (string, error) {
		t.Fatalf("did not expect %s to be run", program)
		return "", nil
	}

	if got := selectDirectoryLinux(); got != "" {
		t.Fatalf("expected empty path, got %q", got)
	}
}

func TestLinuxDefaultStartDir_FallsBackToDot(t *testing.T) {
	restore := saveLinuxSelectDirectoryDeps()
	defer restore()

	linuxUserHome = func() (string, error) {
		return "", errors.New("home unavailable")
	}
	if got := linuxDefaultStartDir(); got != "." {
		t.Fatalf("expected '.', got %q", got)
	}

	linuxUserHome = func() (string, error) {
		return "   ", nil
	}
	if got := linuxDefaultStartDir(); got != "." {
		t.Fatalf("expected '.', got %q", got)
	}
}

func TestLinuxDefaultStartDir_UsesHomeWhenAvailable(t *testing.T) {
	restore := saveLinuxSelectDirectoryDeps()
	defer restore()

	linuxUserHome = func() (string, error) {
		return "/home/alex", nil
	}
	if got := linuxDefaultStartDir(); got != "/home/alex" {
		t.Fatalf("expected /home/alex, got %q", got)
	}
}

func TestSelectDirectoryLinux_TrimsOutput(t *testing.T) {
	restore := saveLinuxSelectDirectoryDeps()
	defer restore()

	tmp := t.TempDir()
	linuxLookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}
	linuxRunCommand = func(program string, _ ...string) (string, error) {
		if program == "kdialog" {
			return strings.Repeat(" ", 2) + tmp + "\n", nil
		}
		return "", nil
	}

	got := selectDirectoryLinux()
	if got != tmp {
		t.Fatalf("expected %q, got %q", tmp, got)
	}
}

func TestSelectDirectoryLinux_NormalizesFileURLOutput(t *testing.T) {
	restore := saveLinuxSelectDirectoryDeps()
	defer restore()

	tmp := t.TempDir()
	linuxLookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}
	linuxRunCommand = func(program string, _ ...string) (string, error) {
		if program == "kdialog" {
			return "file://" + tmp + "\n", nil
		}
		return "", nil
	}

	got := selectDirectoryLinux()
	if got != tmp {
		t.Fatalf("expected %q, got %q", tmp, got)
	}
}

func saveLinuxSelectDirectoryDeps() func() {
	origLookPath := linuxLookPath
	origRunCommand := linuxRunCommand
	origStat := linuxStat
	origUserHome := linuxUserHome

	return func() {
		linuxLookPath = origLookPath
		linuxRunCommand = origRunCommand
		linuxStat = origStat
		linuxUserHome = origUserHome
	}
}
