# A2A Protocol v1.0 — Server Conformance Checklist

This reference document defines the complete normative conformance requirements for an A2A Protocol v1.0 remote agent (server) implementation. Checks are organized by functional area with stable IDs (`A2A-<AREA>-<NNN>`).

---

## Areas & Check ID Schema

| Area Prefix | Functional Area | Spec Sections |
|---|---|---|
| `A2A-CARD` | Agent Card & Discovery | §8, §4.4, §14.3 |
| `A2A-OPS` | Core Protocol Operations | §3.1, §3.2, §3.3 |
| `A2A-TASK` | Task Lifecycle & State Transitions | §3.4, §4.1, §3.7 |
| `A2A-STREAM` | Streaming Event Delivery | §3.1.2, §3.5.2, §11.7 |
| `A2A-PUSH` | Push Notifications | §3.1.7-10, §3.5.3, §4.3, §13.2 |
| `A2A-BIND` | Protocol Bindings & Equivalence | §5, §9, §10, §11, §12 |
| `A2A-VER` | Protocol Versioning & Negotiation | §3.6, §14.2.1 |
| `A2A-SEC` | Security, Auth & Authorization | §7, §13, §3.3.2 |
| `A2A-EXT` | Protocol Extensions | §4.6, §3.3.4 |
| `A2A-QUAL` | Quality, DX & Design (Non-Normative) | Best Practices |

### Conformance Tiers
- **Tier 0 (Static):** Inspected directly from the Agent Card JSON or source code.
- **Tier 1 (Probe):** Verified via `a2acli` commands (`discover`, `conformance`, `send`, `get`, `cancel`, `list tasks`, `push-config`, `a2ui validate`).
- **Tier 2 (TCK):** Verified via the official Technology Compatibility Kit (`a2aproject/a2a-tck`).

---

## 1. Agent Card & Discovery (`A2A-CARD`)

| ID | Level | Tier | Spec § | Requirement | Verification Method |
|---|---|---|---|---|---|
| `A2A-CARD-001` | **MUST** | Tier 1 | §8.1, §14.3 | Agent Card MUST be served at `/.well-known/agent-card.json`. | `a2acli discover -u <url> --output json` |
| `A2A-CARD-002` | **MUST** | Tier 0 | §4.4.1 | Agent Card MUST contain required fields: `name`, `version`, `description`, `supportedInterfaces`, `defaultInputModes`, `defaultOutputModes`, `capabilities`, `skills`. | Inspect card JSON for required non-empty keys |
| `A2A-CARD-003` | **MUST** | Tier 0 | §8.3 | `supportedInterfaces` MUST declare at least one protocol binding with transport protocol and URL. | Verify `supportedInterfaces` is non-empty list |
| `A2A-CARD-004` | **SHOULD** | Tier 0 | §8.3.1 | `supportedInterfaces` SHOULD list protocols in order of server preference (first = preferred). | Inspect interface order |
| `A2A-CARD-005` | **MUST** | Tier 0 | §4.4.6 | Every skill in `skills` MUST declare `id`, `name`, `description`, `tags`. | Inspect `skills[]` array |
| `A2A-CARD-006` | **SHOULD** | Tier 1 | §8.6.1 | Agent Card HTTP endpoint SHOULD return `Cache-Control` header with `max-age`. | `curl -i <url>/.well-known/agent-card.json` |
| `A2A-CARD-007` | **SHOULD** | Tier 1 | §8.6.1 | Agent Card HTTP endpoint SHOULD return an `ETag` header derived from card version or content hash. | Inspect HTTP headers for `ETag` |
| `A2A-CARD-008` | **MUST** | Tier 0 | §8.4.1 | If JWS signed, Agent Card JSON MUST be canonicalized using JCS (RFC 8785) pre-signature, excluding `signatures`. | Verify JCS signature format |

---

## 2. Core Protocol Operations (`A2A-OPS`)

| ID | Level | Tier | Spec § | Requirement | Verification Method |
|---|---|---|---|---|---|
| `A2A-OPS-001` | **MUST** | Tier 1 | §3.1.1 | `SendMessage` MUST return immediately with either a `Task` or a direct `Message`. | `a2acli send "test" -u <url> --output json --wait` |
| `A2A-OPS-002` | **MUST** | Tier 1 | §3.2.2 | In blocking mode (`returnImmediately: false`), server MUST wait until task reaches terminal or interrupted state before returning response. | `a2acli send "test" -u <url> -w --output json` |
| `A2A-OPS-003` | **MUST** | Tier 1 | §3.2.2 | In non-blocking mode (`returnImmediately: true`), server MUST return task status immediately after creation. | `a2acli send "test" -u <url> --immediate --output json` |
| `A2A-OPS-004` | **MUST** | Tier 1 | §3.1.3 | `GetTask` MUST return current task state, status, and artifacts for a valid `taskId`. | `a2acli get <taskId> -u <url> --output json` |
| `A2A-OPS-005` | **MUST** | Tier 1 | §3.2.4 | `GetTask` with `historyLength: 0` MUST omit history messages. | `a2acli get <taskId> -u <url> --output json` (check `history`) |
| `A2A-OPS-006` | **MUST** | Tier 1 | §3.1.4 | `ListTasks` MUST implement cursor-based pagination (`pageToken`, `nextPageToken`) and set `nextPageToken: ""` on final page. | `a2acli list tasks -u <url> --limit 10 --output json` |
| `A2A-OPS-007` | **MUST** | Tier 1 | §3.1.4 | `ListTasks` MUST sort tasks by status timestamp descending (newest first). | Inspect timestamp order in list output |
| `A2A-OPS-008` | **MUST** | Tier 1 | §3.1.4 | When `includeArtifacts` is false, `artifacts` MUST be omitted entirely from `ListTasks` items. | Check `list tasks` output without artifact flag |
| `A2A-OPS-009` | **MUST** | Tier 1 | §3.1.5 | `CancelTask` MUST return `TaskNotCancelableError` (HTTP 400 / gRPC `FAILED_PRECONDITION`) if task is already terminal. | `a2acli cancel <completedTaskId> -u <url> --output json` |
| `A2A-OPS-010` | **MUST** | Tier 1 | §3.3.1 | `CancelTask` MUST be idempotent. | Repeat `cancel` on active task |
| `A2A-OPS-011` | **MUST** | Tier 1 | §3.3.2 | Server MUST validate input parameters and return structured error with `code` and human-readable `message`. | Send invalid JSON / missing fields |

---

## 3. Task Lifecycle & State Transitions (`A2A-TASK`)

| ID | Level | Tier | Spec § | Requirement | Verification Method |
|---|---|---|---|---|---|
| `A2A-TASK-001` | **MUST** | Tier 1 | §3.4.2 | Server MUST generate a globally unique `taskId` for each newly created task. | Compare task IDs across multiple invocations |
| `A2A-TASK-002` | **MUST** | Tier 1 | §3.4.1 | If server accepts/generates a `contextId`, it MUST return it in Task/Message and maintain it across multi-turn turns. | Send task with `--task` or check returned `contextId` |
| `A2A-TASK-003` | **MUST** | Tier 1 | §3.4.1 | Server MUST reject requests with mismatching `contextId` and `taskId`. | Pass invalid context ID for an existing task |
| `A2A-TASK-004` | **MUST** | Tier 1 | §4.1.3 | Task state MUST be one of valid `TaskState` enum values: `TASK_STATE_SUBMITTED`, `WORKING`, `COMPLETED`, `FAILED`, `CANCELED`, `REJECTED`, `INPUT_REQUIRED`, `AUTH_REQUIRED`. | Inspect `status.state` string value |
| `A2A-TASK-005` | **SHOULD** | Tier 1 | §3.7 | Task outputs SHOULD be delivered as `Artifacts` rather than inline `Message` text. | Check if task produces named artifacts |
| `A2A-TASK-006` | **MUST** | Tier 0 | §4.1.6 | Each `Part` in a Message or Artifact MUST contain exactly one of `text`, `raw`, `url`, `data`. | Validate Part JSON structure |

---

## 4. Streaming Event Delivery (`A2A-STREAM`)

| ID | Level | Tier | Spec § | Requirement | Verification Method |
|---|---|---|---|---|---|
| `A2A-STREAM-001` | **MUST** | Tier 1 | §3.1.2 | If streaming is supported, `SendStreamingMessage` MUST begin with a `Task` object (or single `Message`), followed by `TaskStatusUpdateEvent`/`TaskArtifactUpdateEvent`. | `a2acli send "test" -u <url> --output json` (observe stream) |
| `A2A-STREAM-002` | **MUST** | Tier 1 | §3.1.2, §3.1.6 | Streaming connection MUST terminate when task reaches a terminal state (`COMPLETED`, `FAILED`, `CANCELED`, `REJECTED`). | Monitor stream closing on task completion |
| `A2A-STREAM-003` | **MUST** | Tier 1 | §3.5.2 | Events MUST be delivered in strict generation order without reordering. | Verify sequence of status and artifact events |
| `A2A-STREAM-004` | **MUST** | Tier 1 | §3.1.6 | `SubscribeToTask` MUST emit current `Task` snapshot as first event upon client subscription. | `a2acli subscribe <taskId> -u <url> --output json` |
| `A2A-STREAM-005` | **MUST** | Tier 1 | §3.3.4 | If `capabilities.streaming` is false/absent, `SendStreamingMessage` / `SubscribeToTask` MUST return `UnsupportedOperationError`. | Test streaming endpoints on non-streaming server |

---

## 5. Push Notifications (`A2A-PUSH`)

| ID | Level | Tier | Spec § | Requirement | Verification Method |
|---|---|---|---|---|---|
| `A2A-PUSH-001` | **MUST** | Tier 1 | §3.1.7 | If push is supported, `CreateTaskPushConfig` MUST register a webhook URL for task updates. | `a2acli push-config create <taskId> <url> -u <url> --output json` |
| `A2A-PUSH-002` | **MUST** | Tier 1 | §3.1.8-10 | `GetTaskPushConfig`, `ListTaskPushConfigs`, `DeleteTaskPushConfig` MUST perform expected CRUD operations. | `a2acli push-config list/get/delete ...` |
| `A2A-PUSH-003` | **MUST** | Tier 1 | §4.3.3 | Webhook POST requests MUST carry `Content-Type: application/a2a+json` and include authentication credentials per config. | Inspect HTTP webhook receiver logs |
| `A2A-PUSH-004` | **MUST** | Tier 1 | §3.3.4 | If `capabilities.pushNotifications` is false/absent, all push config calls MUST return `PushNotificationNotSupportedError`. | Call push-config endpoints when push disabled |
| `A2A-PUSH-005` | **SHOULD** | Tier 2 | §13.2 | Webhook client SHOULD implement exponential backoff retries and 10–30s timeouts. | TCK push suite |

---

## 6. Protocol Bindings & Equivalence (`A2A-BIND`)

| ID | Level | Tier | Spec § | Requirement | Verification Method |
|---|---|---|---|---|---|
| `A2A-BIND-001` | **MUST** | Tier 1 | §5.1 | Multi-protocol servers MUST provide identical functionality and semantic equivalence across all declared transports (gRPC, JSON-RPC, REST). | Run probe battery forcing each `--transport` |
| `A2A-BIND-002` | **MUST** | Tier 0 | §5.5 | All JSON responses MUST use camelCase keys and ProtoJSON string names for enums (e.g., `"TASK_STATE_COMPLETED"`). | Inspect raw JSON response keys and enum strings |
| `A2A-BIND-003` | **MUST** | Tier 1 | §5.4 | A2A errors MUST map to protocol-specific status codes per spec §5.4 mapping table. | `a2acli conformance -u <url> --output json` |
| `A2A-BIND-004` | **MUST** | Tier 1 | §11.6, §10.6 | HTTP and gRPC error responses MUST include `google.rpc.ErrorInfo` in error details for A2A-specific error types. | Inspect error payload `details[]` array |
| `A2A-BIND-005` | **MUST** | Tier 1 | §9.2, §11.2 | Service parameters MUST be transmitted as HTTP headers (e.g. `A2A-Version`, `Authorization`). | `a2acli send --svc-param ...` |

---

## 7. Versioning & Negotiation (`A2A-VER`)

| ID | Level | Tier | Spec § | Requirement | Verification Method |
|---|---|---|---|---|---|
| `A2A-VER-001` | **MUST** | Tier 1 | §3.6.2 | Server MUST process requests using requested `A2A-Version` (e.g. `"1.0"`). | `a2acli send "hi" -p 1.0.0 -u <url> --output json` |
| `A2A-VER-002` | **MUST** | Tier 1 | §3.6.2 | Server MUST interpret an empty `A2A-Version` header as version `"0.3"` compatibility mode. | `a2acli send "hi" -p 0.3.0 -u <url> --output json` |
| `A2A-VER-003` | **MUST** | Tier 1 | §3.6.2 | Server MUST return `VersionNotSupportedError` for unsupported protocol major.minor versions. | `a2acli send "hi" --svc-param A2A-Version=99.0 -u <url>` |
| `A2A-VER-004` | **MUST NOT** | Tier 0 | §3.6 | Patch versions MUST NOT be used in version negotiation or Agent Card declarations. | Inspect version strings |

---

## 8. Security & Auth (`A2A-SEC`)

| ID | Level | Tier | Spec § | Requirement | Verification Method |
|---|---|---|---|---|---|
| `A2A-SEC-001` | **MUST** | Tier 1 | §7.4, §3.3.2 | Server MUST reject unauthenticated requests with HTTP 401 / gRPC `UNAUTHENTICATED` if auth is required. | `a2acli get <taskId> -u <url>` without credentials |
| `A2A-SEC-002` | **MUST** | Tier 1 | §13.1 | Server MUST enforce resource scoping so callers only access authorized tasks and push configs. | Attempt to access task ID belonging to another user |
| `A2A-SEC-003` | **MUST NOT** | Tier 1 | §13.1, §3.3.2 | Server MUST NOT reveal existence of unauthorized resources (`TaskNotFoundError` / 404). | Request non-existent vs unauthorized task ID |
| `A2A-SEC-004` | **MUST** | Tier 1 | §13.3, §3.1.11 | `GetExtendedAgentCard` MUST require authentication and return public card security schemes. | `a2acli discover --extended -u <url> --output json` |
| `A2A-SEC-005` | **MUST** | Tier 1 | §7.6.1 | In-task authorization MUST transition task to `TASK_STATE_AUTH_REQUIRED`. | Trigger skill requiring mid-task credentials |

---

## 9. Extensions (`A2A-EXT`)

| ID | Level | Tier | Spec § | Requirement | Verification Method |
|---|---|---|---|---|---|
| `A2A-EXT-001` | **MUST** | Tier 0 | §4.6.1 | Supported extensions MUST be declared in `AgentCard.capabilities.extensions[]` with URI. | Inspect `capabilities.extensions` |
| `A2A-EXT-002` | **MUST** | Tier 1 | §3.3.4 | If server requires an extension (`required: true`) that client did not activate, server MUST return `ExtensionSupportRequiredError`. | Invoke required extension skill without `A2A-Extensions` header |
| `A2A-EXT-003` | **MUST** | Tier 1 | Extension Spec | A2UI v1.0 extension MUST pass schema validation for `application/a2ui+json` DataParts. | `a2acli a2ui validate -u <url> --output json` |

---

## 10. Quality, Design & Operability (`A2A-QUAL`) — Non-Normative

| ID | Level | Tier | Recommendation | Evaluation |
|---|---|---|---|---|
| `A2A-QUAL-001` | Best Practice | Tier 0 | **Intent-Based Skills:** Skills SHOULD represent high-level user intents rather than 1:1 REST endpoint wrappers. | Inspect skill names, descriptions, and tags |
| `A2A-QUAL-002` | Best Practice | Tier 0 | **Rich Descriptions & Examples:** Skills SHOULD include clear descriptions and copy-pasteable examples to guide LLM orchestrators. | Inspect `skills[].examples` array |
| `A2A-QUAL-003` | Best Practice | Tier 1 | **Streaming Progress:** Long-running tasks (>3s) SHOULD emit periodic `WORKING` status events or progress artifacts. | Monitor event frequency during execution |
| `A2A-QUAL-004` | Best Practice | Tier 1 | **Graceful Cancellation:** Server SHOULD respond promptly to `CancelTask` and clean up background resources. | Measure latency between `cancel` and state update |
| `A2A-QUAL-005` | Best Practice | Tier 1 | **Structured Error Details:** Errors SHOULD include helpful diagnostic details in `details[]` without leaking secrets. | Inspect error payload fields |
