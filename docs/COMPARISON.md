# a2acli — Proposal Comparison and Alignment

This document compares `a2acli` against the two community proposals that informed
the [a2aproject/A2A#1929](https://github.com/a2aproject/A2A/issues/1929) discussion
on an official canonical CLI:

- **Issue #1929** — feature request defining the initial scope
- **Discussion #306** — detailed command grammar and flag proposal from the `a2a-go` maintainer

---

## Command Coverage

| Operation | #1929 | #306 | a2acli | Notes |
|---|---|---|---|---|
| Agent discovery | `discover` | `discover` | `discover` | Aligned. `describe` kept as alias. |
| Send message | `send` | `send` | `send` | Aligned. |
| Stream response | — | `send --stream` | default | a2acli streams by default. |
| Fire-and-forget | `send --return-immediately` | `send --immediate` | `send --immediate` | Aligned with #306. |
| Blocking send | — | — | `send --wait` | a2acli addition. |
| Get task | `task get` | `get task` | `get` | Same operation; a2acli drops the noun (only gets tasks). |
| List tasks | `task list` | `list tasks` | `list tasks` | Aligned. |
| Cancel task | `task cancel` | `cancel` | `cancel` | Aligned. |
| Subscribe | `task subscribe` | `subscribe` | `subscribe` | Aligned. `watch` kept as alias. |
| Download artifacts | — | — | `download` | a2acli addition. |
| Mock echo server | — | `serve --echo` | `serve --echo` | Aligned. |
| Proxy server | — | `serve --proxy` | planned (`a2ac-6x6`) | Stub exists, not yet implemented. |
| Exec wrapping | — | `serve --exec` | planned (`a2ac-7n2`) | Stub exists, not yet implemented. |

---

## Global Flag Coverage

| Flag | #1929 | #306 | a2acli | Notes |
|---|---|---|---|---|
| Agent URL | positional | positional | `--service-url / -u` | **a2acli difference** — flag enables named environments. |
| Output format | `-o json` | `-o json` | `--output tui/text/json` | a2acli adds `tui` and `text` modes; no `-o` shorthand (conflict with `--out-dir`). |
| Auth | headers | `--auth` | `--token`, `--auth` | a2acli supports both bearer token shorthand and raw headers. |
| Service params | — | `--svc-param` | `--svc-param` | Aligned. |
| Transport | — | `--transport` | `--transport` | Aligned. |
| Timeout | — | `--timeout 30s` | `--timeout` | Aligned. Default is 0 (no timeout); 30s applied to agent card resolution. |
| Tenant | — | `--tenant` | planned | Not yet implemented. |
| Verbose | — | `--verbose / -v` | — | Not yet implemented. |
| Protocol version | — | — | `--protocol` | a2acli addition for v0.3.0 backward compat. |
| Named environment | — | — | `--env / -e` | a2acli addition. |
| Config file | — | — | `--config / -c` | a2acli addition. |

---

## `send` Flag Coverage

| Flag | #306 | a2acli | Notes |
|---|---|---|---|
| `--immediate` | ✓ | ✓ | Aligned. |
| `--wait` / `--sync` | — | ✓ | a2acli addition. |
| `--stream` | ✓ | default | a2acli streams by default; no explicit flag needed. |
| `--task` | `--task` | `--task / -k` | Aligned. |
| `--context` | `--context` | `--ref / -r` | Different semantics: `--ref` references a *completed* task, not a context ID. |
| `--skill` | — | `--skill / -s` | a2acli addition. |
| `--instruction-file` | `-f file` | `--instruction-file / -i` | Similar. #306 reads a full JSON Message; a2acli appends plain text. |
| `--parts` / `--json` | ✓ | planned (`a2ac-79d`) | Structured multi-modal input not yet implemented. |
| `--history` | ✓ | — | Not yet implemented. |
| `--out-dir` | — | `--out-dir / -o` | a2acli addition for artifact saving. |
| `--file` | — | `--file / -f` | a2acli addition. |

---

## Where a2acli Leads

These features exist in `a2acli` but are absent from both proposals:

**Named environment profiles**
Switch between local, staging, and production agents without re-typing URLs and tokens:
```bash
a2acli send "Generate report" --env staging
```

**`--output text` mode**
A third output mode between TUI and JSON — human-readable plain text with no animations.
Useful for CI log output where JSON is too noisy but the TUI won't render.

**`--output` mode integration with `NO_COLOR`**
`NO_COLOR=true` maps to `--output text` (plain human-readable), not `--output json`.
Most CLIs treat `NO_COLOR` as "disable ANSI codes" without switching to machine output.

**Multi-environment config file**
XDG-compliant `~/.config/a2acli/config.yaml` with named environments, precedence chain
(flags > env vars > config > defaults), and `a2acli config` for inspection.

**A2A v0.3.0 backward compatibility**
`--protocol 0.3.0` enables full backward-compatible operation against legacy agents.
No other proposal addresses this.

**`send --instruction-file`**
Appends a supplemental instruction file to the message — useful for large prompts
or structured context that exceeds comfortable shell quoting.

**agentskills.io compliant skill files**
`skills/a2acli/` provides a spec-compliant skill directory with progressive disclosure
(top-level SKILL.md + per-command `references/` files). AI coding agents load these
to learn correct `--output json` and `--wait` usage without manual configuration.

**TCK conformance testing**
An automated end-to-end conformance suite against the official A2A TCK SUT.
Conformance is verified on every release.

**Artifact management**
`--out-dir` and `--file` flags on `send`, `get`, `watch`, and `download` provide
first-class artifact saving without requiring a separate pipeline step.

---

## Planned Alignments

These items from the proposals are tracked in the issue tracker and not yet implemented:

| Item | Issue | Priority |
|---|---|---|
| `serve --proxy` | `a2ac-6x6` | 2 |
| `serve --exec` | `a2ac-7n2` | 2 |
| `send --parts` / `--json` structured input | `a2ac-79d` | 3 |
| `list tasks` filters (`--context`, `--status`, `--since`) | `a2ac-mvu` | 3 |
| `--tenant` global flag | — | — |
| `--verbose / -v` global flag | — | — |

---

## Grammar Differences from #306

#306 proposes URL as the first positional argument after the verb:
```
a2a discover <url>
a2a send <url> "message"
```

`a2acli` uses a flag instead:
```
a2acli discover --service-url <url>
a2acli send "message" --service-url <url>
```

**Rationale:** the flag approach enables named environments (`--env staging`) so users
who configure `~/.config/a2acli/config.yaml` never need to type a URL at all.
The positional approach is slightly more concise for one-off commands but doesn't
compose as well with configuration management.

If the official CLI adopts positional URLs, `a2acli` can accommodate both via
`cobra.Args` validation and a fallback to `--service-url`.
