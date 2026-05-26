# Changelog

All notable changes to this project are documented here.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `CHANGELOG.md` and `CONTRIBUTING.md`

### Fixed
- Renamed `cmd/zedgit/` to `cmd/git-panel/` (zedgit was the project's previous name); updated Makefile, .goreleaser.yml, and README install command to match
- Anchored the `.gitignore` rule for the binary `git-panel` to `/git-panel` so it doesn't accidentally exclude files inside `cmd/git-panel/`
- Updated `.goreleaser.yml` to use the current `archives.formats` syntax (the legacy `format` key is deprecated)

---

## [v0.0.1] — initial

### Added
- TUI for Git: stage/unstage, commit, diff, history, branch switching, push/pull/fetch, stash
- AI commit message generation via Ollama (local LLM, press `g`)
- Mouse support — click to stage, scroll to navigate
- Real-time file watching (UI updates when files change)
- Diff viewer with syntax-highlighted hunks
- Commit graph with expandable file lists
- Settings UI (press `,`)
- Help screen (press `?`)
- Built with Bubble Tea / Lipgloss / Charm libraries
- Makefile, goreleaser config, MIT license
- README with usage, keybindings, screenshots, and split-terminal demos with Claude Code / opencode

[Unreleased]: https://github.com/4nkitd/git-panel/compare/v0.0.1...HEAD
[v0.0.1]: https://github.com/4nkitd/git-panel/releases/tag/v0.0.1
