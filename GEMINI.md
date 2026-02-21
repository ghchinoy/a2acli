# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Coding Standards and Patterns

- **Project Structure**: Go CLI entry points should be placed in `cmd/<app-name>/` (e.g., `cmd/a2acli/main.go`).
- **Build System & Quality Gates**: Use a `Makefile` containing `build`, `run`, `lint`, `test-e2e`, and `clean` targets. `make lint` MUST be run using `golangci-lint` to enforce Google Go code standards before concluding any task. Binaries should be output to the `bin/` directory and ignored in version control.
- **UI Framework**: Use the [Charm](https://charm.sh) ecosystem (Bubble Tea, Lipgloss, Bubbles) for all Terminal User Interfaces (TUIs).
- **Configuration**: Support both local `.env` files and the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) (e.g., `~/.config/a2acli/config.yaml`) for application configuration.
- **Task Management**: Use `bd todo add "<description>"` for quickly capturing tasks and feature requests as they arise during development.
- **Documentation**: Maintain a clear `README.md` containing installation instructions, global flags, and detailed command usage examples.
- **Cross-Repo Context**: The `a2acli` tests depend on the `a2a-go` SDK repository ([https://github.com/a2aproject/a2a-go](https://github.com/a2aproject/a2a-go)) being checked out locally. By default, it assumes the path is `../github/a2a-go`. If it is located elsewhere, `A2A_GO_SRC` must be set. When debugging e2e tests or core protocol behavior, investigate the SDK's TCK source code located in `e2e/tck/sut.go`.

## Testing & QA

- **Conformance Testing (TCK):** Always verify core CLI logic by running `make test-e2e`. This automatically builds the binary and tests it against the local `a2a-go` SDK System Under Test (SUT) server. 
- **Non-Interactive Modes:** When adding new features or outputs, ensure they gracefully bypass Bubble Tea (`--no-tui` or `A2ACLI_NO_TUI=true`) and emit parseable JSON/NDJSON to support the automated e2e tests.
- **Go Tests:** Avoid `bats-core` or external bash scripting frameworks. Rely entirely on the standard Go `testing` package combined with `os/exec` for invoking compiled binaries.

## Releasing & Publishing

- **GoReleaser**: This project uses GoReleaser via GitHub Actions. **Never manually publish binary artifacts or edit `version.go`.** 
- **Triggering a Release**: To release a new version, simply push a semantic version tag (e.g., `git tag v1.0.0 && git push origin v1.0.0`). The CI/CD pipeline will automatically inject linker flags into the binary and publish the archives. Refer to `docs/RELEASING.md` for full details.

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds