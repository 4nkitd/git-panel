package git

import (
	"fmt"
	"strings"
)

// CommitInfo represents a single commit in the log.
type CommitInfo struct {
	Hash       string
	ShortHash  string
	Subject    string
	Author     string
	Date       string // relative date like "2 hours ago"
	Refs       []string
	Parents    []string
	BranchIdx  int  // which graph column this commit is on
	Expanded   bool // whether to show changed files
	Files      []CommitFile
}

// CommitFile represents a file changed in a commit.
type CommitFile struct {
	Path   string
	Status FileStatus
}

// LogResult holds the parsed git log output.
type LogResult struct {
	Commits []CommitInfo
}

// Log returns recent commits with graph info.
func (r *Repo) Log(maxCount int) (*LogResult, error) {
	if maxCount <= 0 {
		maxCount = 50
	}

	// Use a custom format to get structured data
	// %H = full hash, %h = short hash, %s = subject, %an = author, %cr = relative date
	// %D = refs, %P = parent hashes
	format := "%H|%h|%s|%an|%cr|%D|%P"
	out, err := runGit(r.Path, "log",
		fmt.Sprintf("--max-count=%d", maxCount),
		"--format="+format,
		"--all",
	)
	if err != nil {
		return nil, err
	}

	result := &LogResult{}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 7)
		if len(parts) < 7 {
			continue
		}

		commit := CommitInfo{
			Hash:      parts[0],
			ShortHash: parts[1],
			Subject:   parts[2],
			Author:    parts[3],
			Date:      parts[4],
		}

		// Parse refs
		if parts[5] != "" {
			for _, ref := range strings.Split(parts[5], ", ") {
				commit.Refs = append(commit.Refs, strings.TrimSpace(ref))
			}
		}

		// Parse parents
		if parts[6] != "" {
			for _, p := range strings.Fields(parts[6]) {
				commit.Parents = append(commit.Parents, p)
			}
		}

		result.Commits = append(result.Commits, commit)
	}

	// Assign branch columns for graph drawing
	assignGraphColumns(result)

	return result, nil
}

// CommitFiles returns the files changed in a specific commit.
func (r *Repo) CommitFiles(hash string) ([]CommitFile, error) {
	out, err := runGit(r.Path, "diff-tree", "--no-commit-id", "-r", "--name-status", hash)
	if err != nil {
		return nil, err
	}

	var files []CommitFile
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		status := StatusModified
		switch parts[0][0] {
		case 'A':
			status = StatusAdded
		case 'D':
			status = StatusDeleted
		case 'M':
			status = StatusModified
		case 'R':
			status = StatusRenamed
		case 'C':
			status = StatusCopied
		case 'T':
			status = StatusTypeChanged
		}

		path := parts[len(parts)-1]
		files = append(files, CommitFile{Path: path, Status: status})
	}

	return files, nil
}

// assignGraphColumns assigns a graph column index to each commit for drawing
// branch lines. This is a simplified version that tracks active branches.
func assignGraphColumns(result *LogResult) {
	if len(result.Commits) == 0 {
		return
	}

	// Track which column each branch tip is in
	// Map from commit hash -> column index
	columns := make(map[string]int)
	nextCol := 0

	for i := range result.Commits {
		c := &result.Commits[i]

		// If this commit already has a column assigned (it's a known parent)
		if col, ok := columns[c.Hash]; ok {
			c.BranchIdx = col
			delete(columns, c.Hash)
		} else {
			// New branch line, assign next column
			c.BranchIdx = nextCol
			nextCol++
		}

		// Assign columns to parents
		for j, parentHash := range c.Parents {
			if _, exists := columns[parentHash]; !exists {
				if j == 0 {
					// First parent continues in same column
					columns[parentHash] = c.BranchIdx
				} else {
					// Merge parent gets a new column
					columns[parentHash] = nextCol
					nextCol++
				}
			}
		}

		// Cap columns to prevent runaway
		if nextCol > 6 {
			nextCol = 6
		}
	}
}
