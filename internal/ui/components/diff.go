package components

import (
	"fmt"
	"strings"

	"github.com/ankityadav/zedgit/internal/git"
	"github.com/ankityadav/zedgit/internal/ui/styles"
)

// RenderDiff renders a diff result for display.
func RenderDiff(diff *git.DiffResult, width int, scrollOffset int, maxLines int) string {
	if diff == nil || len(diff.Lines) == 0 {
		return styles.HelpDescStyle.Render("  No diff available")
	}

	header := fmt.Sprintf("── Diff: %s ──", diff.FilePath)
	if len(header) > width-2 {
		header = header[:width-2]
	}
	headerLine := styles.DiffHunkStyle.Render(" " + header)

	var lines []string
	lines = append(lines, headerLine)

	end := scrollOffset + maxLines
	if end > len(diff.Lines) {
		end = len(diff.Lines)
	}

	for i := scrollOffset; i < end; i++ {
		dl := diff.Lines[i]
		line := renderDiffLine(dl, width)
		lines = append(lines, line)
	}

	if end < len(diff.Lines) {
		remaining := len(diff.Lines) - end
		lines = append(lines, styles.HelpDescStyle.Render(
			fmt.Sprintf("  ... %d more lines (scroll with j/k)", remaining)))
	}

	return strings.Join(lines, "\n")
}

func renderDiffLine(dl git.DiffLine, width int) string {
	var lineNum string
	content := dl.Content

	// Truncate content to fit width
	maxContent := width - 10
	if maxContent < 20 {
		maxContent = 20
	}
	if len(content) > maxContent {
		content = content[:maxContent]
	}

	switch dl.Type {
	case git.DiffHunkHeader:
		return styles.DiffHunkStyle.Render(" " + dl.Content)

	case git.DiffAdd:
		if dl.NewNum > 0 {
			lineNum = fmt.Sprintf(" %4d ", dl.NewNum)
		} else {
			lineNum = "      "
		}
		return styles.DiffAddStyle.Render(lineNum + "+" + content)

	case git.DiffDelete:
		if dl.OldNum > 0 {
			lineNum = fmt.Sprintf(" %4d ", dl.OldNum)
		} else {
			lineNum = "      "
		}
		return styles.DiffDelStyle.Render(lineNum + "-" + content)

	case git.DiffContext:
		if dl.NewNum > 0 {
			lineNum = fmt.Sprintf(" %4d ", dl.NewNum)
		} else {
			lineNum = "      "
		}
		return fmt.Sprintf("%s %s", styles.HelpDescStyle.Render(lineNum), content)

	default:
		return " " + content
	}
}
