# A2A CLI

[![GitHub Release](https://img.shields.io/github/v/release/ghchinoy/a2acli)](https://github.com/ghchinoy/a2acli/releases)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

A standalone command-line client for interacting with Agent-to-Agent (A2A) services. It is fully compliant with the **A2A Specification v1.0**. It is built using the [a2a-go](https://github.com/a2aproject/a2a-go) SDK and provides both an interactive streaming terminal UI and a scriptable non-interactive JSON mode.

## 🌟 Features

- **A2A-Aligned Discovery**: Inspect an agent's `AgentCard`, skills, and capabilities using protocol-native terminology (`describe`).
- **Structured Messaging**: Full lifecycle support for A2A tasks, including `send` (initiate), `watch` (subscribe), and `get` (retrieve).
- **Agent-First Design**: Built-in support for non-interactive JSON output, deterministic exit codes, and proactive "Hint" guidance for automated agents.
- **Dynamic Transport Selection**: Automatically selects the best available protocol (gRPC, JSON-RPC, or HTTP+JSON) based on advertised agent capabilities.
- **Interactive TUI**: A beautiful [Bubble Tea](https://github.com/charmbracelet/bubbletea) interface for real-time streaming updates and artifact previews.
- **Configuration Management**: First-class support for XDG-compliant multi-environment configurations and auth token interception.

## 📦 Installation

### 1. Via Go Install (Recommended)

If you have [Go 1.25+](https://go.dev/) installed, you can easily install the CLI globally via `go install`:

```bash
go install github.com/ghchinoy/a2acli/cmd/a2acli@latest
```

### 2. Via Install Script

Alternatively, you can run the provided install script to download and install the latest pre-compiled binary:

```bash
curl -sL https://raw.githubusercontent.com/ghchinoy/a2acli/main/scripts/install.sh | bash
```

### 3. Via Source

Ensure you have [Go 1.25+](https://go.dev/) installed.

```bash
git clone https://github.com/ghchinoy/a2acli.git
cd a2acli
make build
```

This will produce the `a2acli` binary in the `bin/` directory.

Alternatively, you can run it directly:

```bash
make run
```

## 🚀 Usage

Use the `--help` flag on any command to see the available options and A2A-aligned command groups.

```bash
a2acli --help
```

### Global Flags

- `-c, --config string`: Path to a specific config file (default is `~/.config/a2acli/config.yaml`)
- `-e, --env string`: Specific named environment to load from the config file
- `-u, --service-url string`: Base URL of the A2A service (default "http://127.0.0.1:9001")
- `-t, --token string`: Authorization token (if required by the agent)
- `-k, --task string`: Existing Task ID to continue a conversation/task (must be non-terminal)
- `-r, --ref string`: Task ID to reference as context (works for completed tasks)
- `-n, --no-tui`: Disables the interactive TUI and outputs JSON/NDJSON. Can also be set via `A2ACLI_NO_TUI=true` or `NO_COLOR=true`.
- `--transport string`: Force a specific transport protocol (`grpc`, `jsonrpc`, `httpjson`). Defaults to auto-selection based on the agent's card.
- `-V, --version`: Print version information.

### Configuration Management

`a2acli` supports managing multiple servers using an XDG Base Directory compliant configuration file. By default, it looks for `~/.config/a2acli/config.yaml`.

You can define multiple named environments to easily switch between local, staging, and production agents without typing URLs and tokens repeatedly:

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

To view your currently active configuration context, run:
```bash
a2acli config
```

To run a command using a specific environment from your config file, use the `--env` (`-e`) flag:
```bash
a2acli send "Generate report" --env staging
```

Environment variables are also supported (e.g., `A2ACLI_SERVICE_URL`). The precedence is: *CLI Flags > Environment Variables > Config File > Defaults.*

### Commands

Commands are organized into three A2A-aligned groups: **Discovery & Identity**, **Messaging & Tasks**, and **Client Configuration**.

#### 1. Discovery & Identity

##### Describe Agent
Inspect the agent's identity, registered skills, and security requirements. 
*(Fetches the AgentCard via the standard A2A discovery endpoint).*
```bash
a2acli describe --service-url http://localhost:9001
```

#### 2. Messaging & Tasks

##### Send a Message
Send a message to an agent to initiate or continue a task.
*(Maps to the A2A Protocol's `SendMessage` RPC).*

```bash
a2acli send "Generate a project plan" --out-dir ./output/
```

**Synchronous vs. Streaming:**
By default, `send` streams real-time updates. Use `--wait` (or `-w`) to perform a blocking call that waits for the final result.

##### Watch a Task
Subscribe to an active task's event stream.
*(Maps to the A2A Protocol's `SubscribeToTask` RPC).*
```bash
a2acli watch <task_id>
```

##### Get Task Status
Retrieve the state and artifacts of a specific task.
*(Maps to the A2A Protocol's `GetTask` RPC).*
```bash
a2acli get <task_id>
```

#### 3. Client Configuration

View the active environment settings and config file location.
```bash
a2acli config
a2acli version
```

## 🤖 Agent & Automation

`a2acli` is designed with coding agents and automation in mind.

### Non-Interactive Mode (`-n`)
Using the `-n` or `--no-tui` flag ensures that all output is emitted as parseable JSON/NDJSON. This is ideal for scripts and agents that need to consume A2A data programmatically.

```bash
a2acli send "Write code" -n --wait
```

### Proactive Error Hints
When a command fails (e.g., server down, invalid skill), the CLI provides a "Hint:" to assist with automated recovery.

Example Error:
```text
Error: failed to resolve AgentCard: connection refused
Hint: Ensure the A2A server is running at http://localhost:9001
```

### Dynamic Transport Selection
Agents can automatically negotiate the most efficient transport protocol without manual intervention. By default, `a2acli` prioritizes protocols in the order: **gRPC > JSON-RPC > HTTP+JSON**.

For specialized environments, you can override this logic:
```bash
# Force gRPC for high-performance streaming
a2acli send "Generate video" --transport grpc
```

## 🛠️ Development

- `make build`: Compiles the binary to `bin/a2acli`.
- `make run`: Builds and runs the CLI.
- `make lint`: Runs `golangci-lint` configured for Google Go standards.
- `make test-e2e`: Runs the end-to-end conformance tests.
- `make clean`: Removes the `bin/` directory.

For details on how to build, publish, and release new versions of `a2acli`, see the [Releasing Guide](docs/RELEASING.md).

### 🏆 Testing Conformance (TCK)

To verify the CLI's v1.0 compliance, you can test it against the official A2A Technology Compatibility Kit (TCK) System Under Test (SUT) server. See the latest [Conformance Report](docs/CONFORMANCE_REPORT.md) for current status.

The `a2acli` contains an automated end-to-end conformance test suite that will build the CLI, spin up the TCK SUT server locally, and run black-box tests asserting the machine-readable JSON output of the CLI.

**Requirement:** Running the end-to-end conformance tests requires the source code of the `a2a-go` SDK to be present on your local machine, as the tests dynamically spin up the TCK SUT server from that repository ([https://github.com/a2aproject/a2a-go](https://github.com/a2aproject/a2a-go)).

By default, the `Makefile` assumes the `a2a-go` repository is cloned at `../../github/a2a-go` relative to the root of this project. 

If you have it cloned elsewhere, you must override the `A2A_GO_SRC` variable when running `make`:

```bash
make test-e2e A2A_GO_SRC=/absolute/or/relative/path/to/a2a-go
```

To run the SUT manually for your own testing:

```bash
# In the a2a-go repository
cd e2e/tck
go run sut.go sut_agent_executor.go
```

## 🤝 Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) file for guidelines on how to get involved. 

We follow a set of [CLI Design Best Practices](docs/CLI_DESIGN_BEST_PRACTICES.md) to ensure the tool remains usable for both humans and agents. Please review these before submitting a PR that adds or modifies commands.

---

## 📄 License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for more details.