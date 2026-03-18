package components

import (
	"fmt"
	"strings"

	"github.com/ankityadav/zedgit/internal/git"
	"github.com/ankityadav/zedgit/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// RenderDiff renders a diff with a header, colored lines, and scroll info.
func RenderDiff(diff *git.DiffResult, width int, scrollOffset int, maxLines int) string {
	if diff == nil || len(diff.Lines) == 0 {
		return styles.HelpDescStyle.Render("  No diff available")
	}

	// Header
	title := " Diff: " + diff.FilePath + " "
	totalLines := len(diff.Lines)
	position := fmt.Sprintf(" %d/%d ", scrollOffset+1, totalLines)
	headerGap := width - lipgloss.Width(title) - lipgloss.Width(position) - 2
	if headerGap < 0 {
		headerGap = 0
	}
	headerLine := styles.DiffHunkStyle.Render(title) +
		styles.SectionSepStyle.Render(strings.Repeat("─", headerGap)) +
		styles.HelpDescStyle.Render(position)

	var lines []string
	lines = append(lines, headerLine)

	// Visible range
	end := scrollOffset + maxLines
	if end > totalLines {
		end = totalLines
	}

	addCount, delCount := 0, 0
	for i := scrollOffset; i < end; i++ {
		dl := diff.Lines[i]
		lines = append(lines, renderDiffLine(dl, width))
		if dl.Type == git.DiffAdd {
			addCount++
		} else if dl.Type == git.DiffDelete {
			delCount++
		}
	}

	// Footer with stats
	footer := " "
	if addCount > 0 {
		footer += styles.StageIconStyle.Render(fmt.Sprintf("+%d", addCount)) + " "
	}
	if delCount > 0 {
		footer += styles.DiscardIconStyle.Render(fmt.Sprintf("-%d", delCount)) + " "
	}
	if end < totalLines {
		footer += styles.HelpDescStyle.Render(fmt.Sprintf("  ↓ %d more lines", totalLines-end))
	}
	lines = append(lines, footer)

	return strings.Join(lines, "\n")
}

func renderDiffLine(dl git.DiffLine, width int) string {
	content := dl.Content
	maxContent := width - 9
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
		num := "     "
		if dl.NewNum > 0 {
			num = fmt.Sprintf("%4d ", dl.NewNum)
		}
		return styles.DiffAddStyle.Render(" " + num + "+" + content)

	case git.DiffDelete:
		num := "     "
		if dl.OldNum > 0 {
			num = fmt.Sprintf("%4d ", dl.OldNum)
		}
		return styles.DiffDelStyle.Render(" " + num + "-" + content)

	case git.DiffContext:
		num := "     "
		if dl.NewNum > 0 {
			num = fmt.Sprintf("%4d ", dl.NewNum)
		}
		return styles.HelpDescStyle.Render(" "+num) + " " + content

	default:
		return " " + content
	}
}
