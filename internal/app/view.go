package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/4nkitd/git-panel/internal/git"
	"github.com/4nkitd/git-panel/internal/ui/components"
	"github.com/4nkitd/git-panel/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// View renders the entire UI and builds the layout map for mouse hit-testing.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if m.mode == ModeHelp {
		return components.RenderHelp(m.width, m.height)
	}
	if m.mode == ModeBranch {
		return m.renderBranchPicker()
	}
	if m.mode == ModeSettings {
		return m.renderSettings()
	}

	lm := NewLayoutMap(m.height + 100)

	// Bottom-fixed bars (separator + help + status = 3 lines)
	helpBar := components.RenderHelpBar(m.width, m.helpContext())
	statusBar := m.renderStatusBar()
	bottomFixed := 3

	// Graph (bottom-sticky)
	var graphLines []string
	hasGraph := m.logResult != nil && len(m.logResult.Commits) > 0
	if hasGraph {
		graphLines = m.renderGraphSectionLines()
	}

	upperAvailable := m.height - bottomFixed - len(graphLines)
	if upperAvailable < 5 {
		upperAvailable = 5
	}

	row := 0
	var lines []string

	// ── Title bar ──
	lines = append(lines, m.renderTitle())
	lm.Set(row, ZoneTitle, -1)
	row++

	// ── Separator ──
	lines = append(lines, thinSep(m.width))
	row++

	// ── Commit area ──
	commitStr := m.renderCommitArea()
	for _, cl := range strings.Split(commitStr, "\n") {
		lines = append(lines, cl)
		if strings.Contains(cl, "Commit") || strings.Contains(cl, "Generate") || strings.Contains(cl, "Amend") {
			lm.Set(row, ZoneCommitButton, -1)
		} else {
			lm.Set(row, ZoneCommitInput, -1)
		}
		row++
	}

	// ── Separator ──
	lines = append(lines, thinSep(m.width))
	row++

	// ── Staged Changes ──
	for _, l := range m.renderStagedSection(lm, &row) {
		lines = append(lines, l)
	}

	// ── Changes ──
	for _, l := range m.renderUnstagedSection(lm, &row) {
		lines = append(lines, l)
	}

	// ── Stashes (only if there are any) ──
	if m.status != nil && len(m.status.Stashes) > 0 {
		for _, l := range m.renderStashSection(lm, &row) {
			lines = append(lines, l)
		}
	}

	// ── Diff view ──
	if m.mode == ModeDiff && m.currentDiff != nil {
		lines = append(lines, thinSep(m.width))
		row++
		lm.Set(row, ZoneDiff, -1)
		diffStr := components.RenderDiff(m.currentDiff, m.width, m.diffScroll, m.diffMaxLines)
		for _, dl := range strings.Split(diffStr, "\n") {
			lines = append(lines, dl)
			lm.Set(row, ZoneDiff, -1)
			row++
		}
	}

	// ── Clean repo message ──
	if m.status != nil && len(m.status.Staged) == 0 && len(m.status.Unstaged) == 0 && m.mode != ModeDiff {
		lines = append(lines, "")
		row++
		emptyMsg := styles.HelpDescStyle.Render("  No pending changes")
		lines = append(lines, emptyMsg)
		row++
	}

	// ── Pad to push graph to bottom ──
	content := strings.Join(lines, "\n")
	pad := upperAvailable - row
	if pad > 0 {
		content += strings.Repeat("\n", pad)
		row += pad
	}

	// ── Graph (bottom-sticky separator + section) ──
	if hasGraph {
		content += "\n" + thinSep(m.width)
		row++
		for _, gl := range graphLines {
			content += "\n" + gl
			row++
		}
		graphStart := row - len(graphLines)
		m.mapGraphLayout(lm, graphStart, graphLines)
	}

	// ── Footer: separator + help bar + status bar ──
	content += "\n" + thinSep(m.width) + "\n" + helpBar + "\n" + statusBar

	lastLayout = lm
	return content
}

var lastLayout *LayoutMap

// ── Layout map for graph ──

func (m Model) mapGraphLayout(lm *LayoutMap, startRow int, graphLines []string) {
	if m.logResult == nil || len(graphLines) == 0 {
		return
	}

	row := startRow
	maxRow := startRow + len(graphLines)

	lm.Set(row, ZoneGraphHeader, -1)
	row++

	if m.graphCollapsed {
		return
	}

	maxVisible := m.graphMaxVisible
	if maxVisible > m.height*40/100 {
		maxVisible = m.height * 40 / 100
	}
	if maxVisible < 5 {
		maxVisible = 5
	}

	visibleStart := m.graphScroll
	visibleEnd := visibleStart + maxVisible
	if visibleEnd > len(m.logResult.Commits) {
		visibleEnd = len(m.logResult.Commits)
	}

	for i := visibleStart; i < visibleEnd && row < maxRow; i++ {
		lm.Set(row, ZoneGraphCommit, i)
		row++
		commit := m.logResult.Commits[i]
		if commit.Expanded {
			for range commit.Files {
				if row < maxRow {
					lm.Set(row, ZoneGraphFile, i)
					row++
				}
			}
		}
	}
	if row < maxRow {
		lm.Set(row, ZoneNone, -1)
	}
}

// ── Title bar ──

func (m Model) renderTitle() string {
	icon := styles.TitleStyle.Render("")
	title := styles.TitleTextStyle.Render("Source Control")
	left := icon + " " + title

	var right string
	if m.spinner.active {
		if m.generating {
			right = styles.SpinnerStyle.Render(m.spinner.AIView()) + " " +
				styles.HelpDescStyle.Render(m.loadingLabel)
		} else {
			right = styles.SpinnerStyle.Render(m.spinner.View()) + " " +
				styles.HelpDescStyle.Render(m.loadingLabel)
		}
	} else if m.successMsg != "" {
		right = styles.SuccessStyle.Render("✓ " + m.successMsg)
	} else if m.errMsg != "" {
		right = styles.ErrorStyle.Render("✗ " + m.errMsg)
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 1
	if gap < 0 {
		gap = 0
	}
	return left + strings.Repeat(" ", gap) + right
}

// ── Commit area ──

func (m Model) renderCommitArea() string {
	inner := m.width - 2 // border only, margin via MarginLeft

	if m.mode == ModeCommit {
		var s strings.Builder

		// Determine border color based on state
		borderColor := styles.BorderFocus
		if m.generating {
			borderColor = styles.Magenta
		}

		// Text input or generating indicator
		var inputContent string
		if m.generating {
			inputContent = styles.HelpDescStyle.Render("✦ Generating commit message...")
		} else {
			inputContent = m.commitInput.View(inner - 2) // -2 for padding
		}

		border := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Width(inner).
			Padding(0, 1).
			MarginLeft(1)
		s.WriteString(border.Render(inputContent))
		s.WriteString("\n")

		// Full-width commit button
		btnFg := styles.BrightWhite
		btnBg := styles.BgButton
		if m.generating {
			btnFg = styles.DimWhite
			btnBg = lipgloss.Color("#1a3a5c")
		}
		s.WriteString(lipgloss.NewStyle().
			Foreground(btnFg).
			Background(btnBg).
			Bold(true).
			Width(inner+2).
			Align(lipgloss.Center).
			MarginLeft(1).
			Render("✓ Commit"))
		s.WriteString("\n")

		// Action row: Generate + Amend + hint
		var parts []string
		if m.generating {
			parts = append(parts, styles.SpinnerStyle.Render(m.spinner.AIView()+" Generating..."))
		} else {
			parts = append(parts, styles.GenerateButtonStyle.Render("✦ Generate"))
		}
		parts = append(parts, styles.CommitSecondaryStyle.Render("Amend"))
		parts = append(parts, styles.HelpDescStyle.Render("esc:cancel"))
		s.WriteString(" " + strings.Join(parts, "  "))

		return s.String()
	}

	// ── Inactive: styled placeholder input + muted commit button ──
	placeholder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Foreground(styles.Subtle).
		Width(inner).
		Padding(0, 1).
		MarginLeft(1).
		Render("Commit message  (c: type  g: AI generate)")

	btn := lipgloss.NewStyle().
		Foreground(styles.DimWhite).
		Background(lipgloss.Color("#1a3a5c")).
		Width(inner+2).
		Align(lipgloss.Center).
		MarginLeft(1).
		Render("✓ Commit")

	return placeholder + "\n" + btn
}

// ── Staged Changes ──

func (m Model) renderStagedSection(lm *LayoutMap, row *int) []string {
	var lines []string
	count := 0
	if m.status != nil {
		count = len(m.status.Staged)
	}

	focused := m.focus == SectionStaged
	actions := ""
	if count > 0 {
		actions = styles.UnstageIconStyle.Render("−") + "  " + styles.DiscardIconStyle.Render("↺")
	}
	lines = append(lines, renderSectionHeader("Staged Changes", count, m.stagedCollapsed, m.width, actions, focused))
	lm.Set(*row, ZoneStagedHeader, -1)
	*row++

	if !m.stagedCollapsed && m.status != nil {
		for i, entry := range m.status.Staged {
			hovered := m.hoverRow == *row
			sel := i == m.cursor && focused
			lines = append(lines, renderFileLine(entry, m.width, sel, hovered, true))
			lm.SetWithAction(*row, ZoneStagedFile, i, m.width-10)
			*row++
		}
		if count == 0 {
			lines = append(lines, dim("  No staged changes", m.width))
			lm.Set(*row, ZoneNone, -1)
			*row++
		}
	}

	return lines
}

// ── Changes ──

func (m Model) renderUnstagedSection(lm *LayoutMap, row *int) []string {
	var lines []string
	count := 0
	if m.status != nil {
		count = len(m.status.Unstaged)
	}

	focused := m.focus == SectionUnstaged
	actions := ""
	if count > 0 {
		actions = styles.StageIconStyle.Render("+") + "  " + styles.DiscardIconStyle.Render("↺")
	}
	lines = append(lines, renderSectionHeader("Changes", count, m.unstagedCollapsed, m.width, actions, focused))
	lm.Set(*row, ZoneUnstagedHeader, -1)
	*row++

	if !m.unstagedCollapsed && m.status != nil {
		for i, entry := range m.status.Unstaged {
			hovered := m.hoverRow == *row
			sel := i == m.cursor && focused
			lines = append(lines, renderFileLine(entry, m.width, sel, hovered, false))
			lm.SetWithAction(*row, ZoneUnstagedFile, i, m.width-10)
			*row++
		}
		if count == 0 {
			lines = append(lines, dim("  No changes", m.width))
			lm.Set(*row, ZoneNone, -1)
			*row++
		}
	}

	return lines
}

// ── Stashes ──

func (m Model) renderStashSection(lm *LayoutMap, row *int) []string {
	var lines []string
	count := 0
	if m.status != nil {
		count = len(m.status.Stashes)
	}

	focused := m.focus == SectionStashes
	lines = append(lines, renderSectionHeader("Stashes", count, m.stashCollapsed, m.width, "", focused))
	lm.Set(*row, ZoneStashHeader, -1)
	*row++

	if !m.stashCollapsed && m.status != nil {
		for i, stash := range m.status.Stashes {
			line := fmt.Sprintf("  stash@{%d}: %s", stash.Index, stash.Message)
			if focused && i == m.cursor {
				line = padW(line, m.width)
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

// ── Graph (bottom-sticky) ──

func (m Model) renderGraphSectionLines() []string {
	var lines []string

	header := components.RenderGraphSection(m.graphCollapsed, m.width, m.focus == SectionGraph)
	lines = append(lines, header)

	if m.graphCollapsed || m.logResult == nil || len(m.logResult.Commits) == 0 {
		return lines
	}

	maxVis := m.graphMaxVisible
	if maxVis > m.height*40/100 {
		maxVis = m.height * 40 / 100
	}
	if maxVis < 5 {
		maxVis = 5
	}

	start := m.graphScroll
	end := start + maxVis
	if end > len(m.logResult.Commits) {
		end = len(m.logResult.Commits)
	}

	headBranch := ""
	if m.status != nil {
		headBranch = m.status.Branch.Name
	}

	for i := start; i < end; i++ {
		c := m.logResult.Commits[i]
		sel := m.focus == SectionGraph && i == m.graphCursor

		isHead := false
		for _, ref := range c.Refs {
			if strings.Contains(ref, "HEAD") || ref == headBranch {
				isHead = true
				break
			}
		}

		lines = append(lines, components.RenderCommitLine(c, m.width, sel, false, isHead))
		if c.Expanded && len(c.Files) > 0 {
			lines = append(lines, components.RenderCommitFiles(c.Files, c.BranchIdx, m.width)...)
		}
	}

	if end < len(m.logResult.Commits) {
		n := len(m.logResult.Commits) - end
		lines = append(lines, styles.HelpDescStyle.Render(fmt.Sprintf("  ... %d more commits", n)))
	}

	return lines
}

// ── Status bar ──

func (m Model) renderStatusBar() string {
	branch := m.getBranch()
	staged, unstaged := 0, 0
	if m.status != nil {
		staged = len(m.status.Staged)
		unstaged = len(m.status.Unstaged)
	}
	return components.RenderStatusBar(branch, staged, unstaged, m.loading, m.errMsg, m.width)
}

// ── Branch picker ──

func (m Model) renderBranchPicker() string {
	var lines []string
	lines = append(lines, styles.TitleTextStyle.Render(" Switch Branch"))
	lines = append(lines, "")

	for i, b := range m.branches {
		cur := ""
		if m.status != nil && b == m.status.Branch.Name {
			cur = styles.HelpDescStyle.Render(" (current)")
		}
		if i == m.branchCursor {
			marker := styles.BranchStyle.Render(">")
			line := " " + marker + " " + styles.FileNameStyle.Render(b) + cur
			line = padW(line, m.width-8)
			lines = append(lines, styles.SelectedStyle.Render(line))
		} else {
			lines = append(lines, "   "+styles.FilePathStyle.Render(b)+cur)
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpDescStyle.Render("  enter:checkout  esc:cancel"))

	content := strings.Join(lines, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderFocus).
		Width(m.width - 4).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box.Render(content))
}

// ── Settings panel ──

func (m Model) renderSettings() string {
	var lines []string

	lines = append(lines, styles.TitleTextStyle.Render("  Settings"))
	lines = append(lines, "")

	// Current config
	lines = append(lines, styles.SectionHeaderStyle.Render("  Ollama Configuration"))
	lines = append(lines, "")
	lines = append(lines,
		styles.HelpDescStyle.Render("  Host:  ")+styles.FileNameStyle.Render(m.ollamaCfg.Host))
	lines = append(lines,
		styles.HelpDescStyle.Render("  Model: ")+styles.BranchStyle.Render(m.ollamaCfg.Model))
	lines = append(lines, "")

	// Separator
	lines = append(lines, styles.SectionSepStyle.Render(strings.Repeat("─", m.width-6)))
	lines = append(lines, "")

	// Model picker
	lines = append(lines, styles.SectionHeaderStyle.Render("  Select Model"))
	lines = append(lines, "")

	if !m.modelListLoaded {
		lines = append(lines, styles.SpinnerStyle.Render("  "+m.spinner.View()+" Loading models from Ollama..."))
	} else if len(m.availableModels) == 0 {
		lines = append(lines, styles.ErrorStyle.Render("  No models found. Is Ollama running?"))
		lines = append(lines, styles.HelpDescStyle.Render("  r: retry"))
	} else {
		for i, model := range m.availableModels {
			current := ""
			if model == m.ollamaCfg.Model {
				current = styles.SuccessStyle.Render(" (active)")
			}

			if i == m.settingsCursor {
				marker := styles.BranchStyle.Render(">")
				line := "  " + marker + " " + styles.FileNameStyle.Render(model) + current
				line = padW(line, m.width-6)
				lines = append(lines, styles.SelectedStyle.Render(line))
			} else {
				lines = append(lines, "    "+styles.FilePathStyle.Render(model)+current)
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.SectionSepStyle.Render(strings.Repeat("─", m.width-6)))
	lines = append(lines, "")

	// Environment hints
	lines = append(lines, styles.SectionHeaderStyle.Render("  Environment Variables"))
	lines = append(lines, "")
	lines = append(lines, styles.HelpDescStyle.Render("  OLLAMA_HOST  — Ollama server URL"))
	lines = append(lines, styles.HelpDescStyle.Render("  OLLAMA_MODEL — Default model name"))

	lines = append(lines, "")
	lines = append(lines, styles.HelpDescStyle.Render("  enter:select  r:refresh  esc:close"))

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Width(m.width - 4).
		Padding(1, 1)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box.Render(content))
}

// ── Context helpers ──

func (m Model) helpContext() components.HelpContext {
	ctx := components.HelpContext{Generating: m.generating}
	switch m.mode {
	case ModeCommit:
		ctx.Mode = "commit"
	case ModeDiff:
		ctx.Mode = "diff"
	case ModeBranch:
		ctx.Mode = "branch"
	case ModeHelp:
		ctx.Mode = "help"
	case ModeSettings:
		ctx.Mode = "settings"
	default:
		ctx.Mode = "normal"
	}
	switch m.focus {
	case SectionStaged:
		ctx.Section = "staged"
	case SectionUnstaged:
		ctx.Section = "unstaged"
	case SectionStashes:
		ctx.Section = "stashes"
	case SectionGraph:
		ctx.Section = "graph"
	}
	if m.status != nil {
		ctx.HasStaged = len(m.status.Staged) > 0
		ctx.HasChanges = len(m.status.Unstaged) > 0
	}
	return ctx
}

func (m Model) getBranch() git.BranchInfo {
	if m.status != nil {
		return m.status.Branch
	}
	return git.BranchInfo{Name: "unknown"}
}

// ── Rendering primitives ──

func renderSectionHeader(title string, count int, collapsed bool, width int, actions string, focused bool) string {
	arrow := "▾"
	if collapsed {
		arrow = "▸"
	}
	arrow = styles.HelpDescStyle.Render(arrow)

	countStr := styles.SectionCountStyle.Render(fmt.Sprintf("%d", count))
	header := fmt.Sprintf(" %s %s %s", arrow, styles.SectionHeaderStyle.Render(title), countStr)

	if actions != "" {
		aw := lipgloss.Width(actions)
		hw := lipgloss.Width(header)
		gap := width - hw - aw - 2
		if gap > 0 {
			header += strings.Repeat(" ", gap) + actions + " "
		}
	}

	header = padW(header, width)
	if focused {
		header = styles.SelectedStyle.Render(header)
	}
	return header
}

func renderFileLine(entry git.StatusEntry, width int, selected bool, hovered bool, isStaged bool) string {
	// Status badge
	st := entry.Status.String()
	badge := styles.StatusStyle(entry.Status).Render(st)

	// Split into filename (bold) + directory (dim) like VS Code
	name := filepath.Base(entry.Path)
	dir := filepath.Dir(entry.Path)
	if dir == "." {
		dir = ""
	}

	nameStyled := styles.FileNameStyle.Render(name)
	dirStyled := ""
	if dir != "" {
		dirStyled = " " + styles.FileDirStyle.Render(dir)
	}

	// Action icons on hover/select
	actions := ""
	if hovered || selected {
		if isStaged {
			actions = styles.UnstageIconStyle.Render("−") + " " + styles.DiscardIconStyle.Render("↺")
		} else {
			actions = styles.StageIconStyle.Render("+") + " " + styles.DiscardIconStyle.Render("↺")
		}
	}

	// Layout: "  M  filename  dir                      + ↺"
	left := fmt.Sprintf("  %s  %s%s", badge, nameStyled, dirStyled)
	lw := lipgloss.Width(left)
	aw := lipgloss.Width(actions)

	// Truncate if needed
	maxLeft := width - aw - 2
	if lw > maxLeft && maxLeft > 10 {
		// Re-render with truncated dir
		maxDir := maxLeft - lipgloss.Width(fmt.Sprintf("  %s  %s ", badge, nameStyled))
		if maxDir > 3 && dir != "" {
			if len(dir) > maxDir {
				dir = dir[:maxDir-2] + ".."
			}
			dirStyled = " " + styles.FileDirStyle.Render(dir)
		} else {
			dirStyled = ""
		}
		left = fmt.Sprintf("  %s  %s%s", badge, nameStyled, dirStyled)
		lw = lipgloss.Width(left)
	}

	gap := width - lw - aw - 1
	if gap < 1 {
		gap = 1
	}

	line := left + strings.Repeat(" ", gap) + actions
	line = padW(line, width)

	if selected {
		return styles.SelectedStyle.Render(line)
	} else if hovered {
		return styles.HoverStyle.Render(line)
	}
	return line
}

func thinSep(width int) string {
	return styles.SectionSepStyle.Render(strings.Repeat("─", width))
}

func dim(s string, _ int) string {
	return styles.HelpDescStyle.Render(s)
}

func padW(s string, width int) string {
	w := lipgloss.Width(s)
	if w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}
