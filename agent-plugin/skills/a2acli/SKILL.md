---
name: a2acli
description: >-
  Drive A2A (Agent2Agent) agents from the command line with the `a2acli` CLI
  (github.com/ghchinoy/a2acli). Use when you need to discover an agent's card,
  send a message to an agent at a URL, stream or poll a running task, or check,
  list, resume, or cancel an A2A task. Not for building or serving an A2A agent.
compatibility: >-
  Requires the `a2acli` binary on PATH. Install it with Homebrew
  (brew install ghchinoy/tap/a2acli), a prebuilt release, install.sh, or
  `go install`. The skill installs nothing itself; verify with `a2acli version`.
license: Apache-2.0
metadata:
  source: https://github.com/ghchinoy/a2acli
  category: cli-usage
---

# Driving A2A agents with the `a2acli` CLI

`a2acli` is an A2A Specification v1.0-compliant client: give it an agent URL and a
message, it negotiates the transport from the agent's card (JSON-RPC, REST, or
gRPC), sends the message, and reports what the agent returned. This skill is lean
by design — it defers to `a2acli <command> --help` for the full, authoritative
flag list.

## Preflight

Run `a2acli version` first. If it is missing, install it (see the compatibility
note above) or fail cleanly — do not guess.

## Rules for non-interactive (agent) use

1. **Always pass `-o json`** (or `--output json` / `-n`) to disable the TUI and
   emit machine-readable output. On failure the spec error envelope is printed on
   **stdout** (`{"error":{"code":"A2ACLI_ERR_...","message":"...","hint":"...","a2aCode":...}}`)
   while diagnostics stay on **stderr**.
2. **`send` blocks by default** and emits a single JSON document. Pass `--stream`
   only when you want the live JSONL event stream.
3. **Check `status.state`** in the output — `TASK_STATE_COMPLETED` is success,
   `TASK_STATE_FAILED`/`TASK_STATE_REJECTED` are failures.
4. **Exit codes:** `0` success, `2` usage error, `1` other failures.

## Common commands

| Goal | Command |
|---|---|
| Inspect an agent | `a2acli discover <url> -o json` |
| Send a message and wait | `a2acli send "..." -a <url> -o json` |
| Fire-and-forget | `a2acli send "..." -a <url> --async -o json` |
| Poll a task to completion | `a2acli get <taskId> --wait -o json` |
| Get task + history | `a2acli get <taskId> --history 20 -o json` |
| Cancel a task | `a2acli cancel <taskId> -o json` |
| Show effective config | `a2acli config show` |

Flags accept both the canonical spec spellings (`--agent-card`/`-a`,
`--endpoint`, `--context-id`, `--task-id`, `--a2a-version`, `--async`) and the
tool's original spellings (`--service-url`/`-u`, `--context`, `--task`/`-k`,
`--protocol`/`-p`, `--immediate`). Consult `a2acli <command> --help` for the
complete, current surface.

## Deeper guidance

The repository ships richer skills (with `references/` and `scripts/`) at
`github.com/ghchinoy/a2acli/tree/main/skills`, covering authentication, building
A2A exposure layers, and conformance auditing.
