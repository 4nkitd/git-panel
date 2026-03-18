package components

import (
	"fmt"
	"strings"

	"github.com/ankityadav/zedgit/internal/git"
	"github.com/ankityadav/zedgit/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// RenderStatusBar renders the bottom status bar with branch, remote sync, and change summary.
func RenderStatusBar(branch git.BranchInfo, staged int, unstaged int, loading bool, errMsg string, width int) string {
	bar := styles.StatusBarStyle.Width(width)
	sep := styles.StatusBarSepStyle.Render(" │ ")

	var parts []string

	// Branch icon + name
	icon := ""
	if branch.IsDetached {
		icon = "⊘ "
	}
	parts = append(parts, styles.BranchStyle.Render(icon+branch.Name))

	// Remote sync status — show commit diff from remote
	if branch.Upstream != "" {
		var syncParts []string

		if branch.Ahead > 0 {
			label := "commit"
			if branch.Ahead > 1 {
				label = "commits"
			}
			syncParts = append(syncParts,
				styles.SuccessStyle.Render(fmt.Sprintf("↑ %d %s to push", branch.Ahead, label)))
		}
		if branch.Behind > 0 {
			label := "commit"
			if branch.Behind > 1 {
				label = "commits"
			}
			syncParts = append(syncParts,
				styles.SpinnerStyle.Render(fmt.Sprintf("↓ %d %s to pull", branch.Behind, label)))
		}

		if len(syncParts) > 0 {
			parts = append(parts, strings.Join(syncParts, "  "))
		} else {
			parts = append(parts, styles.HelpDescStyle.Render("✓ in sync"))
		}
	} else {
		parts = append(parts, styles.HelpDescStyle.Render("no remote"))
	}

	// Change counts
	if staged > 0 || unstaged > 0 {
		var counts []string
		if staged > 0 {
			counts = append(counts, styles.StageIconStyle.Render(fmt.Sprintf("+%d staged", staged)))
		}
		if unstaged > 0 {
			counts = append(counts, styles.StatusStyle(git.StatusModified).Render(fmt.Sprintf("~%d changed", unstaged)))
		}
		parts = append(parts, strings.Join(counts, " "))
	}

	left := " " + strings.Join(parts, sep)

	// Right side — error if any
	right := " "
	if errMsg != "" {
		right = " " + styles.ErrorStyle.Render(truncate(errMsg, width/3)) + " "
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	content := left + strings.Repeat(" ", gap) + right
	return bar.Render(content)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 4 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
