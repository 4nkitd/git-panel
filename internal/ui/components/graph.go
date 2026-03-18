package components

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ankityadav/zedgit/internal/git"
	"github.com/ankityadav/zedgit/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

const (
	dotCurrent = "○"
	dotNormal  = "●"
	dotMerge   = "◎"
	lineVert   = "│"
)

// RenderGraphSection renders the collapsible graph header.
func RenderGraphSection(collapsed bool, width int, focused bool) string {
	arrow := "▾"
	if collapsed {
		arrow = "▸"
	}
	arrow = styles.HelpDescStyle.Render(arrow)

	header := fmt.Sprintf(" %s %s", arrow, styles.SectionHeaderStyle.Render("Graph"))

	header = padRight(header, width)
	if focused {
		header = styles.SelectedStyle.Render(header)
	}
	return header
}

// RenderCommitLine renders a single commit with graph dot, subject, refs, and time.
func RenderCommitLine(commit git.CommitInfo, width int, selected bool, hovered bool, isHead bool) string {
	colIdx := commit.BranchIdx % len(styles.GraphColors)
	branchColor := styles.GraphColors[colIdx]
	dotStyle := lipgloss.NewStyle().Foreground(branchColor)

	// Dot character
	dot := dotNormal
	if isHead {
		dot = dotCurrent
	}
	if len(commit.Parents) > 1 {
		dot = dotMerge
	}

	prefix := dotStyle.Render(dot) + " "

	// Ref badges
	refBadges := renderRefs(commit.Refs)

	// Relative time (right-aligned, dim)
	timeStr := ""
	if commit.Date != "" {
		timeStr = styles.GraphDateStyle.Render(commit.Date)
	}
	timeWidth := lipgloss.Width(timeStr)

	// Calculate available space for subject
	availSubject := width - 3 - timeWidth - 2 // dot+space + time + padding
	if refBadges != "" {
		availSubject -= lipgloss.Width(refBadges) + 1
	}

	subject := commit.Subject
	if len(subject) > availSubject && availSubject > 5 {
		subject = subject[:availSubject-3] + "..."
	}

	line := prefix + styles.GraphCommitStyle.Render(subject)
	if refBadges != "" {
		line += " " + refBadges
	}

	// Right-align the time
	lw := lipgloss.Width(line)
	gap := width - lw - timeWidth - 1
	if gap > 0 {
		line += strings.Repeat(" ", gap) + timeStr
	}

	line = padRight(line, width)

	if selected {
		return styles.SelectedStyle.Render(line)
	} else if hovered {
		return styles.HoverStyle.Render(line)
	}
	return line
}

// RenderCommitFiles renders the expanded file list under a commit.
func RenderCommitFiles(files []git.CommitFile, branchIdx int, width int) []string {
	if len(files) == 0 {
		return []string{styles.HelpDescStyle.Render("    No files changed")}
	}

	colIdx := branchIdx % len(styles.GraphColors)
	branchColor := styles.GraphColors[colIdx]
	lineStyle := lipgloss.NewStyle().Foreground(branchColor)

	var lines []string
	for _, f := range files {
		connector := "  " + lineStyle.Render(lineVert) + " "

		name := filepath.Base(f.Path)
		dir := filepath.Dir(f.Path)
		if dir == "." {
			dir = ""
		}

		nameStyled := styles.FileNameStyle.Render(name)
		dirStyled := ""
		if dir != "" {
			dirStyled = " " + styles.FileDirStyle.Render(dir)
		}

		statusStr := styles.StatusStyle(f.Status).Render(f.Status.String())

		content := connector + nameStyled + dirStyled
		cw := lipgloss.Width(content)
		sw := lipgloss.Width(statusStr)
		gap := width - cw - sw - 1
		if gap < 1 {
			gap = 1
		}

		lines = append(lines, content+strings.Repeat(" ", gap)+statusStr)
	}

	return lines
}

func renderRefs(refs []string) string {
	if len(refs) == 0 {
		return ""
	}
	var parts []string
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		switch {
		case strings.HasPrefix(ref, "HEAD -> "):
			branch := strings.TrimPrefix(ref, "HEAD -> ")
			parts = append(parts, styles.GraphBranchRefStyle.Render(branch))
		case strings.HasPrefix(ref, "tag: "):
			tag := strings.TrimPrefix(ref, "tag: ")
			parts = append(parts, styles.GraphTagStyle.Render(tag))
		case strings.Contains(ref, "/"):
			parts = append(parts, styles.GraphRemoteRefStyle.Render(ref))
		default:
			parts = append(parts, styles.GraphBranchRefStyle.Render(ref))
		}
	}
	return strings.Join(parts, " ")
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}
