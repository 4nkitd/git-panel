package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileStatus represents the status of a file in the git working tree.
type FileStatus int

const (
	StatusModified FileStatus = iota
	StatusAdded
	StatusDeleted
	StatusRenamed
	StatusUntracked
	StatusConflicted
	StatusCopied
	StatusTypeChanged
)

func (s FileStatus) String() string {
	switch s {
	case StatusModified:
		return "M"
	case StatusAdded:
		return "A"
	case StatusDeleted:
		return "D"
	case StatusRenamed:
		return "R"
	case StatusUntracked:
		return "?"
	case StatusConflicted:
		return "U"
	case StatusCopied:
		return "C"
	case StatusTypeChanged:
		return "T"
	default:
		return " "
	}
}

// StatusEntry represents a single file's status.
type StatusEntry struct {
	Path     string
	Status   FileStatus
	OldPath  string // for renames
	IsStaged bool
}

// BranchInfo holds information about the current branch.
type BranchInfo struct {
	Name       string
	IsDetached bool
	Upstream   string
	Ahead      int
	Behind     int
}

// StashEntry represents a single stash.
type StashEntry struct {
	Index   int
	Message string
}

// RepoStatus holds the full status of a repository.
type RepoStatus struct {
	Branch     BranchInfo
	Staged     []StatusEntry
	Unstaged   []StatusEntry
	Stashes    []StashEntry
	MergeHead  bool
	RebaseHead bool
}

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Type    DiffLineType
	Content string
	OldNum  int
	NewNum  int
}

type DiffLineType int

const (
	DiffContext DiffLineType = iota
	DiffAdd
	DiffDelete
	DiffHunkHeader
)

// DiffResult holds the diff output for a file.
type DiffResult struct {
	FilePath string
	Lines    []DiffLine
}

// Repo wraps git operations for a repository.
type Repo struct {
	Path string
}

// Open finds and opens a git repository from the given path.
func Open(path string) (*Repo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Find the git root
	out, err := runGit(absPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	return &Repo{Path: strings.TrimSpace(out)}, nil
}

// Status returns the current repository status.
func (r *Repo) Status() (*RepoStatus, error) {
	status := &RepoStatus{}

	// Get branch info
	branch, err := r.getBranchInfo()
	if err == nil {
		status.Branch = branch
	}

	// Get file statuses
	out, err := runGit(r.Path, "status", "--porcelain=v2", "--branch")
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "# branch.ab"):
			// Parse ahead/behind: # branch.ab +3 -1
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				fmt.Sscanf(parts[2], "+%d", &status.Branch.Ahead)
				fmt.Sscanf(parts[3], "-%d", &status.Branch.Behind)
			}

		case strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 "):
			// Changed entry — check staged and unstaged independently
			stagedEntry := parseStatusEntry(line)
			if stagedEntry != nil {
				status.Staged = append(status.Staged, *stagedEntry)
			}
			unstagedEntry := parseUnstagedEntry(line)
			if unstagedEntry != nil {
				status.Unstaged = append(status.Unstaged, *unstagedEntry)
			}

		case strings.HasPrefix(line, "u "):
			// Unmerged entry (conflict)
			path := parseConflictPath(line)
			if path != "" {
				status.Unstaged = append(status.Unstaged, StatusEntry{
					Path:   path,
					Status: StatusConflicted,
				})
			}

		case strings.HasPrefix(line, "? "):
			// Untracked
			path := strings.TrimSpace(strings.TrimPrefix(line, "? "))
			if path == "" {
				continue
			}
			status.Unstaged = append(status.Unstaged, StatusEntry{
				Path:   path,
				Status: StatusUntracked,
			})
		}
	}

	// Get stash list
	stashes, _ := r.ListStashes()
	status.Stashes = stashes

	// Check merge/rebase state
	status.MergeHead = r.fileExists(".git/MERGE_HEAD")
	status.RebaseHead = r.fileExists(".git/rebase-merge") || r.fileExists(".git/rebase-apply")

	return status, nil
}

func (r *Repo) getBranchInfo() (BranchInfo, error) {
	info := BranchInfo{}

	out, err := runGit(r.Path, "branch", "--show-current")
	if err != nil {
		return info, err
	}
	info.Name = strings.TrimSpace(out)

	if info.Name == "" {
		info.IsDetached = true
		out, err = runGit(r.Path, "rev-parse", "--short", "HEAD")
		if err == nil {
			info.Name = strings.TrimSpace(out)
		}
	}

	// Get upstream
	out, err = runGit(r.Path, "rev-parse", "--abbrev-ref", "@{upstream}")
	if err == nil {
		info.Upstream = strings.TrimSpace(out)
	}

	return info, nil
}

// Stage adds files to the staging area.
func (r *Repo) Stage(paths ...string) error {
	args := append([]string{"add", "--"}, paths...)
	_, err := runGit(r.Path, args...)
	return err
}

// StageAll stages all changes.
func (r *Repo) StageAll() error {
	_, err := runGit(r.Path, "add", "-A")
	return err
}

// Unstage removes files from the staging area.
// Uses "git reset HEAD" which works for both tracked and newly added files,
// unlike "git restore --staged" which can fail on new files in some git versions.
func (r *Repo) Unstage(paths ...string) error {
	args := append([]string{"reset", "HEAD", "--"}, paths...)
	_, err := runGit(r.Path, args...)
	return err
}

// UnstageAll unstages all files.
func (r *Repo) UnstageAll() error {
	_, err := runGit(r.Path, "reset", "HEAD")
	return err
}

// Discard discards local changes to a file (checkout from HEAD).
// For untracked files, this removes them from the working directory.
func (r *Repo) Discard(path string, isUntracked bool) error {
	if isUntracked {
		_, err := runGit(r.Path, "clean", "-f", "--", path)
		return err
	}
	_, err := runGit(r.Path, "checkout", "HEAD", "--", path)
	return err
}

// DiscardAll discards all local changes.
func (r *Repo) DiscardAll() error {
	_, err := runGit(r.Path, "checkout", "--", ".")
	return err
}

// Commit creates a new commit with the given message.
func (r *Repo) Commit(message string) error {
	_, err := runGit(r.Path, "commit", "-m", message)
	return err
}

// CommitAmend amends the last commit with the given message.
func (r *Repo) CommitAmend(message string) error {
	_, err := runGit(r.Path, "commit", "--amend", "-m", message)
	return err
}

// UndoLastCommit soft-resets the last commit.
func (r *Repo) UndoLastCommit() error {
	_, err := runGit(r.Path, "reset", "--soft", "HEAD~1")
	return err
}

// Diff returns the diff for a specific file.
func (r *Repo) Diff(path string, staged bool) (*DiffResult, error) {
	args := []string{"diff", "--no-color"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--", path)

	out, err := runGit(r.Path, args...)
	if err != nil {
		return nil, err
	}

	return parseDiff(path, out), nil
}

// Branches returns a list of local branch names.
func (r *Repo) Branches() ([]string, error) {
	out, err := runGit(r.Path, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	var branches []string
	for _, b := range strings.Split(strings.TrimSpace(out), "\n") {
		if b != "" {
			branches = append(branches, b)
		}
	}
	return branches, nil
}

// RemoteBranches returns a list of remote branch names.
func (r *Repo) RemoteBranches() ([]string, error) {
	out, err := runGit(r.Path, "branch", "-r", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	var branches []string
	for _, b := range strings.Split(strings.TrimSpace(out), "\n") {
		if b != "" {
			branches = append(branches, b)
		}
	}
	return branches, nil
}

// AllBranches returns both local and remote branches.
func (r *Repo) AllBranches() (local []string, remote []string, err error) {
	local, err = r.Branches()
	if err != nil {
		return nil, nil, err
	}
	remote, err = r.RemoteBranches()
	if err != nil {
		return nil, nil, err
	}
	return local, remote, nil
}

// CheckoutBranch switches to the given branch.
func (r *Repo) CheckoutBranch(name string) error {
	_, err := runGit(r.Path, "checkout", name)
	return err
}

// CreateBranch creates and checks out a new branch.
func (r *Repo) CreateBranch(name string) error {
	_, err := runGit(r.Path, "checkout", "-b", name)
	return err
}

// DeleteBranch deletes a local branch.
func (r *Repo) DeleteBranch(name string) error {
	_, err := runGit(r.Path, "branch", "-d", name)
	return err
}

// Push pushes to the remote.
func (r *Repo) Push() error {
	_, err := runGit(r.Path, "push")
	return err
}

// Pull pulls from the remote.
func (r *Repo) Pull() error {
	_, err := runGit(r.Path, "pull")
	return err
}

// Fetch fetches from the remote.
func (r *Repo) Fetch() error {
	_, err := runGit(r.Path, "fetch")
	return err
}

// Stash creates a new stash.
func (r *Repo) Stash(message string) error {
	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	_, err := runGit(r.Path, args...)
	return err
}

// StashPop pops the latest stash.
func (r *Repo) StashPop() error {
	_, err := runGit(r.Path, "stash", "pop")
	return err
}

// StashApply applies a stash by index.
func (r *Repo) StashApply(index int) error {
	_, err := runGit(r.Path, "stash", "apply", fmt.Sprintf("stash@{%d}", index))
	return err
}

// StashDrop drops a stash by index.
func (r *Repo) StashDrop(index int) error {
	_, err := runGit(r.Path, "stash", "drop", fmt.Sprintf("stash@{%d}", index))
	return err
}

// ListStashes returns the list of stashes.
func (r *Repo) ListStashes() ([]StashEntry, error) {
	out, err := runGit(r.Path, "stash", "list", "--format=%gd|%s")
	if err != nil {
		return nil, err
	}

	var stashes []StashEntry
	for i, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		msg := ""
		if len(parts) > 1 {
			msg = parts[1]
		}
		stashes = append(stashes, StashEntry{Index: i, Message: msg})
	}
	return stashes, nil
}

// GetLastCommitMessage returns the message of the last commit.
func (r *Repo) GetLastCommitMessage() (string, error) {
	out, err := runGit(r.Path, "log", "-1", "--format=%s")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (r *Repo) fileExists(relPath string) bool {
	path := filepath.Join(r.Path, relPath)
	_, err := os.Stat(path)
	return err == nil
}

// runGit executes a git command in the given directory.
// Stdout and stderr are captured separately so that stderr warnings
// don't corrupt stdout-based parsing (e.g. status --porcelain).
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		errOut := strings.TrimSpace(stderr.String())
		if errOut == "" {
			errOut = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("%s: %s", err, errOut)
	}
	return stdout.String(), nil
}
