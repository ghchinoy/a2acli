# Tier 1: Live Server Probes via `a2acli`

Tier 1 executes active black-box probe commands against a live A2A server using `a2acli`. Always pass `--output json` (or `-n`) when executing probe commands programmatically.

---

## Probe Battery & Commands

### Probe 1: Discovery & Card Well-Formedness
**Target IDs:** `A2A-CARD-001`, `A2A-CARD-002`, `A2A-CARD-003`, `A2A-CARD-005`

```bash
a2acli discover --service-url <URL> --output json
```

**Expected JSON Result:**
- Non-zero top-level fields: `name`, `version`, `supportedInterfaces`, `skills`.
- `supportedInterfaces` contains valid objects with `protocolBinding` and `url`.
- `skills` contains valid objects with `id`, `name`, `description`.

---

### Probe 2: Authenticated Extended Agent Card
**Target IDs:** `A2A-SEC-004`, `A2A-CARD-001`

```bash
a2acli discover --extended --service-url <URL> --output json
```

**Expected Result:**
- If server supports extended card: returns extended `AgentCard` JSON.
- If `capabilities.extendedAgentCard` is false: returns `UnsupportedOperationError` or `ExtendedAgentCardNotConfiguredError`.

---

### Probe 3: Built-In Conformance Smoke Check
**Target IDs:** `A2A-OPS-001`, `A2A-SEC-001`, `A2A-BIND-003`

```bash
a2acli conformance --service-url <URL> --output json
```

**Expected JSON Output:**
```json
{
  "results": [
    { "name": "AgentCard fetch", "passed": true, "message": "..." },
    { "name": "AgentCard well-formed", "passed": true, "message": "..." },
    { "name": "Auth gating", "passed": true, "skipped": false, "message": "..." },
    { "name": "Round-trip send", "passed": true, "message": "..." }
  ],
  "passed": true
}
```

---

### Probe 4: Blocking Message Send (`SendMessage`)
**Target IDs:** `A2A-OPS-001`, `A2A-OPS-002`, `A2A-TASK-001`, `A2A-TASK-004`

```bash
a2acli send "conformance probe message" --service-url <URL> --wait --output json
```

**Expected Result:**
- Valid `Task` JSON object.
- `id` is non-empty string (`A2A-TASK-001`).
- `status.state` is `TASK_STATE_COMPLETED` (or another terminal/interrupted state).

---

### Probe 5: Non-Blocking Message Send
**Target IDs:** `A2A-OPS-003`

```bash
a2acli send "conformance probe async" --service-url <URL> --immediate --output json
```

**Expected Result:**
- Immediate response containing a `Task` object.
- `status.state` is in-progress (e.g. `TASK_STATE_SUBMITTED` or `TASK_STATE_WORKING`).

---

### Probe 6: Task State & History Retrieval (`GetTask`)
**Target IDs:** `A2A-OPS-004`, `A2A-OPS-005`

```bash
# Retrieve full task state
a2acli get <TASK_ID> --service-url <URL> --output json
```

**Expected Result:**
- Returns task matching `<TASK_ID>`.
- History and status reflect latest execution.

---

### Probe 7: Task Cancellation (`CancelTask`)
**Target IDs:** `A2A-OPS-009`, `A2A-OPS-010`

1. Start a long-running or immediate task to obtain `<TASK_ID>`.
2. Cancel active task:
   ```bash
   a2acli cancel <TASK_ID> --service-url <URL> --output json
   ```
3. Attempt to cancel an already completed task:
   ```bash
   a2acli cancel <COMPLETED_TASK_ID> --service-url <URL> --output json
   ```
   **Expected:** Returns error matching `TaskNotCancelableError` (HTTP 400 / gRPC `FAILED_PRECONDITION`).

---

### Probe 8: Task History Listing (`ListTasks`)
**Target IDs:** `A2A-OPS-006`, `A2A-OPS-007`, `A2A-OPS-008`

```bash
a2acli list tasks --service-url <URL> --limit 5 --output json
```

**Expected Result:**
- `tasks` array sorted descending by timestamp.
- `nextPageToken` string field present (empty string on last page).
- `artifacts` field omitted when `includeArtifacts` is not specified.

---

### Probe 9: Multi-Transport Equivalence Verification
**Target IDs:** `A2A-BIND-001`

Execute `send` across all transports advertised in `supportedInterfaces`:

```bash
# Test JSON-RPC transport
a2acli send "ping" --service-url <URL> --transport jsonrpc --wait --output json

# Test REST transport
a2acli send "ping" --service-url <URL> --transport rest --wait --output json

# Test gRPC transport
a2acli send "ping" --service-url <URL> --transport grpc --wait --output json
```

**Expected Result:**
- All supported transports return semantically equivalent `Task` outputs.

---

### Probe 10: Protocol Version Negotiation
**Target IDs:** `A2A-VER-001`, `A2A-VER-002`, `A2A-VER-003`

```bash
# Explicit v1.0
a2acli send "test" --service-url <URL> --protocol 1.0.0 --wait --output json

# Legacy v0.3
a2acli send "test" --service-url <URL> --protocol 0.3.0 --wait --output json

# Invalid version (expect VersionNotSupportedError)
a2acli send "test" --service-url <URL> --svc-param A2A-Version=99.0 --output json
```

---

### Probe 11: Push Notification Configurations (Optional)
**Target IDs:** `A2A-PUSH-001`, `A2A-PUSH-002`, `A2A-PUSH-004`

```bash
a2acli push-config create <TASK_ID> http://127.0.0.1:9999/webhook --service-url <URL> --output json
a2acli push-config list <TASK_ID> --service-url <URL> --output json
a2acli push-config delete <TASK_ID> <CONFIG_ID> --service-url <URL> --output json
```

---

### Probe 12: A2UI Extension Conformance (Optional)
**Target IDs:** `A2A-EXT-003`

If agent advertises A2UI extension (`https://a2ui.org/a2a-extension/a2ui/v1.0`):

```bash
a2acli a2ui validate --service-url <URL> --output json
```
