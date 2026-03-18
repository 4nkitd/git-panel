package git

import (
	"strings"
)

// parseStatusEntry parses a porcelain v2 status line for staged changes.
// Format: 1 XY sub mH mI mW hH hI path
// or:     2 XY sub mH mI mW hH hI X\tscore path\torigPath
func parseStatusEntry(line string) *StatusEntry {
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return nil
	}

	xy := fields[1]
	if len(xy) < 2 {
		return nil
	}

	indexStatus := xy[0] // staged status
	if indexStatus == '.' {
		return nil // no staged change
	}

	// For renames (type "2"), the path is after the last tab
	var path string
	if fields[0] == "2" {
		// Rename entry: fields joined, split by tab
		rest := strings.SplitN(line, "\t", 3)
		if len(rest) >= 2 {
			path = rest[1]
		}
	} else {
		path = fields[len(fields)-1]
	}

	entry := &StatusEntry{
		Path:     path,
		Status:   charToStatus(indexStatus),
		IsStaged: true,
	}
	return entry
}

// parseUnstagedEntry parses a porcelain v2 status line for unstaged (worktree) changes.
func parseUnstagedEntry(line string) *StatusEntry {
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return nil
	}

	xy := fields[1]
	if len(xy) < 2 {
		return nil
	}

	wtStatus := xy[1] // worktree status
	if wtStatus == '.' {
		return nil // no worktree change
	}

	var path string
	if fields[0] == "2" {
		rest := strings.SplitN(line, "\t", 3)
		if len(rest) >= 2 {
			path = rest[1]
		}
	} else {
		path = fields[len(fields)-1]
	}

	return &StatusEntry{
		Path:   path,
		Status: charToStatus(wtStatus),
	}
}

// parseConflictPath parses a porcelain v2 unmerged entry.
// Format: u XY sub m1 m2 m3 mW h1 h2 h3 path
func parseConflictPath(line string) string {
	fields := strings.Fields(line)
	if len(fields) >= 11 {
		return fields[10]
	}
	return ""
}

func charToStatus(c byte) FileStatus {
	switch c {
	case 'M':
		return StatusModified
	case 'A':
		return StatusAdded
	case 'D':
		return StatusDeleted
	case 'R':
		return StatusRenamed
	case 'C':
		return StatusCopied
	case 'T':
		return StatusTypeChanged
	default:
		return StatusModified
	}
}

// parseDiff parses unified diff output into structured DiffResult.
func parseDiff(filePath, raw string) *DiffResult {
	result := &DiffResult{FilePath: filePath}
	if raw == "" {
		return result
	}

	oldNum := 0
	newNum := 0

	for _, line := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(line, "@@"):
			// Hunk header: @@ -oldStart,oldCount +newStart,newCount @@
			result.Lines = append(result.Lines, DiffLine{
				Type:    DiffHunkHeader,
				Content: line,
			})
			// Parse line numbers
			parseHunkHeader(line, &oldNum, &newNum)

		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			result.Lines = append(result.Lines, DiffLine{
				Type:    DiffAdd,
				Content: line[1:],
				NewNum:  newNum,
			})
			newNum++

		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			result.Lines = append(result.Lines, DiffLine{
				Type:    DiffDelete,
				Content: line[1:],
				OldNum:  oldNum,
			})
			oldNum++

		case strings.HasPrefix(line, " "):
			result.Lines = append(result.Lines, DiffLine{
				Type:    DiffContext,
				Content: line[1:],
				OldNum:  oldNum,
				NewNum:  newNum,
			})
			oldNum++
			newNum++

		// Skip diff header lines (diff --git, index, ---, +++)
		}
	}

	return result
}

func parseHunkHeader(line string, oldNum, newNum *int) {
	// @@ -10,5 +20,8 @@
	parts := strings.Fields(line)
	for _, p := range parts {
		if strings.HasPrefix(p, "-") && strings.Contains(p, ",") {
			n := 0
			for i := 1; i < len(p); i++ {
				if p[i] == ',' {
					break
				}
				n = n*10 + int(p[i]-'0')
			}
			*oldNum = n
		}
		if strings.HasPrefix(p, "+") && strings.Contains(p, ",") {
			n := 0
			for i := 1; i < len(p); i++ {
				if p[i] == ',' {
					break
				}
				n = n*10 + int(p[i]-'0')
			}
			*newNum = n
		}
	}
}
