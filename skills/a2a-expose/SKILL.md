---
name: a2a-expose
description: Guide developers and coding agents in adding an A2A Protocol v1.0 exposure layer to an existing API, microservice, or codebase. Enforces a design-first workflow (intent-based skill modeling, task lifecycle, artifact delivery, and security mapping) before emitting Go or Python SDK code, followed by a verification battery. Use when asked to add A2A support to a service, expose an API via A2A, build an A2A agent, or turn code into an A2A service.
license: Apache-2.0
compatibility: Requires a2acli binary to run the verification battery. Works with Go (a2a-go v2) and Python (a2a-sdk >=1.0).
---

# A2A Exposure Skill

Guide the process of adding a compliant, discoverable Agent-to-Agent (A2A) Protocol v1.0 exposure to any existing service, API, or codebase.

## Core Rule: Design First, Then Code

**Do NOT start by writing code or mapping REST endpoints 1:1.** You MUST perform an architectural exposure design pass, complete the design proposal template (`assets/exposure-design.template.md`), and obtain user sign-off BEFORE implementing server handlers.

## Reference Manifest (Progressive Disclosure)

Load these reference documents during each phase of exposure design and implementation:

- [references/design-worksheet.md](references/design-worksheet.md) — The 10-question architectural decision tree (skill granularity, Message vs Task, artifacts, streaming, task store, auth, transports).
- [references/agent-card.md](references/agent-card.md) — Agent Card v1.0 canonical JSON schema, skill authoring, and discoverability rules.
- [references/task-lifecycle.md](references/task-lifecycle.md) — Task state machine, status events, artifact delivery patterns, and multi-turn state.
- [references/auth.md](references/auth.md) — Security scheme declarations (`bearerAuth`, `apiKeyAuth`, `oauth2`) and in-task authorization (§7.6).
- [references/impl-go.md](references/impl-go.md) — Exact, verified Go implementation guide (`github.com/a2aproject/a2a-go/v2` v2.3.1).
- [references/impl-python.md](references/impl-python.md) — Python implementation guide (`a2a-sdk` >=1.0 with FastAPI/Starlette).
- [references/anti-patterns.md](references/anti-patterns.md) — 14 common server failure modes with spec citations and fixes.
- [references/verify.md](references/verify.md) — Verification battery workflow using `a2acli`.

---

## Step-by-Step Exposure Workflow

### Phase 1: Architectural Design & Modeling
1. Analyze the target service, codebase, or API endpoints.
2. Review the 10 modeling questions in [references/design-worksheet.md](references/design-worksheet.md).
3. Group granular API endpoints into high-level *user intent* skills.
4. Decide on interaction mode (`Task` vs `Message`), output artifacts, streaming, task store, and authentication.
5. Fill out `assets/exposure-design.template.md` with your proposed design and Agent Card JSON preview.
6. Present the design to the user for sign-off.

### Phase 2: Implementation
Upon user approval, implement the server layer using the appropriate SDK reference:
- **Go (`a2a-go/v2` v2.3.1):** Follow [references/impl-go.md](references/impl-go.md). Implement `a2asrv.AgentExecutor`, create `AgentCard`, build `RequestHandler` with options (`WithTaskStore`, `WithCapabilityChecks`), and mount HTTP mux at `/` and `/.well-known/agent-card.json`.
- **Python (`a2a-sdk` >=1.0):** Follow [references/impl-python.md](references/impl-python.md). Check installed SDK version, implement `AgentExecutor`, create `AgentCard`, build `DefaultRequestHandlerV2`, and mount sub-app at `/`.

### Phase 3: Avoid Anti-Patterns
Before compiling, cross-check code against [references/anti-patterns.md](references/anti-patterns.md):
- Ensure `Execute` always yields a terminal status (`TASK_STATE_COMPLETED` or `TASK_STATE_FAILED`).
- Deliver outputs as `Artifacts`, not inline message text (§3.7).
- Verify `AgentCard.capabilities` matches handler capabilities exactly.
- Confirm card is served at `/.well-known/agent-card.json`.

### Phase 4: Run Verification Battery
Build and launch the server locally, then run the verification battery:

```bash
# Automated verification script
./skills/a2a-expose/scripts/verify-exposure.sh http://127.0.0.1:9001

# Or run manual a2acli verification suite per references/verify.md
a2acli discover --service-url http://127.0.0.1:9001 --output json
a2acli conformance --service-url http://127.0.0.1:9001 --output json
a2acli send "Verification test" --service-url http://127.0.0.1:9001 --wait --output json
```

### Phase 5: Conformance Check (Optional Cross-Validation)
Run `a2a-conformance` skill against the newly exposed endpoint to verify 100% MUST requirement compliance.

---

## Mandatory Exposure Checklist

Ensure:
- [ ] User signed off on the Exposure Design Proposal before code was written.
- [ ] Skills represent high-level user intents, not 1:1 REST endpoints.
- [ ] Outputs are delivered as named `Artifacts` attached to the Task (§3.7).
- [ ] Agent Card is served at `/.well-known/agent-card.json`.
- [ ] All verification battery checks passed with `a2acli`.
