# Tier 0: Static Agent Card & Source Code Review

Static review evaluates an A2A agent's declared metadata (Agent Card JSON) and source code (if accessible) without requiring a running server instance or active network calls.

---

## 1. Agent Card Verification (`A2A-CARD-*`)

Obtain the Agent Card JSON either from source code, local file, or direct HTTP fetch.

### Checklist

- [ ] **`A2A-CARD-002` Required Top-Level Keys:**
  Ensure all of the following keys exist and are non-empty:
  - `name` (string)
  - `version` (string, e.g. `"1.0"`)
  - `description` (string)
  - `supportedInterfaces` (array, ≥1 item)
  - `defaultInputModes` (array, e.g. `["text"]`)
  - `defaultOutputModes` (array, e.g. `["text"]`)
  - `capabilities` (object)
  - `skills` (array, ≥1 item)

- [ ] **`A2A-CARD-003` Supported Interfaces:**
  Each object in `supportedInterfaces` MUST contain:
  - `url` (string URL or `host:port`)
  - `protocolBinding` (string: `"JSONRPC"`, `"GRPC"`, or `"HTTP+JSON"`)
  - `protocolVersion` (string: `"1.0"` or `"0.3"`)

- [ ] **`A2A-CARD-005` Skill Definitions:**
  Each item in `skills[]` MUST contain:
  - `id` (string, kebab-case or snake_case ID)
  - `name` (string)
  - `description` (string)
  - `tags` (array of strings)
  *Recommended:* `examples` (array of sample prompts).

- [ ] **`A2A-VER-004` Version String Format:**
  `version` MUST NOT contain patch versions (e.g., use `"1.0"`, not `"1.0.4"`).

---

## 2. Source Code Pattern Audit (Go / Python)

If source code is available, inspect initialization and handler wiring against known SDK anti-patterns.

### Go (`a2a-go/v2`) Checks
1. **Executor Interface:**
   Verify `a2asrv.AgentExecutor` is implemented with `Execute` and `Cancel` returning `iter.Seq2[a2a.Event, error]`.
2. **Terminal State Yield:**
   Ensure `Execute` yields a status update event with a terminal state (`TaskStateCompleted`, `TaskStateFailed`, etc.) on all code paths.
3. **Card Path Registration:**
   Confirm card is mounted at `a2asrv.WellKnownAgentCardPath` (`/.well-known/agent-card.json`).
4. **Transport Equivalence:**
   If multiple transports are wired (e.g. REST and JSON-RPC), ensure both wrap the *same* `a2asrv.RequestHandler` instance.

### Python (`a2a-sdk` >=1.0) Checks
1. **AgentExecutor:**
   Confirm `AgentExecutor.execute` handles exceptions and enqueues a `TaskStatusUpdateEvent` with `TaskState.FAILED` on error.
2. **Event Queue Drain:**
   Ensure task generator yields `TaskStatusUpdateEvent` with `completed` state before terminating.
3. **Route Mounting:**
   Verify card endpoint is mounted at `/.well-known/agent-card.json`.

---

## 3. Tier 0 Reporting Output

When running static review, record pass/fail status for all `A2A-CARD-*` IDs and append findings under the **Tier 0 Static Analysis** section of the report.
