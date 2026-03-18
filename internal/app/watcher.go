package app

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// fileChangedMsg is sent when the filesystem watcher detects changes.
type fileChangedMsg struct{}

// watchRepo starts a filesystem watcher on the repo and returns a tea.Cmd
// that fires fileChangedMsg when files change. Uses debouncing to avoid
// flooding during bulk operations (checkout, rebase, etc).
func watchRepo(repoPath string) tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil
		}

		// Watch the working directory (top-level files)
		watcher.Add(repoPath)

		// Watch .git directory for branch changes, commits from other tools, etc
		gitDir := filepath.Join(repoPath, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			watcher.Add(gitDir)
			// Watch refs for branch/remote changes
			refsDir := filepath.Join(gitDir, "refs")
			if info, err := os.Stat(refsDir); err == nil && info.IsDir() {
				watcher.Add(refsDir)
				filepath.Walk(refsDir, func(path string, info os.FileInfo, err error) error {
					if err == nil && info.IsDir() {
						watcher.Add(path)
					}
					return nil
				})
			}
		}

		// Also watch common subdirectories (1 level deep)
		entries, _ := os.ReadDir(repoPath)
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") && e.Name() != "node_modules" && e.Name() != "vendor" {
				watcher.Add(filepath.Join(repoPath, e.Name()))
			}
		}

		// Debounce: wait for 200ms of quiet before firing
		var debounceTimer *time.Timer

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				// Skip .git/index.lock (transient) and editor swap files
				name := filepath.Base(event.Name)
				if name == "index.lock" || strings.HasSuffix(name, ".swp") || strings.HasSuffix(name, "~") {
					continue
				}

				// Reset debounce timer
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.NewTimer(200 * time.Millisecond)

				// Wait for debounce
				<-debounceTimer.C
				return fileChangedMsg{}

			case _, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				// Ignore errors, just keep watching
			}
		}
	}
}
