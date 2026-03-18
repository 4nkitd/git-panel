package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ankityadav/zedgit/internal/git"
)

// Update handles all messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// textarea width = panel - margin(1) - border(2) - padding(2) = width - 7
		inputW := m.width - 7
		if inputW < 20 {
			inputW = 20
		}
		m.commitInput.SetWidth(inputW)
		m.diffMaxLines = m.height / 3
		return m, nil

	case statusMsg:
		m.status = msg.status
		if !m.generating {
			m.stopLoading()
		}
		m.clampCursor()
		return m, nil

	case errMsg:
		m.stopLoading()
		m.generating = false
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

	case logMsg:
		// Preserve expanded state from old log when refreshing
		if m.logResult != nil && msg.log != nil {
			oldExpanded := make(map[string][]git.CommitFile)
			for _, c := range m.logResult.Commits {
				if c.Expanded {
					oldExpanded[c.Hash] = c.Files
				}
			}
			for i := range msg.log.Commits {
				if files, ok := oldExpanded[msg.log.Commits[i].Hash]; ok {
					msg.log.Commits[i].Expanded = true
					msg.log.Commits[i].Files = files
				}
			}
		}
		m.logResult = msg.log
		return m, nil

	case commitFilesMsg:
		if msg.index >= 0 && msg.index < len(m.logResult.Commits) {
			m.logResult.Commits[msg.index].Expanded = true
			m.logResult.Commits[msg.index].Files = msg.files
		}
		return m, nil

	case generateChunkMsg:
		m.generating = true
		m.commitInput.SetValue(msg.partial)
		return m, nil

	case generateDoneMsg:
		m.stopLoading()
		m.generating = false
		m.mode = ModeCommit
		m.commitInput.SetValue(msg.message)
		m.commitInput.Focus()
		m.successMsg = "AI message generated"
		return m, tea.Batch(m.commitInput.Focus(), m.clearMessage())

	case gitOpDone:
		m.stopLoading()
		m.successMsg = msg.msg
		return m, tea.Batch(m.refreshStatus(), m.refreshLog(), m.clearMessage())

	case ollamaModelsMsg:
		m.availableModels = msg.models
		m.modelListLoaded = true
		return m, nil

	case fileChangedMsg:
		// Real-time: filesystem change detected, refresh and re-watch
		return m, tea.Batch(m.refreshStatus(), watchRepo(m.repo.Path))

	case spinnerTickMsg:
		if m.spinner.active {
			m.spinner.Tick()
			return m, spinnerTick()
		}
		return m, nil

	case clearMsgTick:
		m.errMsg = ""
		m.successMsg = ""
		return m, nil

	case refreshTick:
		return m, tea.Batch(m.refreshStatus(), m.tickRefresh())

	case tea.MouseMsg:
		return m.handleMouse(msg)

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

// ── Mouse handling ──

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Grab the layout map built during last View()
	lm := lastLayout
	if lm == nil {
		return m, nil
	}

	switch msg.Action {
	case tea.MouseActionMotion:
		// Track hover position
		m.hoverRow = msg.Y
		m.hoverCol = msg.X
		return m, nil

	case tea.MouseActionPress:
		switch msg.Button {
		case tea.MouseButtonLeft:
			return m.handleMouseClick(msg.X, msg.Y, lm)

		case tea.MouseButtonWheelUp:
			return m.handleMouseScroll(-3, msg.Y, lm)

		case tea.MouseButtonWheelDown:
			return m.handleMouseScroll(3, msg.Y, lm)
		}

	case tea.MouseActionRelease:
		// Nothing special on release
	}

	return m, nil
}

func (m Model) handleMouseClick(x, y int, lm *LayoutMap) (Model, tea.Cmd) {
	hit := lm.Get(y)

	switch hit.Zone {
	case ZoneCommitInput:
		// Click on commit input → enter commit mode
		if m.mode != ModeCommit {
			m.mode = ModeCommit
			m.commitInput.Focus()
			return m, m.commitInput.Focus()
		}
		return m, nil

	case ZoneCommitButton:
		// Click commit button
		if m.mode == ModeCommit {
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
		return m, nil

	case ZoneStagedHeader:
		// Click on section header → toggle collapse
		m.stagedCollapsed = !m.stagedCollapsed
		m.focus = SectionStaged
		return m, nil

	case ZoneUnstagedHeader:
		m.unstagedCollapsed = !m.unstagedCollapsed
		m.focus = SectionUnstaged
		return m, nil

	case ZoneStashHeader:
		m.stashCollapsed = !m.stashCollapsed
		m.focus = SectionStashes
		return m, nil

	case ZoneGraphHeader:
		m.graphCollapsed = !m.graphCollapsed
		m.focus = SectionGraph
		return m, nil

	case ZoneStagedFile:
		m.focus = SectionStaged
		if hit.FileIndex < 0 {
			return m, nil
		}
		// If clicking the same file that's already selected, toggle (unstage)
		// Or if clicking the action icon area, always toggle
		alreadySelected := m.cursor == hit.FileIndex
		onActionIcon := hit.ColAction >= 0 && x >= hit.ColAction
		m.cursor = hit.FileIndex

		if (alreadySelected || onActionIcon) && m.status != nil && hit.FileIndex < len(m.status.Staged) {
			entry := m.status.Staged[hit.FileIndex]
			spin := m.startLoading("Unstaging...")
			return m, tea.Batch(spin, m.doGitOp(func() error {
				return m.repo.Unstage(entry.Path)
			}, "Unstaged "+entry.Path))
		}
		return m, nil

	case ZoneUnstagedFile:
		m.focus = SectionUnstaged
		if hit.FileIndex < 0 {
			return m, nil
		}
		alreadySelected := m.cursor == hit.FileIndex
		onActionIcon := hit.ColAction >= 0 && x >= hit.ColAction
		m.cursor = hit.FileIndex

		if (alreadySelected || onActionIcon) && m.status != nil && hit.FileIndex < len(m.status.Unstaged) {
			entry := m.status.Unstaged[hit.FileIndex]
			spin := m.startLoading("Staging...")
			return m, tea.Batch(spin, m.doGitOp(func() error {
				return m.repo.Stage(entry.Path)
			}, "Staged "+entry.Path))
		}
		return m, nil

	case ZoneGraphCommit:
		m.focus = SectionGraph
		m.graphCursor = hit.FileIndex
		// Click on a commit → toggle expand/collapse to show files
		return m.toggleCommitExpand(hit.FileIndex)

	case ZoneGraphFile:
		// Click on a file in expanded commit — could open diff in future
		return m, nil

	case ZoneStatusBar:
		// Click on status bar → could open branch picker
		return m, m.loadBranches()
	}

	return m, nil
}

func (m Model) handleMouseScroll(delta int, y int, lm *LayoutMap) (Model, tea.Cmd) {
	hit := lm.Get(y)

	switch hit.Zone {
	case ZoneDiff:
		// Scroll diff view
		m.diffScroll += delta
		if m.diffScroll < 0 {
			m.diffScroll = 0
		}
		if m.currentDiff != nil && m.diffScroll > len(m.currentDiff.Lines)-m.diffMaxLines {
			m.diffScroll = len(m.currentDiff.Lines) - m.diffMaxLines
			if m.diffScroll < 0 {
				m.diffScroll = 0
			}
		}
		return m, nil

	case ZoneGraphCommit, ZoneGraphFile, ZoneGraphHeader:
		// Scroll graph section
		m.graphScroll += delta
		if m.graphScroll < 0 {
			m.graphScroll = 0
		}
		maxScroll := 0
		if m.logResult != nil {
			maxScroll = len(m.logResult.Commits) - m.graphMaxVisible
		}
		if m.graphScroll > maxScroll {
			m.graphScroll = maxScroll
		}
		if m.graphScroll < 0 {
			m.graphScroll = 0
		}
		return m, nil

	default:
		// Scroll the focused section's file list
		switch m.focus {
		case SectionStaged, SectionUnstaged:
			m.cursor += delta
			m.clampCursor()
		case SectionGraph:
			m.graphCursor += delta
			if m.graphCursor < 0 {
				m.graphCursor = 0
			}
			if m.logResult != nil && m.graphCursor >= len(m.logResult.Commits) {
				m.graphCursor = len(m.logResult.Commits) - 1
			}
		}
		return m, nil
	}
}

// ── Keyboard handling ──

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	switch m.mode {
	case ModeCommit:
		return m.handleCommitKey(msg)
	case ModeDiff:
		return m.handleDiffKey(msg)
	case ModeBranch:
		return m.handleBranchKey(msg)
	case ModeHelp:
		return m.handleHelpKey(msg)
	case ModeSettings:
		return m.handleSettingsKey(msg)
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
		if m.focus == SectionGraph {
			if m.logResult != nil && m.graphCursor < len(m.logResult.Commits)-1 {
				m.graphCursor++
			}
		} else {
			m.cursor++
			m.clampCursor()
		}
		return m, nil

	case "k", "up":
		if m.focus == SectionGraph {
			if m.graphCursor > 0 {
				m.graphCursor--
			}
		} else {
			m.cursor--
			m.clampCursor()
		}
		return m, nil

	case "tab":
		m.cycleSection(true)
		return m, nil

	case "shift+tab":
		m.cycleSection(false)
		return m, nil

	// Stage/unstage
	case "enter", " ":
		if m.focus == SectionGraph {
			return m.toggleCommitExpand(m.graphCursor)
		}
		return m.handleStageToggle()

	case "a":
		spin := m.startLoading("Staging all...")
		return m, tea.Batch(spin, m.doGitOp(func() error {
			return m.repo.StageAll()
		}, "Staged all changes"))

	case "A":
		spin := m.startLoading("Unstaging all...")
		return m, tea.Batch(spin, m.doGitOp(func() error {
			return m.repo.UnstageAll()
		}, "Unstaged all changes"))

	// Commit
	case "c":
		m.mode = ModeCommit
		m.commitInput.Focus()
		return m, m.commitInput.Focus()

	// Generate commit message with Ollama AI
	case "g":
		if m.generating {
			return m, nil
		}
		m.generating = true
		m.mode = ModeCommit
		m.commitInput.SetValue("")
		m.commitInput.Focus()
		spin := m.startLoading("Generating commit message...")
		return m, tea.Batch(spin, m.commitInput.Focus(), m.generateCommitMessage())

	// Diff
	case "d":
		return m.handleShowDiff()

	// Branch
	case "b":
		return m, m.loadBranches()

	// Remote ops
	case "p":
		spin := m.startLoading("Pushing to remote...")
		return m, tea.Batch(spin, m.doGitOp(func() error {
			return m.repo.Push()
		}, "Pushed successfully"))

	case "P":
		spin := m.startLoading("Pulling from remote...")
		return m, tea.Batch(spin, m.doGitOp(func() error {
			return m.repo.Pull()
		}, "Pulled successfully"))

	case "f":
		spin := m.startLoading("Fetching from remote...")
		return m, tea.Batch(spin, m.doGitOp(func() error {
			return m.repo.Fetch()
		}, "Fetched successfully"))

	// Stash
	case "s":
		spin := m.startLoading("Stashing changes...")
		return m, tea.Batch(spin, m.doGitOp(func() error {
			return m.repo.Stash("")
		}, "Stashed changes"))

	case "S":
		spin := m.startLoading("Popping stash...")
		return m, tea.Batch(spin, m.doGitOp(func() error {
			return m.repo.StashPop()
		}, "Popped stash"))

	// Undo
	case "z":
		spin := m.startLoading("Undoing last commit...")
		return m, tea.Batch(spin, m.doGitOp(func() error {
			return m.repo.UndoLastCommit()
		}, "Undid last commit"))

	// Refresh
	case "r":
		spin := m.startLoading("Refreshing...")
		return m, tea.Batch(spin, m.refreshStatus(), m.refreshLog())

	// Settings
	case ",":
		m.mode = ModeSettings
		m.settingsCursor = 0
		if !m.modelListLoaded {
			return m, m.fetchOllamaModels()
		}
		return m, nil

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
		message := m.commitInput.Value()
		if message == "" {
			m.errMsg = "Commit message cannot be empty"
			return m, m.clearMessage()
		}
		m.mode = ModeNormal
		m.commitInput.Blur()
		m.commitInput.Reset()
		spin := m.startLoading("Committing...")
		return m, tea.Batch(spin, m.doGitOp(func() error {
			return m.repo.Commit(message)
		}, "Committed: "+truncateMsg(message, 30)))
	}

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

func (m Model) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", ",", "q":
		m.mode = ModeNormal
		return m, nil
	case "j", "down":
		if m.settingsCursor < len(m.availableModels)-1 {
			m.settingsCursor++
		}
		return m, nil
	case "k", "up":
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
		return m, nil
	case "enter":
		// Select the model
		if m.settingsCursor < len(m.availableModels) {
			m.ollamaCfg.Model = m.availableModels[m.settingsCursor]
			m.successMsg = "Model set to " + m.ollamaCfg.Model
			m.mode = ModeNormal
			return m, m.clearMessage()
		}
		return m, nil
	case "r":
		// Refresh model list
		m.modelListLoaded = false
		return m, m.fetchOllamaModels()
	}
	return m, nil
}

func (m Model) fetchOllamaModels() tea.Cmd {
	cfg := m.ollamaCfg
	return func() tea.Msg {
		models, err := git.ListOllamaModels(cfg)
		if err != nil {
			return errMsg{err}
		}
		return ollamaModelsMsg{models}
	}
}

func (m Model) handleStageToggle() (Model, tea.Cmd) {
	entries := m.focusedEntries()
	if m.cursor >= len(entries) || len(entries) == 0 {
		return m, nil
	}

	entry := entries[m.cursor]

	if m.focus == SectionStaged {
		spin := m.startLoading("Unstaging...")
		return m, tea.Batch(spin, m.doGitOp(func() error {
			return m.repo.Unstage(entry.Path)
		}, "Unstaged "+entry.Path))
	}
	spin := m.startLoading("Staging...")
	return m, tea.Batch(spin, m.doGitOp(func() error {
		return m.repo.Stage(entry.Path)
	}, "Staged "+entry.Path))
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

func (m Model) toggleCommitExpand(idx int) (Model, tea.Cmd) {
	if m.logResult == nil || idx < 0 || idx >= len(m.logResult.Commits) {
		return m, nil
	}

	commit := &m.logResult.Commits[idx]
	if commit.Expanded {
		commit.Expanded = false
		commit.Files = nil
		return m, nil
	}

	// Load files for this commit
	hash := commit.Hash
	return m, func() tea.Msg {
		files, err := m.repo.CommitFiles(hash)
		if err != nil {
			return errMsg{err}
		}
		return commitFilesMsg{index: idx, files: files}
	}
}

func (m *Model) cycleSection(forward bool) {
	sections := []Section{SectionStaged, SectionUnstaged, SectionStashes, SectionGraph}
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
	case SectionGraph:
		m.graphCollapsed = !m.graphCollapsed
	}
}

func truncateMsg(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
