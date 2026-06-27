---
name: a2acli
description: Interact with Agent-to-Agent (A2A) protocol services from the command line. Use when you need to discover agent capabilities, send tasks, poll status, stream results, or manage tasks on any A2A-compliant agent.
license: Apache-2.0
compatibility: Requires the a2acli binary. Install via brew, curl, or go install. See README for details.
---

# a2acli — A2A Command-Line Client

`a2acli` is a fully A2A Specification v1.0 compliant CLI for interacting with A2A agents. It supports gRPC, JSON-RPC, and REST transports with automatic selection.

## Critical Rules for Agents

1. **Always pass `-n` / `--no-tui`** — disables the interactive TUI and emits JSON/NDJSON instead. Without this flag the CLI will hang waiting for terminal input.
2. **Always pass `-w` / `--wait` with `send`** — makes the call blocking and returns the final task result. Without `--wait`, `send` streams indefinitely.
3. **Check `status.state`** in the JSON output to determine success (`TASK_STATE_COMPLETED`) or failure (`TASK_STATE_FAILED`).

## Command Index

| Command | What it does |
|---|---|
| `describe` | Fetch an agent's AgentCard (capabilities, skills, auth requirements) |
| `send` | Send a message to initiate or continue a task |
| `watch` | Subscribe to a running task's event stream |
| `get` | Retrieve state and artifacts of a task by ID |
| `list tasks` | List historical tasks (server must support history) |
| `cancel` | Cancel an active task |
| `download` | Download artifacts from a completed task |
| `serve` | Spin up a local mock A2A agent for testing |

## Global Flags (apply to all commands)

| Flag | Short | Default | Description |
|---|---|---|---|
| `--service-url` | `-u` | `http://127.0.0.1:9001` | Base URL of the A2A service |
| `--no-tui` | `-n` | false | **Required for agents.** Emit JSON instead of interactive UI |
| `--token` | `-t` | — | Bearer token for auth |
| `--auth` | — | — | Raw auth header, e.g. `ApiKey secret` (repeatable) |
| `--task` | `-k` | — | Existing Task ID to continue (must be non-terminal) |
| `--ref` | `-r` | — | Completed Task ID to pass as context |
| `--protocol` | `-p` | `1.0.0` | A2A protocol version (`1.0.0` or `0.3.0`) |
| `--transport` | — | auto | Force transport: `grpc`, `jsonrpc`, or `rest` |
| `--env` | `-e` | — | Named environment from config file |

## Minimal Working Examples

```bash
# Discover what an agent can do
a2acli describe --service-url http://localhost:9001 -n

# Send a task and get JSON result
a2acli send "Summarize this document" --service-url http://localhost:9001 -n --wait

# Check status of a running task
a2acli get <task_id> --service-url http://localhost:9001 -n
```

## Detailed Command Reference

For full flag lists and output schemas, load the relevant reference file:

- [send](references/send.md) — initiating and continuing tasks
- [describe](references/describe.md) — agent discovery
- [get](references/get.md) — task status and artifact retrieval
- [watch](references/watch.md) — streaming task subscription
- [list](references/list.md) — listing historical tasks
- [cancel](references/cancel.md) — cancelling tasks
- [download](references/download.md) — downloading artifacts
- [serve](references/serve.md) — running a local mock agent
