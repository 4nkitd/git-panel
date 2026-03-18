package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.commitInput.SetWidth(m.width - 4)
		m.diffMaxLines = m.height / 3
		return m, nil

	case statusMsg:
		m.status = msg.status
		m.loading = false
		m.clampCursor()
		return m, nil

	case errMsg:
		m.loading = false
		m.errMsg = msg.err.Error()
		return m, m.clearMessage()

	case successMsgType:
		m.successMsg = msg.msg
		return m, m.clearMessage()

	case diffMsg:
		m.currentDiff = msg.diff
		m.diffScroll = 0
		m.mode = ModeDiff
		return m, nil

	case branchesMsg:
		m.branches = msg.branches
		m.branchCursor = 0
		m.mode = ModeBranch
		return m, nil

	case gitOpDone:
		m.loading = false
		m.successMsg = msg.msg
		return m, tea.Batch(m.refreshStatus(), m.clearMessage())

	case clearMsgTick:
		m.errMsg = ""
		m.successMsg = ""
		return m, nil

	case refreshTick:
		return m, tea.Batch(m.refreshStatus(), m.tickRefresh())

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Pass through to textarea when in commit mode
	if m.mode == ModeCommit {
		var cmd tea.Cmd
		m.commitInput, cmd = m.commitInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	// Mode-specific handling
	switch m.mode {
	case ModeCommit:
		return m.handleCommitKey(msg)
	case ModeDiff:
		return m.handleDiffKey(msg)
	case ModeBranch:
		return m.handleBranchKey(msg)
	case ModeHelp:
		return m.handleHelpKey(msg)
	default:
		return m.handleNormalKey(msg)
	}
}

func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "?":
		m.mode = ModeHelp
		return m, nil

	// Navigation
	case "j", "down":
		m.cursor++
		m.clampCursor()
		return m, nil

	case "k", "up":
		m.cursor--
		m.clampCursor()
		return m, nil

	case "tab":
		m.cycleSection(true)
		return m, nil

	case "shift+tab":
		m.cycleSection(false)
		return m, nil

	// Stage/unstage
	case "enter", " ":
		return m.handleStageToggle()

	case "a":
		m.loading = true
		return m, m.doGitOp(func() error {
			return m.repo.StageAll()
		}, "Staged all changes")

	case "A":
		m.loading = true
		return m, m.doGitOp(func() error {
			return m.repo.UnstageAll()
		}, "Unstaged all changes")

	// Commit
	case "c":
		m.mode = ModeCommit
		m.commitInput.Focus()
		return m, m.commitInput.Focus()

	// Diff
	case "d":
		return m.handleShowDiff()

	// Branch
	case "b":
		return m, m.loadBranches()

	// Remote ops
	case "p":
		m.loading = true
		return m, m.doGitOp(func() error {
			return m.repo.Push()
		}, "Pushed successfully")

	case "P":
		m.loading = true
		return m, m.doGitOp(func() error {
			return m.repo.Pull()
		}, "Pulled successfully")

	case "f":
		m.loading = true
		return m, m.doGitOp(func() error {
			return m.repo.Fetch()
		}, "Fetched successfully")

	// Stash
	case "s":
		m.loading = true
		return m, m.doGitOp(func() error {
			return m.repo.Stash("")
		}, "Stashed changes")

	case "S":
		m.loading = true
		return m, m.doGitOp(func() error {
			return m.repo.StashPop()
		}, "Popped stash")

	// Undo
	case "z":
		m.loading = true
		return m, m.doGitOp(func() error {
			return m.repo.UndoLastCommit()
		}, "Undid last commit")

	// Refresh
	case "r":
		m.loading = true
		return m, m.refreshStatus()

	// Collapse/expand
	case "h", "left":
		m.toggleCollapse()
		return m, nil

	case "l", "right":
		m.toggleCollapse()
		return m, nil
	}

	return m, nil
}

func (m Model) handleCommitKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.commitInput.Blur()
		return m, nil

	case "ctrl+enter":
		// Commit
		message := m.commitInput.Value()
		if message == "" {
			m.errMsg = "Commit message cannot be empty"
			return m, m.clearMessage()
		}
		m.mode = ModeNormal
		m.commitInput.Blur()
		m.loading = true
		return m, m.doGitOp(func() error {
			return m.repo.Commit(message)
		}, "Committed: "+truncateMsg(message, 30))
	}

	// Pass to textarea
	var cmd tea.Cmd
	m.commitInput, cmd = m.commitInput.Update(msg)
	return m, cmd
}

func (m Model) handleDiffKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "d", "q":
		m.mode = ModeNormal
		m.currentDiff = nil
		return m, nil

	case "j", "down":
		if m.currentDiff != nil && m.diffScroll < len(m.currentDiff.Lines)-m.diffMaxLines {
			m.diffScroll++
		}
		return m, nil

	case "k", "up":
		if m.diffScroll > 0 {
			m.diffScroll--
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleBranchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "b":
		m.mode = ModeNormal
		return m, nil

	case "j", "down":
		if m.branchCursor < len(m.branches)-1 {
			m.branchCursor++
		}
		return m, nil

	case "k", "up":
		if m.branchCursor > 0 {
			m.branchCursor--
		}
		return m, nil

	case "enter":
		if m.branchCursor < len(m.branches) {
			branch := m.branches[m.branchCursor]
			m.mode = ModeNormal
			m.loading = true
			return m, m.doGitOp(func() error {
				return m.repo.CheckoutBranch(branch)
			}, "Switched to "+branch)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "?", "q":
		m.mode = ModeNormal
		return m, nil
	}
	return m, nil
}

func (m Model) handleStageToggle() (Model, tea.Cmd) {
	entries := m.focusedEntries()
	if m.cursor >= len(entries) || len(entries) == 0 {
		return m, nil
	}

	entry := entries[m.cursor]
	m.loading = true

	if m.focus == SectionStaged {
		return m, m.doGitOp(func() error {
			return m.repo.Unstage(entry.Path)
		}, "Unstaged "+entry.Path)
	}
	return m, m.doGitOp(func() error {
		return m.repo.Stage(entry.Path)
	}, "Staged "+entry.Path)
}

func (m Model) handleShowDiff() (Model, tea.Cmd) {
	entries := m.focusedEntries()
	if m.cursor >= len(entries) || len(entries) == 0 {
		return m, nil
	}

	entry := entries[m.cursor]
	staged := m.focus == SectionStaged
	return m, m.loadDiff(entry.Path, staged)
}

func (m *Model) cycleSection(forward bool) {
	sections := []Section{SectionStaged, SectionUnstaged, SectionStashes}
	for i, s := range sections {
		if s == m.focus {
			if forward {
				m.focus = sections[(i+1)%len(sections)]
			} else {
				m.focus = sections[(i-1+len(sections))%len(sections)]
			}
			m.cursor = 0
			return
		}
	}
}

func (m *Model) toggleCollapse() {
	switch m.focus {
	case SectionStaged:
		m.stagedCollapsed = !m.stagedCollapsed
	case SectionUnstaged:
		m.unstagedCollapsed = !m.unstagedCollapsed
	case SectionStashes:
		m.stashCollapsed = !m.stashCollapsed
	}
}

func truncateMsg(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
