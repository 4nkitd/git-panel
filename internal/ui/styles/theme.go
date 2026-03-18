package styles

import (
	"github.com/ankityadav/zedgit/internal/git"
	"github.com/charmbracelet/lipgloss"
)

// Colors matching VS Code's Source Control panel.
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
	BgSelected  = lipgloss.Color("#04395e")
	BgHover     = lipgloss.Color("#2a2d2e")
	BgHeader    = lipgloss.Color("#252526")
	BgDiffAdd   = lipgloss.Color("#1e3a21")
	BgDiffDel   = lipgloss.Color("#3a1e1e")
	BgInput     = lipgloss.Color("#3c3c3c")
	BgButton    = lipgloss.Color("#0e639c")
	Border      = lipgloss.Color("#404040")
)

// Graph branch colors (VS Code uses these for commit graph lines).
var GraphColors = []lipgloss.Color{
	lipgloss.Color("#4fc1ff"), // blue
	lipgloss.Color("#e5c07b"), // yellow/gold
	lipgloss.Color("#c678dd"), // magenta/purple
	lipgloss.Color("#98c379"), // green
	lipgloss.Color("#e06c75"), // red/pink
	lipgloss.Color("#56b6c2"), // cyan
	lipgloss.Color("#d19a66"), // orange
}

// Component styles.
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(BrightWhite).
			Padding(0, 1)

	SectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(White)

	SectionCountStyle = lipgloss.NewStyle().
				Foreground(DimWhite)

	SelectedStyle = lipgloss.NewStyle().
			Background(BgSelected)

	HoverStyle = lipgloss.NewStyle().
			Background(BgHover)

	FilePathStyle = lipgloss.NewStyle().
			Foreground(White)

	FileDirStyle = lipgloss.NewStyle().
			Foreground(DimWhite)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(DimWhite).
			Background(BgHeader)

	BranchStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	CommitInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Border).
				Padding(0, 1)

	CommitButtonStyle = lipgloss.NewStyle().
				Foreground(BrightWhite).
				Background(BgButton).
				Bold(true).
				Padding(0, 2)

	CommitSecondaryStyle = lipgloss.NewStyle().
				Foreground(White).
				Background(BgHeader).
				Padding(0, 1)

	// Per-file action icon styles (VS Code hover icons)
	StageIconStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	UnstageIconStyle = lipgloss.NewStyle().
				Foreground(Yellow).
				Bold(true)

	DiscardIconStyle = lipgloss.NewStyle().
				Foreground(Red)

	DiffAddStyle = lipgloss.NewStyle().
			Foreground(Green).
			Background(BgDiffAdd)

	DiffDelStyle = lipgloss.NewStyle().
			Foreground(Red).
			Background(BgDiffDel)

	DiffHunkStyle = lipgloss.NewStyle().
			Foreground(Cyan)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(Blue).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(DimWhite)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Blue)

	// Commit graph styles
	GraphCommitStyle = lipgloss.NewStyle().
				Foreground(White)

	GraphHashStyle = lipgloss.NewStyle().
			Foreground(Yellow)

	GraphAuthorStyle = lipgloss.NewStyle().
				Foreground(Blue)

	GraphDateStyle = lipgloss.NewStyle().
			Foreground(DimWhite)

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

// StatusStyle returns the style for a given file status.
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
