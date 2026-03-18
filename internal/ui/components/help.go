package components

import (
	"strings"

	"github.com/ankityadav/zedgit/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

type KeyBinding struct {
	Key  string
	Desc string
}

var Bindings = []KeyBinding{
	{"j/k", "Navigate up/down"},
	{"Tab", "Switch section"},
	{"Enter", "Stage/unstage file"},
	{"a", "Stage all"},
	{"A", "Unstage all"},
	{"d", "Show diff"},
	{"c", "Start commit"},
	{"Enter", "Confirm commit (in commit mode)"},
	{"Esc", "Cancel / close"},
	{"b", "Branch picker"},
	{"p", "Push"},
	{"P", "Pull"},
	{"f", "Fetch"},
	{"s", "Stash"},
	{"S", "Stash pop"},
	{"z", "Undo last commit"},
	{"r", "Refresh"},
	{"?", "Toggle help"},
	{"q", "Quit"},
}

// RenderHelp renders the help overlay.
func RenderHelp(width, height int) string {
	title := styles.TitleStyle.Render("Keybindings")

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	for _, b := range Bindings {
		key := styles.HelpKeyStyle.Width(8).Align(lipgloss.Right).Render(b.Key)
		desc := styles.HelpDescStyle.Render("  " + b.Desc)
		lines = append(lines, "  "+key+desc)
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpDescStyle.Render("  Press ? or Esc to close"))

	content := strings.Join(lines, "\n")

	boxWidth := width - 4
	if boxWidth > 50 {
		boxWidth = 50
	}
	boxHeight := len(lines) + 2
	if boxHeight > height-4 {
		boxHeight = height - 4
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Width(boxWidth).
		Height(boxHeight).
		Padding(1, 2)

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		box.Render(content))
}

// RenderHelpBar renders the compact help bar at the bottom.
func RenderHelpBar(width int) string {
	hints := []string{"j/k:nav", "Enter:stage", "c:commit", "d:diff", "?:help", "q:quit"}
	return styles.HelpDescStyle.Render(" " + strings.Join(hints, " │ "))
}
