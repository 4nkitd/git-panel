package components

import (
	"fmt"
	"strings"

	"github.com/ankityadav/zedgit/internal/git"
	"github.com/ankityadav/zedgit/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// RenderStatusBar renders the bottom status bar.
func RenderStatusBar(branch git.BranchInfo, loading bool, errMsg string, width int) string {
	bar := styles.StatusBarStyle.Width(width)

	var parts []string

	// Branch name
	branchIcon := ""
	if branch.IsDetached {
		branchIcon = "⊘"
	}
	parts = append(parts, styles.BranchStyle.Render(branchIcon+" "+branch.Name))

	// Ahead/behind
	if branch.Upstream != "" {
		sync := ""
		if branch.Ahead > 0 {
			sync += fmt.Sprintf("↑%d", branch.Ahead)
		}
		if branch.Behind > 0 {
			if sync != "" {
				sync += " "
			}
			sync += fmt.Sprintf("↓%d", branch.Behind)
		}
		if sync == "" {
			sync = "✓"
		}
		parts = append(parts, sync)
	}

	// Loading indicator
	if loading {
		parts = append(parts, styles.SpinnerStyle.Render("⟳"))
	}

	left := strings.Join(parts, " │ ")

	// Error or help hint on right
	right := "?:help"
	if errMsg != "" {
		right = styles.ErrorStyle.Render(truncate(errMsg, width/2))
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	content := " " + left + strings.Repeat(" ", gap) + right + " "
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
