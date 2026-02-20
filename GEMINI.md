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
- **Build System**: Use a `Makefile` containing at least `build`, `run`, and `clean` targets. Binaries should be output to the `bin/` directory and ignored in version control.
- **UI Framework**: Use the [Charm](https://charm.sh) ecosystem (Bubble Tea, Lipgloss, Bubbles) for all Terminal User Interfaces (TUIs).
- **Configuration**: Support both local `.env` files and the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) (e.g., `~/.config/a2acli/config.yaml`) for application configuration.
- **Task Management**: Use `bd todo add "<description>"` for quickly capturing tasks and feature requests as they arise during development.
- **Documentation**: Maintain a clear `README.md` containing installation instructions, global flags, and detailed command usage examples.

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
