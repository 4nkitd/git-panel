package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/4nkitd/git-panel/internal/app"
)

var version = "dev"

func main() {
	path := flag.String("path", ".", "Path to git repository")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("git-panel %s\n", version)
		os.Exit(0)
	}

	// Also accept positional arg
	repoPath := *path
	if flag.NArg() > 0 {
		repoPath = flag.Arg(0)
	}

	m, err := app.New(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
