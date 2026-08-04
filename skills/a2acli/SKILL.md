---
name: a2acli
description: Interact with Agent-to-Agent (A2A) protocol services from the command line. Use when you need to discover agent capabilities, send tasks, poll status, stream results, manage tasks, or authenticate with OAuth 2.1-protected A2A agents.
license: Apache-2.0
compatibility: Requires the a2acli binary on PATH (brew install ghchinoy/tap/a2acli or go install). Verify via 'a2acli version'. See references/install.md.
metadata:
  category: cli-usage
---

# a2acli — A2A Command-Line Client

`a2acli` is a fully A2A Specification v1.0 compliant CLI for interacting with A2A agents. It supports gRPC, JSON-RPC, and REST transports with automatic selection, and OAuth 2.1 auth-code + PKCE authentication.

## Critical Rules for Agents

1. **Always pass `--output json`** (or `-n`) — disables the interactive TUI and emits JSON/NDJSON instead. Errors emit structured JSON objects on `stderr` (`{"code": "...", "error": "...", "hint": "..."}`). Without this flag the CLI degrades in non-TTY contexts.
2. **Always pass `--wait` with `send`** — makes the call blocking and returns the final task result. Without `--wait`, `send` streams indefinitely.
3. **Check `status.state`** in the JSON output to determine success (`TASK_STATE_COMPLETED`) or failure (`TASK_STATE_FAILED`).
4. **For OAuth-protected agents** — run `auth login` once interactively (requires a browser). For non-interactive agent use, retrieve the stored token via `auth token` and pass it as `--token`.
5. **Verify binary availability first** — run `a2acli version` before execution. If missing, follow [references/install.md](references/install.md) or fail cleanly.

## Command Index

| Command | What it does |
|---|---|
| `discover` | Fetch an agent's AgentCard (capabilities, skills, security schemes); `--extended` for the authenticated card |
| `send` | Send a message to initiate or continue a task; multi-modal via `--parts/--json/--attach/--data` |
| `subscribe` | Subscribe to a running task's event stream |
| `get` | Retrieve state and artifacts of a task by ID |
| `list tasks` | List historical tasks (server must support history); filter with `--context`/`--status` |
| `cancel` | Cancel an active task |
| `download` | Download artifacts from a completed task |
| `push-config` | Manage push-notification callbacks for a task |
| `conformance` | Run A2A conformance smoke checks against a live server |
| `a2ui validate` | Validate A2UI v1.0 extension wire conformance |
| `auth login` | Obtain an OAuth 2.1 token (browser-based, one-time) |
| `auth token` | Print the stored access token (for scripting) |
| `auth status` | Check stored token validity |
| `config env` | Manage named environments (`add`/`remove`/`use`/`list`) |
| `serve` | Spin up a local mock A2A agent for testing |

## Global Flags (apply to all commands)

| Flag | Short | Default | Description |
|---|---|---|---|
| `--service-url` | `-u` | `http://127.0.0.1:9001` | Base URL of the A2A service |
| `--output json` | `-n` | tui | **Required for agents.** Emit JSON instead of interactive UI |
| `--wait` | `-w` | false | **Required with `send` for agents.** Block until task completes |
| `--token` | `-t` | — | Bearer token. If omitted, stored token from `auth login` is used automatically |
| `--auth` | — | — | Raw auth header, e.g. `ApiKey secret` (repeatable) |
| `--task` | `-k` | — | Existing Task ID to continue (for active tasks) |
| `--context` | — | — | Context ID for multi-turn conversation thread |
| `--ref` | `-r` | — | Task ID to reference for cross-task artifact chaining |
| `--strict` | — | false | Fail fast on warnings (e.g. continuing terminal tasks) |
| `--protocol` | `-p` | `1.0.0` | A2A protocol version (`1.0.0` or `0.3.0`) |
| `--transport` | — | auto | Force transport: `grpc`, `jsonrpc`, or `rest` |
| `--env` | `-e` | — | Named environment from config file |
| `--verbose` | `-v` | false | Diagnostic output to stderr (transport, token resolution) |

## Authentication

For OAuth 2.1-protected agents:

```bash
# One-time interactive login (human must complete browser flow)
a2acli auth login --service-url https://agent.example.com

# After login, all commands use the stored token automatically — no --token needed
a2acli send "hello" --service-url https://agent.example.com --output json --wait

# For non-interactive agent use, retrieve the token explicitly
TOKEN=$(a2acli auth token --service-url https://agent.example.com)
a2acli send "hello" --service-url https://agent.example.com --token "$TOKEN" --output json --wait
```

See [references/auth.md](references/auth.md) for the full auth workflow.

## Minimal Working Examples

```bash
# Discover what an agent can do (no auth)
a2acli discover --service-url http://localhost:9001 --output json

# Send a task and get JSON result (no auth)
a2acli send "Summarize this document" --service-url http://localhost:9001 --output json --wait

# Send a task with auto-authentication (after auth login)
a2acli send "translate 'hello' to Sindarin" --env mithlond --skill translate --output json --wait

# Check status of a running task
a2acli get <task_id> --service-url http://localhost:9001 --output json
```

## Detailed Command Reference

For full flag lists and output schemas, load the relevant reference file:

- [send](references/send.md) — initiating and continuing tasks
- [discover](references/describe.md) — agent discovery and security scheme inspection
- [get](references/get.md) — task status and artifact retrieval
- [subscribe](references/watch.md) — streaming task subscription
- [list](references/list.md) — listing historical tasks
- [cancel](references/cancel.md) — cancelling tasks
- [download](references/download.md) — downloading artifacts
- [serve](references/serve.md) — running a local mock agent
- [auth](references/auth.md) — OAuth 2.1 authentication workflow
- [install](references/install.md) — obtaining & verifying binary and skills
