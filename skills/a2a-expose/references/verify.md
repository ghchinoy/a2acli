# Verifying an A2A Exposure with `a2acli`

After implementing an A2A server, execute this verification battery using `a2acli` to ensure functional correctness and spec compliance before requesting user sign-off.

---

## The Verification Battery

Set `SERVICE_URL` to your local endpoint (e.g. `http://127.0.0.1:9001`).

### Step 1: Verify Agent Card Discovery
```bash
a2acli discover --service-url "$SERVICE_URL" --output json
```
**Assertion:** Returns valid JSON. Contains `name`, `version: "1.0"`, `supportedInterfaces`, and `skills`.

---

### Step 2: Run Built-In Smoke Test
```bash
a2acli conformance --service-url "$SERVICE_URL" --output json
```
**Assertion:** `passed` field is `true`. All checks (`AgentCard fetch`, `AgentCard well-formed`, `Auth gating`, `Round-trip send`) report `passed: true`.

---

### Step 3: Test Synchronous / Blocking Task Execution
```bash
a2acli send "Test prompt" --service-url "$SERVICE_URL" --wait --output json
```
**Assertion:** Returns a `Task` object. `status.state` is `TASK_STATE_COMPLETED`.

---

### Step 4: Test Non-Blocking Asynchronous Send
```bash
a2acli send "Async prompt" --service-url "$SERVICE_URL" --immediate --output json
```
**Assertion:** Returns immediately with a `Task` in `TASK_STATE_SUBMITTED` or `TASK_STATE_WORKING`. Note the returned `id` as `<TASK_ID>`.

---

### Step 5: Test Task Retrieval (`GetTask`)
```bash
a2acli get <TASK_ID> --service-url "$SERVICE_URL" --output json
```
**Assertion:** Returns current state of `<TASK_ID>`.

---

### Step 6: Test Task Listing (`ListTasks`)
```bash
a2acli list tasks --service-url "$SERVICE_URL" --limit 5 --output json
```
**Assertion:** Returns list of tasks sorted newest-first. `nextPageToken` is present.

---

### Step 7: Test Cancellation (`CancelTask`)
```bash
# Start a task and cancel it
TASK_ID=$(a2acli send "Long task" --service-url "$SERVICE_URL" --immediate --output json | jq -r '.id')
a2acli cancel "$TASK_ID" --service-url "$SERVICE_URL" --output json
```
**Assertion:** Task state updates to `TASK_STATE_CANCELED`.

---

### Step 8: Multi-Transport Equivalence Verification
For every transport declared in `supportedInterfaces`, run:

```bash
# JSON-RPC
a2acli send "hello" --service-url "$SERVICE_URL" --transport jsonrpc --wait --output json

# HTTP+JSON (REST)
a2acli send "hello" --service-url "$SERVICE_URL" --transport rest --wait --output json

# gRPC (if enabled)
a2acli send "hello" --service-url "$SERVICE_URL" --transport grpc --wait --output json
```

**Assertion:** All transports complete successfully and return semantically equivalent outputs.

---

### Step 9: Run Automated Verification Script
Execute the bundled verification script:

```bash
./skills/a2a-expose/scripts/verify-exposure.sh "$SERVICE_URL"
```
