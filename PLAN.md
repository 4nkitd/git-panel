# PLAN — git-panel

> Git TUI designed to live next to Claude Code / opencode in a split terminal. Strong narrative.

## Status snapshot

- **Stack**: Go, Bubble Tea (likely), Charm libs
- **Distribution**: `go install`, source build, release binaries
- **AI features**: Ollama-powered commit message generation (press `g`)
- Has clean README with screenshots + keybindings

## Strengths

- Excellent positioning: "TUI Git for the AI-coding-era developer"
- Screenshots showing split-terminal use with Claude Code + OpenCode = killer demo
- Already differentiated from `lazygit` / `gh dash` via AI commit + opencode framing
- README is one of the strongest in your portfolio

## Gaps

- Install command is wrong: `go install github.com/4nkitd/git-panel/cmd/zedgit@latest` — leftover `zedgit` name. Should be `cmd/git-panel`.
- No GitHub Releases / Homebrew (README says "Download from Releases" but they may not exist yet)
- No tests
- No CI
- AI commit message tied to Ollama only — should use `sapiens` for multi-provider
- Mouse support claimed in README — verify across terminals (kitty, alacritty, iTerm2, gnome-terminal, wezterm)
- No screencast / asciinema in README (just static screenshots — TUI deserves motion)

## Plan

### Documentation
- [ ] Fix install command (drop `zedgit`)
- [ ] Add asciinema cast embedded in README showing stage→commit→push flow
- [ ] Add `docs/keybindings.md` (or in-app `?` help mirrored)
- [ ] Add comparison with `lazygit`, `gitui`, `tig`, `gh dash` in a `docs/comparison.md`
- [ ] Document `--ai-provider` config once it's multi-provider

### Testing
- [ ] Unit tests for git wrapper (mock `exec.Command` via interface)
- [ ] Tests for diff parsing and hunk rendering
- [ ] Snapshot tests for TUI views (teatest from charmbracelet)
- [ ] Mouse event handling tests

### CI/CD

_Deferred: GitHub Actions is disabled at the account level. Revisit when re-enabled. Until then, run lint/test/build locally before pushing._

### Releases
- [ ] Tag v0.1.0 with binaries
- [ ] Homebrew: `brew install 4nkitd/tap/git-panel`
- [ ] Linux package managers (AUR, deb via nfpm)
- [ ] `CHANGELOG.md`

### Code quality
- [ ] Replace direct `ollama` calls with `sapiens` → support Gemini / OpenAI / Anthropic for AI commits
- [ ] Persistent settings at `~/.config/git-panel/config.yaml` (already has settings UI by `,` key — formalize storage)
- [ ] Performance: large-repo testing (>10k files; check `git status` parallelization)
- [ ] Handle detached-HEAD state gracefully in branch view
- [ ] Plugin / hook system for custom keybindings

### Roadmap
- **v0.2**: Multi-provider AI commits via sapiens; better diff highlighting (treesitter)
- **v0.2**: Interactive rebase view
- **v0.3**: PR creation flow (`gh` integration) — create PR from TUI after push
- **v0.3**: Conflict resolution helper
- **v0.4**: Project-aware workflows (detect `.opencode/` or `.claude/` and surface relevant hints)
- **v0.4**: Worktree manager
- **v0.5**: TUI extension API — let users add panels

## Decisions to make

1. **AI providers**: Ollama-only (privacy story) or multi-provider via sapiens (broader user base)? Recommend multi-provider with Ollama as default.
2. **Naming**: `git-panel` is fine; the `zedgit` leftover suggests it had a previous name. Pick `git-panel` and purge `zedgit` everywhere.

## Milestones

| Timeline | Outcome |
|---|---|
| Week 1 | `zedgit` references removed; v0.1.0 released; Homebrew tap; asciinema demo |
| Week 2 | local checks green; multi-provider via sapiens (v0.2 prep) |
| Week 4 | v0.2 with interactive rebase + better diffs |
| Quarter | 100 ⭐, mentioned alongside lazygit in "best git TUI" lists |
