package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ankityadav/zedgit/internal/git"
	"github.com/ankityadav/zedgit/internal/ui/components"
	"github.com/ankityadav/zedgit/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// View renders the entire UI and builds the layout map for mouse hit-testing.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Help overlay
	if m.mode == ModeHelp {
		return components.RenderHelp(m.width, m.height)
	}

	// Branch picker overlay
	if m.mode == ModeBranch {
		return m.renderBranchPicker()
	}

	lm := NewLayoutMap(m.height)
	row := 0

	var lines []string

	// ── Title bar ──
	titleLine := m.renderTitle()
	lines = append(lines, titleLine)
	lm.Set(row, ZoneTitle, -1)
	row++

	// ── Commit input area ──
	commitLines := m.renderCommitArea()
	for _, cl := range strings.Split(commitLines, "\n") {
		lines = append(lines, cl)
		if strings.Contains(cl, "Commit") || strings.Contains(cl, "✓ Commit") {
			lm.Set(row, ZoneCommitButton, -1)
		} else {
			lm.Set(row, ZoneCommitInput, -1)
		}
		row++
	}

	// ── Staged Changes section ──
	stagedLines := m.renderStagedSection(lm, &row)
	lines = append(lines, stagedLines...)

	// ── Unstaged Changes section ──
	unstagedLines := m.renderUnstagedSection(lm, &row)
	lines = append(lines, unstagedLines...)

	// ── Stash section ──
	stashLines := m.renderStashSection(lm, &row)
	lines = append(lines, stashLines...)

	// ── Diff view ──
	if m.mode == ModeDiff && m.currentDiff != nil {
		lines = append(lines, "")
		lm.Set(row, ZoneDiff, -1)
		row++
		diffStr := components.RenderDiff(m.currentDiff, m.width, m.diffScroll, m.diffMaxLines)
		for _, dl := range strings.Split(diffStr, "\n") {
			lines = append(lines, dl)
			lm.Set(row, ZoneDiff, -1)
			row++
		}
	}

	// Join content
	content := strings.Join(lines, "\n")

	// Pad to push status bar to bottom
	contentHeight := row
	helpBar := components.RenderHelpBar(m.width)
	statusBar := m.renderStatusBar()

	paddingNeeded := m.height - contentHeight - 2
	if paddingNeeded > 0 {
		content += strings.Repeat("\n", paddingNeeded)
		row += paddingNeeded
	}

	lm.Set(m.height-2, ZoneHelpBar, -1)
	lm.Set(m.height-1, ZoneStatusBar, -1)

	content += "\n" + helpBar + "\n" + statusBar

	// Store layout map (we return a new model from View via pointer trick)
	// Since View is called on value receiver, we store it via the pointer in the map
	// Actually, we store it in the model during Update via a post-render message.
	// For simplicity, we use a package-level var (safe because TUI is single-threaded).
	lastLayout = lm

	return content
}

// lastLayout is the most recently built layout map. Safe because bubbletea is single-threaded.
var lastLayout *LayoutMap

func (m Model) renderTitle() string {
	title := styles.TitleStyle.Render("SOURCE CONTROL")

	var rightParts []string
	if m.loading {
		rightParts = append(rightParts, styles.SpinnerStyle.Render("⟳"))
	}
	if m.successMsg != "" {
		rightParts = append(rightParts, styles.SuccessStyle.Render("✓ "+m.successMsg))
	}
	right := strings.Join(rightParts, " ")

	gap := m.width - lipgloss.Width(title) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return title + strings.Repeat(" ", gap) + right
}

func (m Model) renderCommitArea() string {
	var s strings.Builder

	if m.mode == ModeCommit {
		s.WriteString(styles.CommitInputStyle.Width(m.width - 2).Render(m.commitInput.View()))
		s.WriteString("\n")

		// Commit button row (clickable)
		btn := styles.CommitButtonStyle.Render(" ✓ Commit ")
		amendBtn := styles.CommitSecondaryStyle.Render(" Amend ")
		cancelHint := styles.HelpDescStyle.Render(" Esc: cancel")
		s.WriteString(" " + btn + " " + amendBtn + cancelHint)
	} else {
		placeholder := styles.CommitInputStyle.
			Width(m.width - 2).
			Foreground(styles.DimWhite).
			Render("Press 'c' to commit...")
		s.WriteString(placeholder)
	}

	return s.String()
}

func (m Model) renderStagedSection(lm *LayoutMap, row *int) []string {
	var lines []string

	count := 0
	if m.status != nil {
		count = len(m.status.Staged)
	}

	// Section header with action icons
	actionIcons := ""
	if count > 0 {
		actionIcons = styles.UnstageIconStyle.Render(" − ") + " " + styles.DiscardIconStyle.Render(" ↺ ")
	}
	header := renderSectionHeaderWithActions(
		"Staged Changes", count, m.stagedCollapsed, m.width,
		actionIcons, m.focus == SectionStaged)
	lines = append(lines, header)
	lm.Set(*row, ZoneStagedHeader, -1)
	*row++

	if !m.stagedCollapsed && m.status != nil {
		for i, entry := range m.status.Staged {
			hovered := m.isRowHovered(*row)
			line := renderFileLineWithActions(entry, m.width, i == m.cursor && m.focus == SectionStaged, hovered, true)
			lines = append(lines, line)
			// Action icons start near right edge
			lm.SetWithAction(*row, ZoneStagedFile, i, m.width-10)
			*row++
		}
		if len(m.status.Staged) == 0 {
			lines = append(lines, styles.HelpDescStyle.Render("  No staged changes"))
			lm.Set(*row, ZoneNone, -1)
			*row++
		}
	}

	return lines
}

func (m Model) renderUnstagedSection(lm *LayoutMap, row *int) []string {
	var lines []string

	count := 0
	if m.status != nil {
		count = len(m.status.Unstaged)
	}

	actionIcons := ""
	if count > 0 {
		actionIcons = styles.StageIconStyle.Render(" + ") + " " + styles.DiscardIconStyle.Render(" ↺ ")
	}
	header := renderSectionHeaderWithActions(
		"Changes", count, m.unstagedCollapsed, m.width,
		actionIcons, m.focus == SectionUnstaged)
	lines = append(lines, header)
	lm.Set(*row, ZoneUnstagedHeader, -1)
	*row++

	if !m.unstagedCollapsed && m.status != nil {
		for i, entry := range m.status.Unstaged {
			hovered := m.isRowHovered(*row)
			line := renderFileLineWithActions(entry, m.width, i == m.cursor && m.focus == SectionUnstaged, hovered, false)
			lines = append(lines, line)
			lm.SetWithAction(*row, ZoneUnstagedFile, i, m.width-10)
			*row++
		}
		if len(m.status.Unstaged) == 0 {
			lines = append(lines, styles.HelpDescStyle.Render("  No changes"))
			lm.Set(*row, ZoneNone, -1)
			*row++
		}
	}

	return lines
}

func (m Model) renderStashSection(lm *LayoutMap, row *int) []string {
	var lines []string

	count := 0
	if m.status != nil {
		count = len(m.status.Stashes)
	}

	header := renderSectionHeaderWithActions(
		"Stashes", count, m.stashCollapsed, m.width,
		"", m.focus == SectionStashes)
	lines = append(lines, header)
	lm.Set(*row, ZoneStashHeader, -1)
	*row++

	if !m.stashCollapsed && m.status != nil {
		for i, stash := range m.status.Stashes {
			line := fmt.Sprintf("  stash@{%d}: %s", stash.Index, stash.Message)
			if m.focus == SectionStashes && i == m.cursor {
				line = padToWidth(line, m.width)
				line = styles.SelectedStyle.Render(line)
			} else {
				line = styles.HelpDescStyle.Render(line)
			}
			lines = append(lines, line)
			lm.Set(*row, ZoneStashFile, i)
			*row++
		}
	}

	return lines
}

func (m Model) renderStatusBar() string {
	branch := m.getBranch()
	return components.RenderStatusBar(branch, m.loading, m.errMsg, m.width)
}

func (m Model) renderBranchPicker() string {
	var lines []string
	lines = append(lines, styles.TitleStyle.Render("Switch Branch"))
	lines = append(lines, "")

	for i, b := range m.branches {
		current := ""
		if m.status != nil && b == m.status.Branch.Name {
			current = " (current)"
		}

		if i == m.branchCursor {
			line := "> " + b + current
			line = padToWidth(line, m.width-8)
			lines = append(lines, styles.SelectedStyle.Render(line))
		} else {
			lines = append(lines, "  "+b+current)
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpDescStyle.Render("  Enter: checkout │ Esc: cancel"))

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Width(m.width - 4).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box.Render(content))
}

func (m Model) getBranch() git.BranchInfo {
	if m.status != nil {
		return m.status.Branch
	}
	return git.BranchInfo{Name: "unknown"}
}

func (m Model) isRowHovered(row int) bool {
	return m.hoverRow == row
}

// ── Rendering helpers ──

func renderSectionHeaderWithActions(title string, count int, collapsed bool, width int, actionIcons string, focused bool) string {
	arrow := "▼"
	if collapsed {
		arrow = "▶"
	}

	countStr := styles.SectionCountStyle.Render(fmt.Sprintf("(%d)", count))
	header := fmt.Sprintf(" %s %s %s", arrow, styles.SectionHeaderStyle.Render(title), countStr)

	if actionIcons != "" {
		iconsWidth := lipgloss.Width(actionIcons)
		headerWidth := lipgloss.Width(header)
		gap := width - headerWidth - iconsWidth - 1
		if gap > 0 {
			header += strings.Repeat(" ", gap) + actionIcons
		}
	}

	if focused {
		header = padToWidth(header, width)
		header = styles.SelectedStyle.Render(header)
	}

	return header
}

func renderFileLineWithActions(entry git.StatusEntry, width int, selected bool, hovered bool, isStaged bool) string {
	statusStr := entry.Status.String()
	statusStyled := styles.StatusStyle(entry.Status).Render(statusStr)

	// File path display
	display := entry.Path
	actionWidth := 10 // reserve space for action icons
	maxPath := width - 6 - actionWidth
	if maxPath < 10 {
		maxPath = 10
	}

	if width < 50 {
		display = filepath.Base(entry.Path)
	}
	if len(display) > maxPath {
		display = "..." + display[len(display)-maxPath+3:]
	}

	line := fmt.Sprintf("  %s  %s", statusStyled, styles.FilePathStyle.Render(display))

	// Action icons on hover or selection (VS Code shows on hover)
	actions := ""
	if hovered || selected {
		if isStaged {
			actions = styles.UnstageIconStyle.Render("−") + " " + styles.DiscardIconStyle.Render("↺")
		} else {
			actions = styles.StageIconStyle.Render("+") + " " + styles.DiscardIconStyle.Render("↺")
		}
	}

	// Pad and place actions at right edge
	lineWidth := lipgloss.Width(line)
	actionsWidth := lipgloss.Width(actions)
	gap := width - lineWidth - actionsWidth - 1
	if gap > 0 {
		line += strings.Repeat(" ", gap) + actions
	} else if gap <= 0 && actions != "" {
		// Truncate path more to fit actions
		line = line[:max(0, len(line)+gap-1)] + " " + actions
	}

	// Full-width pad for background
	line = padToWidth(line, width)

	if selected {
		line = styles.SelectedStyle.Render(line)
	} else if hovered {
		line = styles.HoverStyle.Render(line)
	}

	return line
}

func padToWidth(s string, width int) string {
	w := lipgloss.Width(s)
	if w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}
