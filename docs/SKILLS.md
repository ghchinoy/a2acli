# Agent Skills Guide (`docs/SKILLS.md`)

This guide explains how AI coding agents (Claude Code, OpenCode, Cursor, Gemini CLI, Copilot CLI) use the skills shipped in `a2acli` to build, evaluate, and operate Agent-to-Agent (A2A) protocol services.

---

## 1. What Agent Skills Are

An **agent skill** is an [`agentskills.io`](https://agentskills.io/) compliant package that teaches AI coding agents how to perform specialized domain tasks deterministically and safely.

Skills use a **progressive disclosure model**:

```
1. Startup Index   ──> Agent reads name + description (~100 tokens per skill)
2. Activation      ──> Agent loads full SKILL.md body (<500 lines) when triggered
3. On-Demand Resources ──> Agent reads references/, scripts/, or assets/ as needed
```

Because skill descriptions govern when an agent activates a skill, description wording is optimized with specific trigger keywords ("Use when asked to...").

---

## 2. The Three Skills at a Glance

| Skill | Purpose | Key Trigger Words | Main Artifacts | Needs Live Server? |
|---|---|---|---|---|
| [**`a2acli`**](../skills/a2acli/SKILL.md) | Drive `a2acli` CLI to discover, invoke, and monitor A2A agents | `a2acli`, `discover agent`, `send task`, `stream results`, `a2a auth` | CLI command executions, JSON/NDJSON outputs | Yes |
| [**`a2a-expose`**](../skills/a2a-expose/SKILL.md) | Add an A2A v1.0 exposure layer to an API or service | `expose api`, `add a2a support`, `build a2a agent`, `turn code into a2a` | `exposure-design.md`, `agent-card.json`, SDK server code | No (builds one) |
| [**`a2a-conformance`**](../skills/a2a-conformance/SKILL.md) | Audit an A2A server for spec compliance, security, and quality | `audit a2a agent`, `conformance check`, `a2a spec compliance`, `test a2a agent` | Conformance report (`A2A-*` IDs, verdict) | Yes (or card JSON) |

---

## 3. Installation & Activation

### Option A: Open Agent Skills CLI (`npx skills`)
Recommended for OpenCode, Claude Code, Cursor, Cline, and 70+ supported agents:

```bash
# Discover skills in the repository
npx skills add ghchinoy/a2acli --list

# Install specific skill to current project (saves in ./<agent>/skills/)
npx skills add ghchinoy/a2acli --skill a2a-expose

# Install all skills globally across all projects
npx skills add ghchinoy/a2acli --all -g
```

### Option B: Gemini CLI (`gemini skills`)
```bash
# Install specific skill globally
gemini skills install https://github.com/ghchinoy/a2acli.git --path skills/a2a-conformance

# Or link from a local clone
git clone https://github.com/ghchinoy/a2acli.git
cd a2acli && gemini skills link
```

### Option C: Manual Copy
Copy skill folders directly into your workspace's skill directory:
- **OpenCode:** `.opencode/skill/`
- **Claude Code:** `.claude/skills/`
- **General / Antigravity:** `.agents/skills/` or `~/.agents/skills/`

---

## 4. `a2acli` Skill Deep Dive

The `a2acli` skill teaches coding agents how to run `a2acli` in non-interactive environments.

### Core Determinism Rules
When driving `a2acli`, agents MUST adhere to two critical rules:
1. **Always pass `--output json` (or `-n`):** Disables the interactive Bubble Tea TUI and emits parseable NDJSON/JSON.
2. **Always pass `--wait` with `send`:** Blocks execution until task reaches a terminal state (`TASK_STATE_COMPLETED` or `TASK_STATE_FAILED`).

### Capability Summary
- Agent card discovery (`a2acli discover --extended`)
- Blocking and streaming message send (`a2acli send`)
- Task state inspection & cancellation (`a2acli get`, `a2acli cancel`)
- Historical task listing with pagination (`a2acli list tasks`)
- PKCE OAuth 2.1 token management (`a2acli auth login/token/status`)
- Local mock server instantiation (`a2acli serve --echo`)

---

## 5. `a2a-expose` Skill Deep Dive

The `a2a-expose` skill guides developers and agents through creating a compliant A2A exposure layer for an existing codebase or API.

### The Design-First Gate
To prevent mechanical 1:1 mappings of REST endpoints to A2A skills, the skill enforces a mandatory design pass before code generation:
1. Agent completes `exposure-design.template.md` using the 10 modeling questions in `references/design-worksheet.md`.
2. Agent presents the design proposal and Agent Card JSON preview to the user.
3. Code generation proceeds ONLY after explicit user sign-off.

### The 10 Modeling Questions Summary
1. **Skill Granularity:** Group REST endpoints into intent-based skills.
2. **Message vs Task:** Sub-second sync -> `Message`; >2s or streaming -> `Task`.
3. **Artifact Delivery:** Per §3.7, outputs belong in `Artifacts`, not inline status message text.
4. **Streaming:** Only declare `capabilities.streaming = true` if handlers yield intermediate events.
5. **Multi-Turn State:** Choose `InMemory` or DB-backed task store (`PostgreSQL`/`SQLite`).
6. **Auth Mapping:** Map existing API keys/JWTs to `AgentCard.securitySchemes`.
7. **Push Notifications:** Require `ConfigStore` + `PushSender` if declared.
8. **Transport Bindings:** Default to `JSONRPC` at `/invoke`; ensure §5.1 functional equivalence if adding REST/gRPC.
9. **Modalities:** Declare default and skill-level `inputModes` and `outputModes`.
10. **Discoverability:** Include intent-rich descriptions, tags, and concrete user examples.

### SDK Coverage
- **Go (`a2a-go/v2` v2.4.0):** Full implementation guide in `references/impl-go.md`.
- **Python (`a2a-sdk` >=1.0):** Full implementation guide in `references/impl-python.md`.

---

## 6. `a2a-conformance` Skill Deep Dive

The `a2a-conformance` skill evaluates A2A servers against normative specification requirements and generates structured audit reports.

### Multi-Tier Assessment Strategy

```
Tier 0 (Static)  ──> Inspect Agent Card JSON & SDK source code (No server required)
Tier 1 (Probes)  ──> Run automated a2acli black-box probe battery against live URL
Tier 2 (TCK)     ──> Orchestrate official a2aproject/a2a-tck pytest compliance suite
```

### Stable Check ID Schema
Checks are assigned stable IDs (`A2A-<AREA>-<NNN>`):
- `A2A-CARD-*` (Agent Card & Discovery)
- `A2A-OPS-*` (Core Operations)
- `A2A-TASK-*` (Task Lifecycle)
- `A2A-STREAM-*` (Streaming Event Delivery)
- `A2A-PUSH-*` (Push Notifications)
- `A2A-BIND-*` (Protocol Bindings)
- `A2A-VER-*` (Versioning)
- `A2A-SEC-*` (Security & Authorization)
- `A2A-EXT-*` (Extensions)
- `A2A-QUAL-*` (Quality & Design - Non-Normative)

### Verdict Model
- **`CONFORMANT`**: 100% MUST checks pass, 0 MUST failures.
- **`CONFORMANT WITH WARNINGS`**: 100% MUST checks pass, but SHOULD warnings exist.
- **`NOT CONFORMANT`**: 1 or more MUST checks fail.

### Real Conformance Report Excerpt (from TCK SUT Run)

```markdown
# A2A Protocol Conformance Evaluation Report
**Target Service:** `http://127.0.0.1:9999`
**Agent Name:** `TCK Core Agent`

## Summary
- **Overall Verdict:** **CONFORMANT WITH WARNINGS**
- **MUST Checks Passed:** 24 / 24
- **SHOULD Checks Passed:** 11 / 12

## Detailed Findings
| Check ID | Requirement | Severity | Result | Finding |
|---|---|---|---|---|
| `A2A-CARD-001` | Card served at well-known path | MUST | PASS | Card fetched at `/.well-known/agent-card.json`. |
| `A2A-CARD-006` | Card response includes Cache-Control | SHOULD | WARN | `Cache-Control` header missing from response. |
| `A2A-OPS-001` | SendMessage returns immediately | MUST | PASS | Returned valid Task in 42ms. |
| `A2A-BIND-001` | Multi-transport equivalence | MUST | PASS | Equivalent responses across JSON-RPC and gRPC. |
```

---

## 7. Composing the Skills Across the Lifecycle

The three skills work together across the software development lifecycle:

![A2A Skills Workflow](img/skills-workflow.webp)

1. **Design & Implementation (`a2a-expose`):** Create proposal → user sign-off → generate SDK code.
2. **Local Verification (`a2a-expose`):** Run `./skills/a2a-expose/scripts/verify-exposure.sh`.
3. **Audit & Compliance (`a2a-conformance`):** Run static review + Tier 1/2 probes → emit conformance report.
4. **Production Operation (`a2acli`):** Discover capabilities, authenticate, and send tasks in production.

---

## 8. Worked Scenarios

### Scenario A: Expose an Existing REST API over A2A
**Situation:** You have a Python/FastAPI service with 12 endpoints at `./src/api`.
**User Prompt:** `"I want to expose my document processing API at ./src/api as an A2A service."`
**Execution:**
1. `a2a-expose` skill activates.
2. Agent reads `./src/api` and groups 12 endpoints into 2 intent skills (`analyze_document`, `extract_tables`).
3. Agent emits `exposure-design.md` and requests user approval.
4. Upon sign-off, agent writes FastAPI A2A handler using `a2a-sdk` >=1.0.
5. Agent runs `verify-exposure.sh http://localhost:9001` to verify.

### Scenario B: Audit a Third-Party Agent Pre-Integration
**Situation:** Vendor provides an A2A endpoint at `https://agent.partner.com`.
**User Prompt:** `"Audit https://agent.partner.com before we connect our orchestrator to it."`
**Execution:**
1. `a2a-conformance` skill activates.
2. Agent runs Tier 0 card review (`A2A-CARD-*`).
3. Agent executes `./skills/a2a-conformance/scripts/probe.sh https://agent.partner.com`.
4. Agent generates `report.md` with `CONFORMANT WITH WARNINGS` verdict and top 3 remediation items.

### Scenario C: CI Conformance Gate
**Situation:** Maintainers want to prevent spec regressions in pull requests.
**User Prompt:** `"Set up a GitHub Actions workflow to run A2A conformance checks on every PR."`
**Execution:**
1. Agent inspects `.github/workflows/conformance.yml`.
2. Agent adds step calling `./skills/a2a-conformance/scripts/run-tck.sh http://127.0.0.1:9001 --level must`.
3. Build fails if any MUST check regresses.

### Scenario D: Debug a Hanging Event Stream
**Situation:** Client hangs indefinitely when calling an A2A agent's streaming endpoint.
**User Prompt:** `"Our A2A streaming client hangs on task completion when calling http://localhost:9001."`
**Execution:**
1. `a2a-conformance` runs probe 4 (`SendStreamingMessage`).
2. Isolates `A2A-STREAM-002` failure (missing terminal status event).
3. Cross-references `skills/a2a-expose/references/anti-patterns.md` #2 ("Missing Terminal State Event").
4. Fixes server executor code by yielding `TaskStateCompleted`.

### Scenario E: Day-to-Day CLI Operation
**Situation:** Developer wants to query a remote agent from the command line.
**User Prompt:** `"Discover skills on http://localhost:9001 and run the summarize skill."`
**Execution:**
1. `a2acli` skill activates.
2. Agent runs `a2acli discover -u http://localhost:9001 --output json`.
3. Agent inspects card, finds skill `summarize`, and runs:
   `a2acli send "Summarize this text" -u http://localhost:9001 --skill summarize --wait --output json`.

---

## 9. Prompt Library

Use these prompt templates to trigger specific skills reliably:

| Intended Goal | Recommended Prompt | Activated Skill |
|---|---|---|
| Discover Agent | `"Discover capabilities and skills for agent at http://localhost:9001"` | `a2acli` |
| Send Task | `"Send 'Summarize Q3 report' to http://localhost:9001 and wait for result"` | `a2acli` |
| OAuth Login | `"Log in to OAuth-protected A2A agent at https://agent.example.com"` | `a2acli` |
| Expose Service | `"Add an A2A exposure layer to our API in ./src/api"` | `a2a-expose` |
| Design Proposal | `"Create an A2A exposure design proposal for this codebase"` | `a2a-expose` |
| Conformance Audit | `"Audit http://localhost:9001 against the A2A v1.0 specification"` | `a2a-conformance` |
| Run TCK | `"Run official A2A TCK conformance tests against http://localhost:9001"` | `a2a-conformance` |

### Explicit Fallback Prompts
If an agent fails to auto-activate a skill, use explicit skill names:
- `"Use skill a2a-expose to design an A2A wrapper for ./service"`
- `"Use skill a2a-conformance to audit http://localhost:9001"`
- `"Use skill a2acli to query http://localhost:9001"`

---

## 10. Troubleshooting

| Issue | Root Cause | Resolution |
|---|---|---|
| Skill didn't activate | Prompt lacked trigger keywords | Use explicit fallback prompt (`"Use skill <name>..."`) |
| `a2acli` command hangs | Missing `--output json` or `--wait` | Ensure `--output json` is passed in non-TTY environments |
| `a2acli` binary not found | Binary not compiled or missing from PATH | Run `make build` inside `/workspace/a2acli` |
| Tier 2 TCK fails to run | Missing `uv` or Python 3.11+ | Install `uv` via `curl -LsSf https://astral.sh/uv/install.sh \| sh` |
| Conformance probes skipped | Server requires authentication | Provide token via `--token` or `auth login` |

---

## 11. Authoring a New A2A Skill

To author a new skill in this repository, adhere to these guidelines:

### Directory Structure & Naming
- Directory name MUST be kebab-case (e.g., `skills/a2a-my-skill`).
- Frontmatter `name` MUST match directory name exactly.

### Frontmatter Conventions (`SKILL.md`)
```yaml
---
name: a2a-my-skill
description: Concise description of what the skill does (max 1024 chars). Always end with a "Use when asked to..." trigger clause.
license: Apache-2.0
compatibility: Specify environment or binary requirements.
---
```

### Progressive Disclosure Rules
- Keep `SKILL.md` under 500 lines (~5,000 tokens).
- Put detailed reference documentation in `references/<name>.md`.
- Put templates/scaffolds in `assets/`.
- Put executable POSIX sh scripts in `scripts/`.
- POSIX scripts MUST handle optional tools gracefully (e.g. `command -v`).

### Registration
Register new skills in:
1. `README.md` under `## Agent Skills`
2. `skills/README.md`
3. This guide (`docs/SKILLS.md`)
