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

<!--
Single source of truth: the authoritative a2acli skill content lives at the
repository root, ../../../skills/a2acli/SKILL.md (github.com/ghchinoy/a2acli
tree/main/skills/a2acli). This plugin copy is deliberately a thin pointer: it
does NOT restate the command index, flag tables, or usage rules, so there is
nothing here to keep in sync and nothing to drift. Update the root skill; this
file should stay a pointer.
-->

# Driving A2A agents with the `a2acli` CLI

`a2acli` is an A2A Specification v1.0-compliant client: give it an agent URL and a
message, it negotiates the transport from the agent's card (JSON-RPC, REST, or
gRPC), sends the message, and reports what the agent returned.

This is the Agent Plugin packaging of the `a2acli` skill. To avoid two copies
drifting apart, it does not restate the command index, flag tables, or usage
rules here — those are maintained in a single place.

## Preflight

Run `a2acli version` first. If it is missing, install it (see the compatibility
note above) or fail cleanly — do not guess.

## Canonical usage guidance (single source of truth)

The authoritative, maintained skill content — command index, global flags,
canonical-vs-legacy flag spellings, non-interactive (`-o json`) rules, exit
codes, authentication, and worked examples — lives in the **repository-root
`a2acli` skill**, not in this file:

- In a checkout: [`../../../skills/a2acli/SKILL.md`](../../../skills/a2acli/SKILL.md)
- On GitHub: <https://github.com/ghchinoy/a2acli/tree/main/skills/a2acli>

That root skill also ships `references/` (auth, send, get, serve, and more) for
depth. For the current, definitive flag surface of any subcommand, defer to
`a2acli <command> --help` on the installed binary — it is always in sync with
the binary you are driving.
