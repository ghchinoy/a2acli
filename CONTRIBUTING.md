# Contributing to a2acli

Contributions are welcome. Before starting work on a new feature or significant change, please open an issue to discuss your approach. This prevents duplicate effort and ensures alignment with the project's goals.

## Setup

```bash
git clone https://github.com/ghchinoy/a2acli.git
cd a2acli
go mod tidy
make build
```

You'll need [Go 1.25+](https://go.dev/) installed.

## Development workflow

```bash
make build      # Compile to bin/a2acli
make lint       # Run golangci-lint (fix all warnings before submitting a PR)
make test-e2e   # Run end-to-end conformance tests (see below)
make clean      # Remove build artifacts
```

## Conformance tests

The end-to-end test suite spins up the official A2A TCK SUT server and runs black-box tests against the compiled CLI. It requires the [`a2a-go`](https://github.com/a2aproject/a2a-go) SDK source locally:

```bash
git clone https://github.com/a2aproject/a2a-go.git /path/to/a2a-go
make test-e2e A2A_GO_SRC=/path/to/a2a-go
```

All conformance tests must pass before a PR will be merged.

## Design conventions

Before adding or modifying commands, read [docs/CLI_DESIGN_BEST_PRACTICES.md](docs/CLI_DESIGN_BEST_PRACTICES.md). Key rules:

- Every command must have a `Short`, `Long`, and `Example` string
- All data-producing commands must support `-n` / `--no-tui` for JSON output
- Failures must include a `Hint:` line pointing toward resolution
- Flag short-names must be consistent with existing commands

## Pull request process

1. Fork the repository and create a feature branch
2. Make your changes and ensure `make lint` passes cleanly
3. Run `make test-e2e` and confirm all tests pass
4. Open a pull request against `main` with a clear description of what changed and why
5. Be patient — reviews may take a few days
