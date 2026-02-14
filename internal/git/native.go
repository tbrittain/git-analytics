package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// nativeRepo implements Repository by shelling out to the native git CLI.
// This is dramatically faster than go-git for large repositories because
// native git has optimized packfile handling and memory-mapped I/O.
type nativeRepo struct {
	path string
}

// NativeOpen opens an existing git repository using the native git CLI.
func NativeOpen(path string) (Repository, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	hideWindow(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("not a git repository (or git not installed): %s", strings.TrimSpace(string(out)))
	}
	return &nativeRepo{path: path}, nil
}

func (r *nativeRepo) RepoName() string {
	return filepath.Base(r.path)
}

func (r *nativeRepo) CurrentBranch() string {
	cmd := exec.Command("git", "-C", r.path, "symbolic-ref", "--short", "HEAD")
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "HEAD"
	}
	return strings.TrimSpace(string(out))
}

func (r *nativeRepo) HeadHash() (string, error) {
	cmd := exec.Command("git", "-C", r.path, "rev-parse", "HEAD")
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *nativeRepo) Log(sinceHash string) (CommitIter, error) {
	args := []string{
		"-C", r.path, "log",
		"--format=GITANALYTICS_COMMIT%n%H%n%aN%n%aE%n%aI%n%s",
		"--numstat",
	}
	if sinceHash != "" {
		args = append(args, sinceHash+"..HEAD")
	}

	cmd := exec.Command("git", args...)
	hideWindow(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting git log: %w", err)
	}

	return &nativeCommitIter{
		scanner: bufio.NewScanner(stdout),
		cmd:     cmd,
	}, nil
}

func (r *nativeRepo) Close() error {
	return nil
}

// nativeCommitIter parses streaming output from git log --numstat.
type nativeCommitIter struct {
	scanner   *bufio.Scanner
	cmd       *exec.Cmd
	peeked    bool   // true if we've already scanned a line that needs re-reading
	peekLine  string // the line we peeked at
	exhausted bool
}

func (it *nativeCommitIter) nextLine() (string, bool) {
	if it.peeked {
		it.peeked = false
		return it.peekLine, true
	}
	if it.scanner.Scan() {
		return it.scanner.Text(), true
	}
	return "", false
}

func (it *nativeCommitIter) unread(line string) {
	it.peeked = true
	it.peekLine = line
}

func (it *nativeCommitIter) Next() (*Commit, error) {
	if it.exhausted {
		return nil, nil
	}

	// Scan until we find the sentinel line.
	for {
		line, ok := it.nextLine()
		if !ok {
			it.exhausted = true
			return nil, nil
		}
		if line == "GITANALYTICS_COMMIT" {
			break
		}
	}

	// Read 5 metadata lines: hash, name, email, date, subject.
	meta := make([]string, 5)
	for i := range meta {
		line, ok := it.nextLine()
		if !ok {
			return nil, fmt.Errorf("unexpected end of git log output (expected metadata line %d)", i)
		}
		meta[i] = line
	}

	date, err := time.Parse(time.RFC3339, meta[3])
	if err != nil {
		return nil, fmt.Errorf("parsing date %q: %w", meta[3], err)
	}

	// Read numstat lines until next sentinel or EOF.
	var files []FileStat
	for {
		line, ok := it.nextLine()
		if !ok {
			it.exhausted = true
			break
		}
		if line == "GITANALYTICS_COMMIT" {
			it.unread(line)
			break
		}
		if line == "" {
			continue
		}

		fs, err := parseNumstatLine(line)
		if err != nil {
			continue // skip unparseable lines
		}
		files = append(files, fs)
	}

	return &Commit{
		Hash:         meta[0],
		AuthorName:   meta[1],
		AuthorEmail:  meta[2],
		Date:         date,
		Message:      meta[4],
		FilesChanged: files,
	}, nil
}

func (it *nativeCommitIter) Close() {
	if it.cmd != nil && it.cmd.Process != nil {
		it.cmd.Process.Kill()
		it.cmd.Wait()
	}
}

// parseNumstatLine parses a single --numstat output line.
// Format: "additions\tdeletions\tpath"
// Binary files show "-\t-\tpath" — treated as 0/0.
func parseNumstatLine(line string) (FileStat, error) {
	parts := strings.SplitN(line, "\t", 3)
	if len(parts) != 3 {
		return FileStat{}, fmt.Errorf("expected 3 tab-separated fields, got %d", len(parts))
	}

	var additions, deletions int
	if parts[0] != "-" {
		var err error
		additions, err = strconv.Atoi(parts[0])
		if err != nil {
			return FileStat{}, fmt.Errorf("parsing additions %q: %w", parts[0], err)
		}
	}
	if parts[1] != "-" {
		var err error
		deletions, err = strconv.Atoi(parts[1])
		if err != nil {
			return FileStat{}, fmt.Errorf("parsing deletions %q: %w", parts[1], err)
		}
	}

	return FileStat{
		Path:      parts[2],
		Additions: additions,
		Deletions: deletions,
	}, nil
}
