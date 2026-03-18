package styles

import (
	"github.com/ankityadav/zedgit/internal/git"
	"github.com/charmbracelet/lipgloss"
)

// Colors — VS Code dark theme palette.
var (
	Green       = lipgloss.Color("#73c991")
	Yellow      = lipgloss.Color("#cca700")
	Red         = lipgloss.Color("#f44747")
	Blue        = lipgloss.Color("#6796e6")
	Cyan        = lipgloss.Color("#4ec9b0")
	Magenta     = lipgloss.Color("#c586c0")
	Orange      = lipgloss.Color("#ce9178")
	Gray        = lipgloss.Color("#858585")
	White       = lipgloss.Color("#cccccc")
	BrightWhite = lipgloss.Color("#ffffff")
	DimWhite    = lipgloss.Color("#969696")
	Subtle      = lipgloss.Color("#555555")
	BgSelected  = lipgloss.Color("#04395e")
	BgHover     = lipgloss.Color("#2a2d2e")
	BgHeader    = lipgloss.Color("#1e1e1e")
	BgPanel     = lipgloss.Color("#181818")
	BgDiffAdd   = lipgloss.Color("#1e3a21")
	BgDiffDel   = lipgloss.Color("#3a1e1e")
	BgInput     = lipgloss.Color("#313131")
	BgButton    = lipgloss.Color("#0e639c")
	BgGenBtn    = lipgloss.Color("#6c3483")
	Border      = lipgloss.Color("#333333")
	BorderFocus = lipgloss.Color("#007acc")
)

// Graph branch colors.
var GraphColors = []lipgloss.Color{
	lipgloss.Color("#4fc1ff"),
	lipgloss.Color("#e5c07b"),
	lipgloss.Color("#c678dd"),
	lipgloss.Color("#98c379"),
	lipgloss.Color("#e06c75"),
	lipgloss.Color("#56b6c2"),
	lipgloss.Color("#d19a66"),
}

// ── Component styles ──

var (
	// Title bar
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(DimWhite).
			Padding(0, 1)

	TitleTextStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(BrightWhite)

	// Section headers
	SectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(White)

	SectionCountStyle = lipgloss.NewStyle().
				Foreground(DimWhite).
				Italic(true)

	SectionSepStyle = lipgloss.NewStyle().
			Foreground(Subtle)

	// Selection / hover
	SelectedStyle = lipgloss.NewStyle().
			Background(BgSelected)

	HoverStyle = lipgloss.NewStyle().
			Background(BgHover)

	// File display
	FileNameStyle = lipgloss.NewStyle().
			Foreground(White)

	FileDirStyle = lipgloss.NewStyle().
			Foreground(DimWhite)

	FilePathStyle = lipgloss.NewStyle().
			Foreground(White)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(DimWhite).
			Background(BgHeader)

	StatusBarSepStyle = lipgloss.NewStyle().
				Foreground(Subtle).
				Background(BgHeader)

	BranchStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	// Commit input
	CommitInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Border).
				Padding(0, 1)

	CommitInputFocusStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(BorderFocus).
				Padding(0, 1)

	CommitButtonStyle = lipgloss.NewStyle().
				Foreground(BrightWhite).
				Background(BgButton).
				Bold(true).
				Padding(0, 2)

	CommitSecondaryStyle = lipgloss.NewStyle().
				Foreground(DimWhite).
				Background(BgHover).
				Padding(0, 1)

	GenerateButtonStyle = lipgloss.NewStyle().
				Foreground(BrightWhite).
				Background(BgGenBtn).
				Bold(true).
				Padding(0, 1)

	// Per-file action icons
	StageIconStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	UnstageIconStyle = lipgloss.NewStyle().
				Foreground(Yellow).
				Bold(true)

	DiscardIconStyle = lipgloss.NewStyle().
				Foreground(Red)

	// Diff
	DiffAddStyle = lipgloss.NewStyle().
			Foreground(Green).
			Background(BgDiffAdd)

	DiffDelStyle = lipgloss.NewStyle().
			Foreground(Red).
			Background(BgDiffDel)

	DiffHunkStyle = lipgloss.NewStyle().
			Foreground(Cyan)

	DiffBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border)

	// Help
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(Blue).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(DimWhite)

	HelpBarStyle = lipgloss.NewStyle().
			Foreground(Subtle)

	HelpBarKeyStyle = lipgloss.NewStyle().
			Foreground(DimWhite)

	// Feedback
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Blue)

	// Graph
	GraphCommitStyle = lipgloss.NewStyle().
				Foreground(White)

	GraphHashStyle = lipgloss.NewStyle().
			Foreground(Yellow)

	GraphAuthorStyle = lipgloss.NewStyle().
				Foreground(Blue)

	GraphDateStyle = lipgloss.NewStyle().
			Foreground(Subtle)

	GraphTagStyle = lipgloss.NewStyle().
			Foreground(Yellow).
			Bold(true)

	GraphBranchRefStyle = lipgloss.NewStyle().
				Foreground(Green).
				Bold(true)

	GraphRemoteRefStyle = lipgloss.NewStyle().
				Foreground(Red).
				Bold(true)
)

// StatusStyle returns the colored style for a git file status.
func StatusStyle(status git.FileStatus) lipgloss.Style {
	switch status {
	case git.StatusAdded:
		return lipgloss.NewStyle().Foreground(Green).Bold(true)
	case git.StatusModified:
		return lipgloss.NewStyle().Foreground(Yellow).Bold(true)
	case git.StatusDeleted:
		return lipgloss.NewStyle().Foreground(Red).Bold(true)
	case git.StatusRenamed:
		return lipgloss.NewStyle().Foreground(Blue).Bold(true)
	case git.StatusUntracked:
		return lipgloss.NewStyle().Foreground(Gray)
	case git.StatusConflicted:
		return lipgloss.NewStyle().Foreground(Red).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(White)
	}
}
