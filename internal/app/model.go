package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ankityadav/zedgit/internal/git"
)

// Ensure context import is used
var _ = context.Background

// Section identifies which part of the UI has focus.
type Section int

const (
	SectionCommit Section = iota
	SectionStaged
	SectionUnstaged
	SectionStashes
	SectionGraph
	SectionDiff
)

// Mode represents the current app mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeCommit
	ModeDiff
	ModeBranch
	ModeHelp
)

// Model is the main application model for bubbletea.
type Model struct {
	repo   *git.Repo
	status *git.RepoStatus

	// UI state
	width  int
	height int
	mode   Mode
	focus  Section
	cursor int // cursor index within focused section

	// Section collapsed state
	stagedCollapsed   bool
	unstagedCollapsed bool
	stashCollapsed    bool
	graphCollapsed    bool

	// Commit input
	commitInput textarea.Model

	// Diff view
	currentDiff  *git.DiffResult
	diffScroll   int
	diffMaxLines int

	// Branch picker
	branches     []string
	branchCursor int

	// Commit graph
	logResult       *git.LogResult
	graphCursor     int
	graphScroll     int
	graphMaxVisible int

	// Mouse state
	layout   *LayoutMap
	hoverRow int
	hoverCol int

	// Ollama AI commit message generation
	ollamaCfg    git.OllamaConfig
	generating   bool // true while Ollama is generating

	// Feedback
	loading    bool
	errMsg     string
	successMsg string
	msgExpiry  time.Time
}

// Messages
type statusMsg struct{ status *git.RepoStatus }
type errMsg struct{ err error }
type successMsgType struct{ msg string }
type diffMsg struct{ diff *git.DiffResult }
type branchesMsg struct{ branches []string }
type logMsg struct{ log *git.LogResult }
type commitFilesMsg struct {
	index int
	files []git.CommitFile
}
type generateChunkMsg struct{ partial string }
type generateDoneMsg struct{ message string }
type clearMsgTick struct{}
type refreshTick struct{}
type gitOpDone struct{ msg string }

// New creates a new app model.
func New(path string) (Model, error) {
	repo, err := git.Open(path)
	if err != nil {
		return Model{}, err
	}

	ti := textarea.New()
	ti.Placeholder = "Commit message..."
	ti.CharLimit = 500
	ti.SetWidth(40)
	ti.SetHeight(3)
	ti.ShowLineNumbers = false

	m := Model{
		repo:            repo,
		mode:            ModeNormal,
		focus:           SectionUnstaged,
		commitInput:     ti,
		stashCollapsed:  true,
		graphCollapsed:  false,
		graphMaxVisible: 20,
		hoverRow:        -1,
		hoverCol:        -1,
		ollamaCfg:       git.DefaultOllamaConfig(),
	}

	return m, nil
}

// Init initializes the model and starts the first status refresh.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshStatus(),
		m.refreshLog(),
		m.tickRefresh(),
	)
}

func (m Model) refreshStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.repo.Status()
		if err != nil {
			return errMsg{err}
		}
		return statusMsg{status}
	}
}

func (m Model) refreshLog() tea.Cmd {
	return func() tea.Msg {
		log, err := m.repo.Log(50)
		if err != nil {
			return errMsg{err}
		}
		return logMsg{log}
	}
}

func (m Model) tickRefresh() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return refreshTick{}
	})
}

func (m Model) clearMessage() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearMsgTick{}
	})
}

func (m Model) loadDiff(path string, staged bool) tea.Cmd {
	return func() tea.Msg {
		diff, err := m.repo.Diff(path, staged)
		if err != nil {
			return errMsg{err}
		}
		return diffMsg{diff}
	}
}

func (m Model) loadBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.repo.Branches()
		if err != nil {
			return errMsg{err}
		}
		return branchesMsg{branches}
	}
}

func (m Model) doGitOp(op func() error, successText string) tea.Cmd {
	return func() tea.Msg {
		if err := op(); err != nil {
			return errMsg{err}
		}
		return gitOpDone{successText}
	}
}

// generateCommitMessage starts an Ollama generation (non-streaming for simplicity,
// since we don't have the tea.Program reference for p.Send).
func (m Model) generateCommitMessage() tea.Cmd {
	cfg := m.ollamaCfg
	repo := m.repo

	return func() tea.Msg {
		diff, err := repo.GetStagedDiff()
		if err != nil {
			return errMsg{err}
		}
		if strings.TrimSpace(diff) == "" {
			// Fall back to unstaged diff if nothing is staged
			diff, err = runGitDiff(repo)
			if err != nil {
				return errMsg{err}
			}
			if strings.TrimSpace(diff) == "" {
				return errMsg{fmt.Errorf("no changes to generate a message for")}
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		msg, err := git.GenerateCommitMessage(ctx, cfg, diff, nil)
		if err != nil {
			return errMsg{err}
		}

		return generateDoneMsg{message: msg}
	}
}

func runGitDiff(repo *git.Repo) (string, error) {
	return repo.GetStagedDiff()
}

// focusedEntries returns the entries for the currently focused section.
func (m Model) focusedEntries() []git.StatusEntry {
	if m.status == nil {
		return nil
	}
	switch m.focus {
	case SectionStaged:
		return m.status.Staged
	case SectionUnstaged:
		return m.status.Unstaged
	default:
		return nil
	}
}

// clampCursor ensures the cursor stays within bounds.
func (m *Model) clampCursor() {
	entries := m.focusedEntries()
	if m.cursor >= len(entries) {
		m.cursor = len(entries) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}
