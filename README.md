# A2A CLI

[![GitHub Release](https://img.shields.io/github/v/release/ghchinoy/a2acli)](https://github.com/ghchinoy/a2acli/releases)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

A standalone, A2A Specification v1.0 compliant command-line client for discovering, messaging, and managing agents. Built on the [a2a-go](https://github.com/a2aproject/a2a-go) SDK with an interactive streaming TUI and a scriptable JSON mode.

## Quick Start

```bash
# Describe an agent — fetch its AgentCard, skills, and capabilities
a2acli discover --service-url http://localhost:9001

# Send a message and stream responses in real time
a2acli send "Generate a project plan" --service-url http://localhost:9001

# Send a message and get a single JSON result (for scripting and agents)
a2acli send "Generate a project plan" --service-url http://localhost:9001 -n --wait
```

## Installation

### macOS and Linux — Homebrew

```bash
brew tap ghchinoy/tap
brew trust ghchinoy/tap
brew install a2acli
```

### Linux — Install Script

```bash
curl -sL https://raw.githubusercontent.com/ghchinoy/a2acli/main/scripts/install.sh | bash
```

### Linux — apt (Debian / Ubuntu)

Download the `.deb` from the [latest release](https://github.com/ghchinoy/a2acli/releases/latest):

```bash
sudo dpkg -i a2acli_*.deb
```

### Linux — rpm (Fedora / RHEL)

Download the `.rpm` from the [latest release](https://github.com/ghchinoy/a2acli/releases/latest):

```bash
sudo rpm -i a2acli_*.rpm
```

### Any platform — Go Install

```bash
go install github.com/ghchinoy/a2acli/cmd/a2acli@latest
```

### From Source

```bash
git clone https://github.com/ghchinoy/a2acli.git
cd a2acli
make build
```

The binary is written to `bin/a2acli`.

## Command Grammar

```
a2acli <verb> [noun] [positional-args] [flags]
```

**Verbs** map directly to A2A Protocol RPCs. **Nouns** are used where a verb spans multiple resource types (`list tasks`). **Positional args** are required IDs or message text. **Flags** are optional modifiers.

| Pattern | Example | Notes |
|---|---|---|
| `verb` | `a2acli discover` | Verb only — single resource type |
| `verb message` | `a2acli send "Hello"` | Positional message text |
| `verb id` | `a2acli get <taskID>` | Positional task ID |
| `verb noun` | `a2acli list tasks` | Noun disambiguates resource |

The agent URL is always a flag (`--service-url / -u`) rather than a positional argument. This enables named environment profiles in config — you rarely need to type a URL at all once configured.

**Output modes** are controlled by `--output`:

| Mode | Flag | Use when |
|---|---|---|
| `tui` | default | Interactive terminal with streaming UI |
| `text` | `--output text` | Non-interactive, human-readable (CI logs, piped output) |
| `json` | `--output json` | Machine-readable NDJSON (scripts, agents) |

`-n` / `--no-tui` is a backwards-compatible shorthand for `--output json`.

## Commands

Commands are organized into four A2A-aligned groups: **Discovery & Identity**, **Messaging & Tasks**, **Server & Mocking**, and **Client Configuration**.

### Discovery & Identity

#### `describe` — Inspect an Agent
Fetch and display an agent's `AgentCard`, registered skills, and security requirements.
*(Fetches via the standard A2A discovery endpoint.)*

```bash
a2acli discover --service-url http://localhost:9001
```

### Messaging & Tasks

#### `send` — Send a Message
Send a message to initiate or continue a task.
*(Maps to the A2A Protocol's `SendMessage` RPC.)*

```bash
a2acli send "Generate a project plan" --out-dir ./output/
```

By default, `send` streams real-time updates. Use `--wait` (`-w`) for a blocking call that returns the final result only. When stdin is not a terminal, the message is read from stdin:

```bash
echo "Summarize Q3 results" | a2acli send --skill summarize --wait --output json
cat prompt.txt | a2acli send --wait
```

| Flag | Short | Description |
|---|---|---|
| `--skill` | `-s` | Target a specific skill on the agent |
| `--wait` / `--sync` | `-w` | Block until task completes (returns final JSON) |
| `--out-dir` | `-o` | Save artifacts to a directory |
| `--file` | `-f` | Save artifact to a specific filename |
| `--instruction-file` | `-i` | Path to a file with supplemental instructions |

#### `watch` — Subscribe to a Task
Subscribe to an active task's event stream.
*(Maps to the A2A Protocol's `SubscribeToTask` RPC.)*

```bash
a2acli subscribe <task_id> --out-dir ./output/
```

| Flag | Short | Description |
|---|---|---|
| `--out-dir` | `-o` | Save artifacts to a directory as they arrive |
| `--file` | `-f` | Save artifact to a specific filename |

#### `get` — Get Task Status
Retrieve the state and artifacts of a specific task.
*(Maps to the A2A Protocol's `GetTask` RPC.)*

```bash
a2acli get <task_id> --out-dir ./output/
```

| Flag | Short | Description |
|---|---|---|
| `--out-dir` | `-o` | Save artifacts to a directory |
| `--file` | `-f` | Save artifact to a specific filename |

#### `list` — List Tasks
Query an agent for historical tasks.
*(Maps to the A2A Protocol's `ListTasks` RPC. The server must support history.)*

```bash
a2acli list tasks --limit 10
```

| Flag | Default | Description |
|---|---|---|
| `--limit` | `10` | Maximum number of tasks to return |
| `--page-token` | — | Pagination token for the next page |

#### `cancel` — Cancel a Task
Cancel an active task.
*(Maps to the A2A Protocol's `CancelTask` RPC.)*

```bash
a2acli cancel <task_id>
```

#### `push-config` — Push Notification Configs
Register, list, retrieve, and delete push-notification callbacks for a task.
*(Maps to the A2A Protocol's `CreateTaskPushNotificationConfig` and related RPCs.)*

```bash
a2acli push-config create <task_id> https://myserver.example.com/notify
a2acli push-config create <task_id> https://cb.example.com/notify \
  --auth-scheme Bearer --auth-credentials mytoken --id my-config
a2acli push-config list <task_id>
a2acli push-config get <task_id> <config_id>
a2acli push-config delete <task_id> <config_id>
```

#### `download` — Download Artifacts
Download artifacts produced by a task to a local directory.

```bash
a2acli download <task_id> --out-dir ./downloads
```

| Flag | Short | Description |
|---|---|---|
| `--out-dir` | `-o` | Directory to save artifacts to |
| `--file` | `-f` | Save artifact to a specific filename |

### Server & Mocking

#### `conformance` — A2A Conformance Smoke Check
Run a quick sequence of checks against a live A2A server: AgentCard well-formed,
auth gating (if applicable), and a round-trip send. Non-zero exit code on failure.

```bash
a2acli conformance --service-url http://localhost:9001
a2acli conformance --service-url https://agent.example.com --token mytoken --output json
```

#### `serve` — Run a Mock Agent
Spin up an A2A-compliant echo agent locally for testing and development.

```bash
a2acli serve --echo --port 9001
```

| Flag | Default | Description |
|---|---|---|
| `--port` | `9001` | Listen port |
| `--host` | `127.0.0.1` | Bind address |
| `--echo` | — | Return the user's message as the response |

### Client Configuration

```bash
a2acli config    # Show active environment and config file location
a2acli version   # Print version information
```

## Shell Completion

`a2acli` can generate completion scripts for bash, zsh, fish, and PowerShell.

### zsh

```bash
a2acli completion zsh > "${fpath[1]}/_a2acli"
```

Or for a single session:
```bash
source <(a2acli completion zsh)
```

### bash

```bash
a2acli completion bash > /etc/bash_completion.d/a2acli
```

Or for a single session:
```bash
source <(a2acli completion bash)
```

### fish

```bash
a2acli completion fish > ~/.config/fish/completions/a2acli.fish
```

### PowerShell

```powershell
a2acli completion powershell | Out-String | Invoke-Expression
```

## Configuration

`a2acli` supports named environments via an XDG Base Directory compliant config file at `~/.config/a2acli/config.yaml`. This lets you switch between local, staging, and production agents without repeating URLs and tokens.

```yaml
# ~/.config/a2acli/config.yaml
default_env: "local"

envs:
  local:
    service_url: "http://127.0.0.1:9001"
  staging:
    service_url: "https://staging-agent.internal.corp"
    token: "my-staging-auth-token"
  prod:
    service_url: "https://agent.example.com"
    token: "my-secure-prod-token"
```

Use the `--env` (`-e`) flag to select an environment:

```bash
a2acli send "Generate report" --env staging
```

Precedence: **CLI Flags > Environment Variables > Config File > Defaults.**

Environment variables follow the pattern `A2ACLI_<FLAG>` (e.g. `A2ACLI_SERVICE_URL`).

### Global Flags

| Flag | Description |
|---|---|
| `-u, --service-url` | Base URL of the A2A service (default: `http://127.0.0.1:9001`) |
| `-t, --token` | Authorization token |
| `--auth` | Authorization headers, e.g. `Bearer …` (repeatable) |
| `--svc-param` | Service parameters, e.g. `key=value` (repeatable) |
| `-k, --task` | Existing Task ID to continue (must be non-terminal) |
| `-r, --ref` | Task ID to reference as context |
| `-n, --no-tui` | Output JSON/NDJSON instead of the interactive TUI |
| `-p, --protocol` | A2A protocol version: `1.0.0` or `0.3.0` (default: `1.0.0`) |
| `--transport` | Force transport: `grpc`, `jsonrpc`, or `rest` |
| `-e, --env` | Named environment from config file |
| `-c, --config` | Path to config file |
| `-V, --version` | Print version information |

**Example: auth and service parameters**

```bash
a2acli send "Generate report" --service-url http://localhost:9001 \
  --auth "ApiKey secret-key-here" \
  --svc-param "tenant_id=123" \
  --svc-param "debug=true"
```

## Agent & Automation

`a2acli` is designed to be driven by AI coding agents (Claude Code, Cursor, GitHub Copilot CLI) as well as shell scripts.

### Non-Interactive Mode (`-n`)

The `-n` / `--no-tui` flag switches all output to newline-delimited JSON (NDJSON), giving scripts and agents a stable, parseable stream. It can also be set via `A2ACLI_NO_TUI=true` or `NO_COLOR=true`.

```bash
a2acli send "Write code" -n --wait
```

### Transport Selection

`a2acli` auto-selects the best available transport based on the agent's advertised capabilities, in priority order: **gRPC > JSON-RPC > REST**. Override when needed:

```bash
a2acli send "Generate video" --transport grpc
```

> When using `--protocol 0.3.0`, only `jsonrpc` is available. gRPC is disabled for legacy connections to prevent protobuf namespace conflicts.

### Proactive Error Hints

On failure, the CLI emits a `Hint:` to assist automated recovery:

```
Error: failed to resolve AgentCard: connection refused
Hint: Ensure the A2A server is running at http://localhost:9001
```

### Agent Skills

This repository ships an [`agentskills.io`](https://agentskills.io/) compliant `skills/` directory. AI coding agents load these skills automatically and learn to use `--no-tui` and `--wait` for deterministic JSON output.

## Development

```bash
make build      # Compile to bin/a2acli
make run        # Build and run
make lint       # Run golangci-lint
make test-e2e   # Run end-to-end conformance tests
make clean      # Remove bin/
```

For release instructions see [docs/RELEASING.md](docs/RELEASING.md).

### Conformance (TCK)

`a2acli` is tested against the official A2A Technology Compatibility Kit for both **v0.3.0** and **v1.0.0**. See the [Conformance Report](docs/CONFORMANCE_REPORT.md) for current status.

Running the tests requires the [a2a-go](https://github.com/a2aproject/a2a-go) SDK source locally, as the suite spins up the TCK SUT server dynamically:

```bash
# Default path: ../../github/a2a-go
make test-e2e

# Custom path
make test-e2e A2A_GO_SRC=/path/to/a2a-go
```

To run the SUT manually:

```bash
# In the a2a-go repository
cd e2e/tck
go run sut.go sut_agent_executor.go
```

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines and [docs/CLI_DESIGN_BEST_PRACTICES.md](docs/CLI_DESIGN_BEST_PRACTICES.md) for design conventions to follow before adding or modifying commands.

## License

Apache 2.0. See [LICENSE](LICENSE).
