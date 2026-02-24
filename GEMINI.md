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

- **Project Structure**: Go CLI entry points should be placed in `cmd/<app-name>/` (e.g., `cmd/a2acli/main.go`). Logic for help and styling should be separated (e.g., `help.go`, `style.go`).
- **Build System & Quality Gates**: Use a `Makefile` containing `build`, `run`, `lint`, `test-e2e`, `install`, and `clean` targets. Run `make help` to see available targets. `make lint` MUST be run using `golangci-lint` before concluding any task.
- **UI Framework & Design**: 
  - Use the [Charm](https://charm.sh) ecosystem (Bubble Tea, Lipgloss, Bubbles) for TUIs.
  - Adhere to the principles in `docs/CLI_DESIGN_BEST_PRACTICES.md` (Tufte-inspired, high data-ink ratio).
  - Use the centralized semantic tokens in `cmd/a2acli/style.go` (Ayu-theme aligned) for all UI output to ensure consistency across light/dark themes.
- **A2A Protocol Alignment**: Command groups should reflect A2A concepts: **Discovery & Identity**, **Messaging & Tasks**, and **Client Configuration**.
- **Self-Documentation**: Every Cobra command MUST include `Long` and `Example` fields. Use `rootCmd.SetHelpFunc(colorizedHelpFunc)` to maintain the styled help hierarchy.
- **Error Handling**: Use the `fatalf(format, err, hint)` helper to provide proactive "Hints" for common failure states (e.g., missing config or server connectivity issues).
- **Transport Selection**: The CLI supports dynamic transport negotiation (gRPC > JSON-RPC > HTTP+JSON). Prefer high-performance gRPC when advertised by the agent's `AgentCard`.
- **Configuration**: Support both local `.env` files and the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) (e.g., `~/.config/a2acli/config.yaml`).
- **Task Management**: Use `bd todo add "<description>"` for quickly capturing tasks and feature requests as they arise during development.
- **Documentation**: Maintain a clear `README.md` containing installation instructions, global flags, and detailed command usage examples.
- **Cobra Commands**: Due to strict `revive` linting rules, always use the blank identifier `_` for unused parameters in Cobra command handlers.
- **Cross-Repo Context**: The `a2acli` tests depend on the `a2a-go` SDK repository. By default, it assumes `../github/a2a-go`. Override via `A2A_GO_SRC`. 
  - **Zombie Process Hygiene**: When debugging e2e tests, investigate the SDK's TCK source code in `e2e/tck/sut.go`. Always verify if there are lingering SUT processes (`pgrep -f sut.go`) from previous runs and terminate them to ensure a clean state.
  - **Branch Alignment**: Explicitly verify that the SDK repository is on the branch corresponding to the target specification version (e.g., `release/spec-v1` for Spec 1.0) before proceeding with conformance analysis.

## Protocol Versioning & Tooling

- **Protobuf Conflicts**: When importing multiple versions of the A2A SDK (e.g., v0 and v1), always ensure `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore` is set in the environment or via `os.Setenv` in `init()` to prevent panics during proto registration.
- **SDK Path Verification**: Before running conformance tests, verify that `A2A_GO_SRC` points to a valid SDK source. If the default in the `Makefile` is incorrect, override it via environment variables.
- **Linter Compatibility**: If `golangci-lint` fails due to a version mismatch in the config file, prioritize the local environment's version by removing the `version` field from `.golangci.yml` temporarily rather than downgrading the entire project's standards.

## Testing & QA

- **Conformance Testing (TCK):** Always verify core CLI logic by running `make test-e2e`. This automatically builds the binary and tests it against the local `a2a-go` SDK System Under Test (SUT) server. 
- **Conformance Report:** Use `make conformance-report` to generate or update `docs/CONFORMANCE_REPORT.md` before significant releases. This target captures the full TCK output and environment context.
- **Non-Interactive Modes:** When adding new features or outputs, ensure they gracefully bypass Bubble Tea (`--no-tui` or `A2ACLI_NO_TUI=true`) and emit parseable JSON/NDJSON to support the automated e2e tests.
- **Go Tests:** Avoid `bats-core` or external bash scripting frameworks. Rely entirely on the standard Go `testing` package combined with `os/exec` for invoking compiled binaries.
- **Code Formatting:** Rely exclusively on `go fmt ./...`. Do not assume `goimports` is installed in the local environment. If `golangci-lint` fails due to unused imports, remove them manually.

## Releasing & Publishing

- **GoReleaser**: This project uses GoReleaser via GitHub Actions. **Never manually publish binary artifacts or edit `version.go`.** 
- **Pre-Release Verification**: Before tagging a new release, always execute `make conformance-report` to verify Spec v1.0 compliance. This ensures that `docs/CONFORMANCE_REPORT.md` is updated with the latest TCK results and environment context (SDK branch/commit), providing a clear audit trail for the release.
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
   git add .beads/
   git commit -m "chore(bd): sync tasks"
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