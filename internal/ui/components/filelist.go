package components

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/4nkitd/git-panel/internal/git"
	"github.com/4nkitd/git-panel/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// RenderFileList renders a list of status entries as a file list section.
func RenderFileList(entries []git.StatusEntry, cursor int, focused bool, width int) string {
	if len(entries) == 0 {
		return styles.HelpDescStyle.Render("  No changes")
	}

	var lines []string
	for i, entry := range entries {
		line := renderFileLine(entry, width, i == cursor && focused)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func renderFileLine(entry git.StatusEntry, width int, selected bool) string {
	statusStr := entry.Status.String()
	statusStyled := styles.StatusStyle(entry.Status).Render(statusStr)

	// Show just filename for narrow widths, full path for wider
	display := entry.Path
	if width < 50 {
		display = filepath.Base(entry.Path)
	} else if width < 70 {
		// Truncate long paths
		maxLen := width - 10
		if len(display) > maxLen {
			display = "..." + display[len(display)-maxLen+3:]
		}
	}

	line := fmt.Sprintf("  %s  %s", statusStyled, styles.FilePathStyle.Render(display))

	if selected {
		// Pad to full width and apply selected background
		padding := width - lipgloss.Width(line)
		if padding > 0 {
			line += strings.Repeat(" ", padding)
		}
		line = styles.SelectedStyle.Render(line)
	}

	return line
}

// RenderSectionHeader renders a collapsible section header.
func RenderSectionHeader(title string, count int, collapsed bool, width int, actionHint string) string {
	arrow := "▼"
	if collapsed {
		arrow = "▶"
	}

	countStr := styles.SectionCountStyle.Render(fmt.Sprintf("(%d)", count))
	header := fmt.Sprintf(" %s %s %s", arrow, styles.SectionHeaderStyle.Render(title), countStr)

	if actionHint != "" {
		hintWidth := lipgloss.Width(actionHint)
		headerWidth := lipgloss.Width(header)
		gap := width - headerWidth - hintWidth - 1
		if gap > 0 {
			header += strings.Repeat(" ", gap) + styles.HelpDescStyle.Render(actionHint)
		}
	}

	return header
}
