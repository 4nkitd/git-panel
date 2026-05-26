# Contributing to git-panel

A terminal UI for Git, built with Bubble Tea. Designed to live next to AI coding assistants (Claude Code, opencode) in a split terminal.

## Dev setup

Prereqs:

- Go 1.25+ (the Makefile expects modern toolchain; older versions may work)
- Git
- A terminal with 256-color support
- (Optional) Ollama for AI commit messages

```bash
git clone https://github.com/4nkitd/git-panel.git
cd git-panel
make build
./git-panel
```

For a quick install:

```bash
go install github.com/4nkitd/git-panel/cmd/git-panel@latest
```

## Running tests

```bash
go test ./...
```

(Test coverage is light. TUI snapshot tests via `charmbracelet/x/exp/teatest` are welcome.)

## Project layout

```
git-panel/
├── cmd/git-panel/        # main entry point
├── internal/
│   ├── app/              # Bubble Tea model + update + view
│   ├── git/              # git wrapper (shell out to `git`)
│   ├── llm/              # Ollama client for AI commit messages
│   └── ...
├── assets/               # README screenshots
├── Makefile
├── .goreleaser.yml
└── README.md
```

## Adding a feature

1. Define the Bubble Tea command/message in `internal/app/`
2. If it shells out to git, add the wrapper in `internal/git/`
3. Wire up the keybinding in the relevant view
4. Document the keybinding in README.md
5. Add a CHANGELOG entry under `## [Unreleased]`

## Branches and commits

- Branch from `main`: `feat/<name>`, `fix/<name>`, `docs/<name>`.
- Conventional Commits encouraged.

## PR checklist

- [ ] `make build` succeeds
- [ ] `go vet ./...` clean
- [ ] `go test ./...` passes
- [ ] Manually exercised in a real git repo (state which terminal in the PR)
- [ ] README updated if behavior changed
- [ ] `CHANGELOG.md` entry under `## [Unreleased]`

## Releases

Maintainers only. Tag with `vX.Y.Z`, then run goreleaser locally (Actions disabled at the account level for now):

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
GITHUB_TOKEN=<token> goreleaser release --clean
```

## Reporting issues

[GitHub issues](https://github.com/4nkitd/git-panel/issues).
