package components

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ankityadav/zedgit/internal/git"
	"github.com/ankityadav/zedgit/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// graphDot characters for the commit graph.
const (
	dotCurrent = "○" // HEAD / current branch commit
	dotNormal  = "●" // regular commit
	dotMerge   = "◎" // merge commit
	lineVert   = "│"
	lineConn   = "│"
)

// RenderGraphSection renders the commit graph section header.
func RenderGraphSection(collapsed bool, width int, focused bool) string {
	arrow := "▼"
	if collapsed {
		arrow = "▶"
	}

	header := fmt.Sprintf(" %s %s", arrow, styles.SectionHeaderStyle.Render("Graph"))

	if focused {
		header = padRight(header, width)
		header = styles.SelectedStyle.Render(header)
	}

	return header
}

// RenderCommitLine renders a single commit in the graph.
func RenderCommitLine(commit git.CommitInfo, width int, selected bool, hovered bool, isHead bool) string {
	colIdx := commit.BranchIdx % len(styles.GraphColors)
	branchColor := styles.GraphColors[colIdx]
	dotStyle := lipgloss.NewStyle().Foreground(branchColor)

	// Choose dot character
	dot := dotNormal
	if isHead {
		dot = dotCurrent
	}
	if len(commit.Parents) > 1 {
		dot = dotMerge
	}

	// Build graph prefix (the branch lines + dot)
	prefix := dotStyle.Render(dot) + " "

	// Commit subject (truncated to fit)
	maxSubject := width - 4 // dot + space + padding
	subject := commit.Subject

	// Add ref badges
	refBadges := renderRefs(commit.Refs)
	if refBadges != "" {
		maxSubject -= lipgloss.Width(refBadges) + 1
	}

	if len(subject) > maxSubject {
		if maxSubject > 3 {
			subject = subject[:maxSubject-3] + "..."
		}
	}

	line := prefix + styles.GraphCommitStyle.Render(subject)
	if refBadges != "" {
		line += " " + refBadges
	}

	// Expand icon on hover
	if hovered || selected {
		expandIcon := " " + styles.HelpDescStyle.Render("⤡")
		lineW := lipgloss.Width(line) + lipgloss.Width(expandIcon)
		if lineW < width {
			gap := width - lineW
			line += strings.Repeat(" ", gap) + expandIcon
		}
	}

	line = padRight(line, width)

	if selected {
		line = styles.SelectedStyle.Render(line)
	} else if hovered {
		line = styles.HoverStyle.Render(line)
	}

	return line
}

// RenderCommitFiles renders the expanded file list for a commit.
func RenderCommitFiles(files []git.CommitFile, branchIdx int, width int) []string {
	if len(files) == 0 {
		return []string{styles.HelpDescStyle.Render("    No files changed")}
	}

	colIdx := branchIdx % len(styles.GraphColors)
	branchColor := styles.GraphColors[colIdx]
	lineStyle := lipgloss.NewStyle().Foreground(branchColor)

	var lines []string
	for _, f := range files {
		// Draw connecting line from graph
		graphLine := "  " + lineStyle.Render(lineVert) + "  "

		// File icon based on extension (simplified)
		icon := fileIcon(f.Path)

		// File name and directory
		name := filepath.Base(f.Path)
		dir := filepath.Dir(f.Path)
		if dir == "." {
			dir = ""
		}

		// Status indicator on right
		statusStr := styles.StatusStyle(f.Status).Render(f.Status.String())

		// Build the line
		nameStyled := styles.FilePathStyle.Render(name)
		dirStyled := ""
		if dir != "" {
			dirStyled = " " + styles.FileDirStyle.Render(dir)
		}

		content := graphLine + icon + " " + nameStyled + dirStyled
		contentWidth := lipgloss.Width(content)
		statusWidth := lipgloss.Width(statusStr)

		gap := width - contentWidth - statusWidth - 1
		if gap < 1 {
			gap = 1
		}
		line := content + strings.Repeat(" ", gap) + statusStr

		lines = append(lines, line)
	}

	return lines
}

// RenderGraphLines draws vertical branch lines for rows between commits.
func RenderGraphLines(branchIdx int) string {
	colIdx := branchIdx % len(styles.GraphColors)
	branchColor := styles.GraphColors[colIdx]
	return lipgloss.NewStyle().Foreground(branchColor).Render(lineConn)
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
			parts = append(parts, styles.GraphTagStyle.Render("🏷 "+tag))
		case strings.Contains(ref, "/"):
			// Remote ref like origin/main
			parts = append(parts, styles.GraphRemoteRefStyle.Render(ref))
		default:
			parts = append(parts, styles.GraphBranchRefStyle.Render(ref))
		}
	}

	return strings.Join(parts, " ")
}

func fileIcon(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return styles.HelpDescStyle.Render("Go")
	case ".js", ".jsx":
		return styles.GraphTagStyle.Render("JS")
	case ".ts", ".tsx":
		return lipgloss.NewStyle().Foreground(styles.Blue).Render("TS")
	case ".py":
		return lipgloss.NewStyle().Foreground(styles.Yellow).Render("Py")
	case ".rs":
		return lipgloss.NewStyle().Foreground(styles.Orange).Render("Rs")
	case ".md":
		return lipgloss.NewStyle().Foreground(styles.Cyan).Render("Md")
	case ".json", ".toml", ".yaml", ".yml":
		return lipgloss.NewStyle().Foreground(styles.DimWhite).Render("{}")
	case ".css", ".scss":
		return lipgloss.NewStyle().Foreground(styles.Magenta).Render("Cs")
	default:
		return styles.HelpDescStyle.Render("··")
	}
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}
