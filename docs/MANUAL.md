# a2acli Reference Manual

The complete command and flag reference for `a2acli`. For a gentle introduction,
start with the [README's 5-minute tutorial](../README.md). This document is the
manual you reach for once you know your way around.

## Contents

- [Command Grammar](#command-grammar)
- [Output Modes](#output-modes)
- [Discovery & Identity](#discovery--identity)
- [Messaging & Tasks](#messaging--tasks)
- [Server & Mocking](#server--mocking)
- [Authentication](#authentication)
- [Conformance](#conformance)
- [Client Configuration](#client-configuration)
- [Global Flags](#global-flags)
- [Shell Completion](#shell-completion)
- [Agent & Automation](#agent--automation)

## Command Grammar

```
a2acli <verb> [noun] [positional-args] [flags]
```

**Verbs** map directly to A2A Protocol RPCs. **Nouns** are used where a verb spans
multiple resource types (`list tasks`). **Positional args** are required IDs or
message text. **Flags** are optional modifiers.

| Pattern | Example | Notes |
|---|---|---|
| `verb` | `a2acli discover` | Verb only — single resource type |
| `verb message` | `a2acli send "Hello"` | Positional message text |
| `verb id` | `a2acli get <taskID>` | Positional task ID |
| `verb noun` | `a2acli list tasks` | Noun disambiguates resource |

The agent URL is always a flag (`--service-url / -u`) rather than a positional
argument. This enables named environment profiles in config — you rarely need to
type a URL at all once configured.

## Output Modes

Output modes are controlled by `--output`:

| Mode | Flag | Use when |
|---|---|---|
| `tui` | default | Interactive terminal with streaming UI |
| `text` | `--output text` | Non-interactive, human-readable (CI logs, piped output) |
| `json` | `--output json` | Machine-readable NDJSON (scripts, agents); errors emit structured JSON objects on `stderr` |

`-n` / `--no-tui` is a backwards-compatible shorthand for `--output json`.

In `--output json` mode, errors on `stderr` are emitted as structured JSON objects:
```json
{
  "code": "UNAUTHENTICATED",
  "error": "SendMessage failed: unexpected HTTP status: 401 Unauthorized",
  "hint": "Stored token for https://agent.example.com is expired. Run: a2acli auth login -u https://agent.example.com"
}
```
Machine-readable error codes include `UNAUTHENTICATED`, `TIMEOUT`, `TASK_FAILED`, `INVALID_ARGUMENT`, `NOT_FOUND`, and `INTERNAL_ERROR`.

When stdout is not a terminal, `a2acli` automatically degrades from `tui` to
`text`, so streaming works correctly in pipes, CI, and agent contexts without any
flags. It also degrades on `CI=true` and `NO_COLOR`.

## Discovery & Identity

### `discover` — Inspect an Agent

Fetch and display an agent's `AgentCard`, registered skills, and security
requirements. *(Fetches via the standard A2A discovery endpoint.)*

```bash
a2acli discover --service-url http://localhost:9001

# Fetch the richer, authenticated extended card (requires auth)
a2acli discover --service-url https://agent.example.com --extended
```

| Flag | Description |
|---|---|
| `--extended` | Fetch the authenticated extended AgentCard via `GetExtendedAgentCard` (requires a token; the agent must advertise `extendedAgentCard: true`) |

## Messaging & Tasks

### `send` — Send a Message

Send a message to initiate or continue a task. *(Maps to the A2A Protocol's
`SendMessage` RPC.)*

```bash
a2acli send "Generate a project plan" --out-dir ./output/
```

By default, `send` streams real-time updates. Use `--wait` (`-w`) for a blocking
call that returns the final result only. When stdin is not a terminal, the message
is read from stdin:

```bash
echo "Summarize Q3 results" | a2acli send --skill summarize --wait --output json
cat prompt.txt | a2acli send --wait
```

| Flag | Short | Description |
|---|---|---|
| `--skill` | `-s` | Target a specific skill on the agent |
| `--wait` / `--sync` | `-w` | Block until task completes (returns final JSON) |
| `--full` | — | Show complete artifact content without truncation (default: 500 char preview) |
| `--out-dir` | `-o` | Save artifacts to a directory |
| `--file` | `-f` | Save artifact to a specific filename |
| `--instruction-file` | `-i` | Path to a file with supplemental instructions |

### `subscribe` (`watch`) — Subscribe to a Task

Subscribe to an active task's event stream. *(Maps to the A2A Protocol's
`SubscribeToTask` RPC.)*

```bash
a2acli subscribe <task_id> --out-dir ./output/
```

| Flag | Short | Description |
|---|---|---|
| `--out-dir` | `-o` | Save artifacts to a directory as they arrive |
| `--file` | `-f` | Save artifact to a specific filename |

### `get` — Get Task Status

Retrieve the state and artifacts of a specific task. *(Maps to the A2A Protocol's
`GetTask` RPC.)*

```bash
a2acli get <task_id> --out-dir ./output/
```

| Flag | Short | Description |
|---|---|---|
| `--out-dir` | `-o` | Save artifacts to a directory |
| `--file` | `-f` | Save artifact to a specific filename |
| `--full` | — | Show complete artifact content without truncation |

### `list` — List Tasks

Query an agent for historical tasks. *(Maps to the A2A Protocol's `ListTasks` RPC.
The server must support history.)*

```bash
a2acli list tasks --limit 10
a2acli list tasks --status completed
a2acli list tasks --context ctx-123
```

| Flag | Default | Description |
|---|---|---|
| `--limit` | `10` | Maximum number of tasks to return |
| `--page-token` | — | Pagination token for the next page |
| `--context` | — | Filter by context ID |
| `--status` | — | Filter by task state (`submitted`, `working`, `completed`, `failed`, `canceled`, `rejected`) |

### `cancel` — Cancel a Task

Cancel an active task. *(Maps to the A2A Protocol's `CancelTask` RPC.)*

```bash
a2acli cancel <task_id>
```

### `push-config` — Push Notification Configs

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

### `download` — Download Artifacts

Download artifacts produced by a task to a local directory.

```bash
a2acli download <task_id> --out-dir ./downloads
```

| Flag | Short | Description |
|---|---|---|
| `--out-dir` | `-o` | Directory to save artifacts to |
| `--file` | `-f` | Save artifact to a specific filename |

## Server & Mocking

### `serve` — Run a Mock Agent

Spin up an A2A-compliant echo agent locally for testing and development.

```bash
a2acli serve --echo --port 9001
```

| Flag | Default | Description |
|---|---|---|
| `--port` | `9001` | Listen port |
| `--host` | `127.0.0.1` | Bind address |
| `--echo` | — | Return the user's message as the response |

## Authentication

The `auth` command obtains, inspects, and revokes OAuth 2.1 tokens for agents that
require authentication. Tokens are stored in `~/.config/a2acli/tokens/` (0600
permissions) and used automatically — no `--token` flag needed once logged in.

```bash
# Log in once — browser opens for OAuth 2.1 auth-code + PKCE flow
a2acli auth login --service-url https://agent.example.com

# Or log in using a named environment
a2acli auth login --env mithlond

# Check stored token validity
a2acli auth status --service-url https://agent.example.com

# Print the raw JWT (useful for scripts that need the token explicitly)
TOKEN=$(a2acli auth token --service-url https://agent.example.com)

# Remove the stored token
a2acli auth logout --service-url https://agent.example.com
```

After `auth login`, all commands (`send`, `discover`, `conformance`, etc.) use the
stored token automatically. If an access token expires and a valid refresh token
is stored, `a2acli` proactively exchanges the refresh token for a new access token
automatically before executing requests. The `auth token` subcommand is the
`--token $(make token)` equivalent for scripts.

> **Note for non-interactive contexts (CI, agents):** `auth login` requires a
> browser. Use `auth token` to retrieve a previously stored token for `--token` in
> automated contexts, or pass the JWT directly via `--token`.

## Conformance

### `conformance` — A2A Conformance Smoke Check

Run a quick sequence of checks against a live A2A server: AgentCard well-formed,
auth gating (auto-uses stored token if available), and a round-trip send. Non-zero
exit code on failure.

```bash
a2acli conformance --service-url http://localhost:9001
a2acli conformance --env mithlond --output json   # uses stored token automatically
```

### `a2ui validate` — A2UI Extension Conformance

Validate that a server emits a conformant **A2UI v1.0** extension stream — at the
wire level, against the official A2UI JSON Schemas, with no UI renderer required.

```bash
a2acli a2ui validate --service-url http://localhost:9002
a2acli a2ui validate -u http://localhost:9002 --output json
```

## Client Configuration

```bash
a2acli config                   # Show active environment and config file location
a2acli config env list          # List configured environments and token status
a2acli config env add <name> --service-url <url> [--transport <grpc|jsonrpc|rest>]
                                # Add or update a named environment profile
a2acli config env use <name>    # Set the default environment
a2acli config env remove <name> # Delete an environment profile
a2acli version                  # Print version information
```

`a2acli` supports named environments via an XDG Base Directory compliant config
file at `~/.config/a2acli/config.yaml`. This lets you switch between local,
staging, and production agents without repeating URLs and tokens.

```yaml
# ~/.config/a2acli/config.yaml
default_env: "local"

envs:
  local:
    service_url: "http://127.0.0.1:9001"

  staging:
    service_url: "https://staging-agent.internal.corp"
    token: "my-staging-auth-token"         # static token; or omit and use auth login

  mithlond:
    service_url: "https://candir.mithlond.com"
    # token omitted — stored automatically by 'a2acli auth login --env mithlond'
    # transport omitted — auto-selected from AgentCard (JSONRPC in this case)
```

Use the `--env` (`-e`) flag to select an environment:

```bash
a2acli auth login --env mithlond           # one-time browser login
a2acli send "name star silver quenya" --env mithlond --skill name-generate --wait
a2acli conformance --env mithlond          # auth gating auto-resolved from token store
```

Supported fields per environment: `service_url`, `token` (static, takes precedence
over token store), `transport` (pin a specific transport, e.g. `jsonrpc`).

Precedence: **CLI Flags > Environment Variables > Config File > Defaults.**

Environment variables follow the pattern `A2ACLI_<FLAG>` (e.g.
`A2ACLI_SERVICE_URL`).

## Global Flags

| Flag | Description |
|---|---|
| `-u, --service-url` | Base URL of the A2A service (default: `http://127.0.0.1:9001`) |
| `-t, --token` | Authorization token (if omitted, stored token from `auth login` is used automatically) |
| `--auth` | Authorization headers, e.g. `Bearer …` (repeatable) |
| `--svc-param` | Service parameters, e.g. `key=value` (repeatable) |
| `-k, --task` | Existing Task ID to continue (for active tasks) |
| `--context` | Context ID for multi-turn conversation thread |
| `-r, --ref` | Task ID to reference for cross-task artifact chaining (does not continue conversation) |
| `--strict` | Fail fast on warnings (e.g. continuing terminal tasks) |
| `-n, --no-tui` | Output JSON/NDJSON instead of the interactive TUI |
| `--output` | Output mode: `tui` (default), `text` (plain/CI), `json` (NDJSON for scripting) |
| `-v, --verbose` | Print diagnostic info to stderr (transport, token resolution, events) |
| `-p, --protocol` | A2A protocol version: `1.0.0` or `0.3.0` (default: `1.0.0`) |
| `--transport` | Force transport: `grpc`, `jsonrpc`, or `rest` |
| `--timeout` | Request timeout, e.g. `30s`, `2m` (default: no timeout) |
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

## Agent & Automation

`a2acli` is designed to be driven by AI coding agents (Claude Code, Cursor, GitHub
Copilot CLI) as well as shell scripts.

### Non-Interactive Mode (`-n`)

The `-n` / `--no-tui` flag switches all output to newline-delimited JSON (NDJSON),
giving scripts and agents a stable, parseable stream. It can also be set via
`A2ACLI_NO_TUI=true` or `NO_COLOR=true`.

```bash
a2acli send "Write code" -n --wait
```

### Transport Selection

`a2acli` auto-selects the best available transport based on the agent's advertised
capabilities, in priority order: **gRPC > JSON-RPC > REST**. Override when needed:

```bash
a2acli send "Generate video" --transport grpc
```

> When using `--protocol 0.3.0`, only `jsonrpc` is available. gRPC is disabled for
> legacy connections to prevent protobuf namespace conflicts.

### Proactive Error Hints

On failure, the CLI emits a `Hint:` to assist automated recovery:

```
Error: failed to resolve AgentCard: connection refused
Hint: Ensure the A2A server is running at http://localhost:9001
```

### Agent Skills

This repository ships three [`agentskills.io`](https://agentskills.io/) compliant agent skills in [`skills/`](../skills/):

- [`a2acli`](../skills/a2acli/SKILL.md) — Teaches agents how to drive `a2acli` using `--output json` and `--wait`.
- [`a2a-expose`](../skills/a2a-expose/SKILL.md) — Guides agents in adding an A2A Protocol v1.0 exposure layer to an API or service.
- [`a2a-conformance`](../skills/a2a-conformance/SKILL.md) — Evaluates A2A servers against normative spec rules, stable check IDs, and TCK orchestration.

For full usage scenarios, prompt libraries, and authoring guidelines, see the [**Agent Skills Guide (`docs/SKILLS.md`)**](SKILLS.md).
