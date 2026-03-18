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
	{"g", "Generate commit msg (Ollama AI)"},
	{"Ctrl+Enter", "Confirm commit"},
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

// HelpContext provides current UI state for context-aware help bar.
type HelpContext struct {
	Mode       string // "normal", "commit", "diff", "branch", "help"
	Section    string // "staged", "unstaged", "stashes", "graph"
	HasStaged  bool
	HasChanges bool
	Generating bool
}

// RenderHelpBar renders a context-aware shortcut bar at the bottom.
func RenderHelpBar(width int, ctx HelpContext) string {
	var hints []string

	switch ctx.Mode {
	case "commit":
		hints = []string{"Ctrl+Enter:commit", "Esc:cancel"}
		if !ctx.Generating {
			hints = append(hints, "g:AI generate")
		} else {
			hints = append(hints, "✦ generating...")
		}

	case "diff":
		hints = []string{"j/k:scroll", "Esc:close", "d:close"}

	case "branch":
		hints = []string{"j/k:navigate", "Enter:checkout", "Esc:cancel"}

	case "help":
		hints = []string{"Esc:close", "?:close"}

	case "settings":
		hints = []string{"j/k:navigate", "Enter:select model", "r:refresh", "Esc:close"}

	default: // normal mode
		hints = []string{"j/k:nav", "Tab:section"}

		switch ctx.Section {
		case "staged":
			hints = append(hints, "Enter:unstage", "A:unstage all", "d:diff")
		case "unstaged":
			hints = append(hints, "Enter:stage", "a:stage all", "d:diff")
		case "stashes":
			hints = append(hints, "s:stash", "S:pop")
		case "graph":
			hints = append(hints, "Enter:expand", "scroll:↑↓")
		}

		// Common actions available in normal mode
		if ctx.HasStaged {
			hints = append(hints, "c:commit", "g:AI gen")
		} else if ctx.HasChanges {
			hints = append(hints, "a:stage all", "g:AI gen")
		}

		hints = append(hints, "b:branch", "p:push", ",:settings", "?:help")
	}

	// Render each hint with key highlighted
	var rendered []string
	for _, h := range hints {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			rendered = append(rendered,
				styles.HelpBarKeyStyle.Render(parts[0])+
					styles.HelpBarStyle.Render(":"+parts[1]))
		} else {
			rendered = append(rendered, styles.HelpBarStyle.Render(h))
		}
	}

	sep := styles.HelpBarStyle.Render(" · ")
	bar := " " + strings.Join(rendered, sep)

	// Truncate if too wide
	if lipgloss.Width(bar) > width {
		bar = bar[:width-1]
	}

	return bar
}
