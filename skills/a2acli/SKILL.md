---
name: a2acli
description: Interact with Agent-to-Agent (A2A) protocol services from the command line. Use when you need to discover agent capabilities, send tasks, poll status, stream results, manage tasks, or authenticate with OAuth 2.1-protected A2A agents.
license: Apache-2.0
compatibility: Requires the a2acli binary. Install via brew, curl, or go install. See README for details.
---

# a2acli — A2A Command-Line Client

`a2acli` is a fully A2A Specification v1.0 compliant CLI for interacting with A2A agents. It supports gRPC, JSON-RPC, and REST transports with automatic selection, and OAuth 2.1 auth-code + PKCE authentication.

## Critical Rules for Agents

1. **Always pass `--output json`** (or `-n`) — disables the interactive TUI and emits JSON/NDJSON instead. Without this flag the CLI hangs in non-TTY contexts.
2. **Always pass `--wait` with `send`** — makes the call blocking and returns the final task result. Without `--wait`, `send` streams indefinitely.
3. **Check `status.state`** in the JSON output to determine success (`TASK_STATE_COMPLETED`) or failure (`TASK_STATE_FAILED`).
4. **For OAuth-protected agents** — run `auth login` once interactively (requires a browser). For non-interactive agent use, retrieve the stored token via `auth token` and pass it as `--token`.

## Command Index

| Command | What it does |
|---|---|
| `discover` | Fetch an agent's AgentCard (capabilities, skills, security schemes) |
| `send` | Send a message to initiate or continue a task |
| `subscribe` | Subscribe to a running task's event stream |
| `get` | Retrieve state and artifacts of a task by ID |
| `list tasks` | List historical tasks (server must support history) |
| `cancel` | Cancel an active task |
| `download` | Download artifacts from a completed task |
| `push-config` | Manage push-notification callbacks for a task |
| `conformance` | Run A2A conformance smoke checks against a live server |
| `a2ui validate` | Validate A2UI v1.0 extension wire conformance |
| `auth login` | Obtain an OAuth 2.1 token (browser-based, one-time) |
| `auth token` | Print the stored access token (for scripting) |
| `auth status` | Check stored token validity |
| `serve` | Spin up a local mock A2A agent for testing |

## Global Flags (apply to all commands)

| Flag | Short | Default | Description |
|---|---|---|---|
| `--service-url` | `-u` | `http://127.0.0.1:9001` | Base URL of the A2A service |
| `--output json` | `-n` | tui | **Required for agents.** Emit JSON instead of interactive UI |
| `--wait` | `-w` | false | **Required with `send` for agents.** Block until task completes |
| `--token` | `-t` | — | Bearer token. If omitted, stored token from `auth login` is used automatically |
| `--auth` | — | — | Raw auth header, e.g. `ApiKey secret` (repeatable) |
| `--task` | `-k` | — | Existing Task ID to continue (must be non-terminal) |
| `--ref` | `-r` | — | Completed Task ID to pass as context |
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
