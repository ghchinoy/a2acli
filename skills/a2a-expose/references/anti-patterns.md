# Common A2A Server Anti-Patterns & Failure Modes

Avoid these 14 common implementation mistakes when building an A2A exposure. Each entry includes the spec section violated and the required fix.

---

## 1. Mapped Endpoint Anti-Pattern
- **Mistake:** Creating 1 A2A skill for every REST endpoint in your API (e.g. 18 skills for CRUD operations).
- **Spec Violation:** Non-normative quality / DX flaw (§A2A-QUAL-001).
- **Fix:** Consolidate endpoints under high-level *user intent* skills (e.g. `manage_user_account`).

## 2. Missing Terminal State Event
- **Mistake:** `Execute` generator terminates after emitting artifacts or working status without yielding `TaskStateCompleted` or `TaskStateFailed`.
- **Spec Violation:** §3.1.2, §3.1.6 (`A2A-STREAM-002`). The client streaming connection hangs indefinitely waiting for a terminal event.
- **Fix:** Always yield a final `TaskStatusUpdateEvent` with state `COMPLETED` or `FAILED`.

## 3. Results in Message Text Instead of Artifacts
- **Mistake:** Returning generated reports, files, or structured JSON inside `Message.parts[].text`.
- **Spec Violation:** §3.7 (`A2A-TASK-005`). Messages MUST NOT be treated as a reliable delivery mechanism for task outputs.
- **Fix:** Emit outputs as named `Artifacts` attached to the Task.

## 4. Undeclared Capabilities
- **Mistake:** Implementing SSE streaming or push configs in server code without declaring `streaming: true` or `pushNotifications: true` in `AgentCard.capabilities`.
- **Spec Violation:** §3.3.4 (`A2A-CARD-002`, `A2A-STREAM-005`). Server behavior contradicts declared discovery metadata.
- **Fix:** Keep `AgentCard.capabilities` in exact lockstep with server options.

## 5. False Capability Declaration
- **Mistake:** Setting `capabilities.streaming: true` in AgentCard, but returning `UnsupportedOperationError` when streaming RPCs are invoked.
- **Spec Violation:** §3.3.4.
- **Fix:** Only declare capabilities that are actually wired and functional.

## 6. Patch Version in Card
- **Mistake:** Setting `version: "1.0.4"` or `protocolVersion: "1.0.0"` in AgentCard.
- **Spec Violation:** §3.6 (`A2A-VER-004`). Patch versions MUST NOT be used in protocol declarations or version negotiation.
- **Fix:** Use major.minor format: `"1.0"`.

## 7. Card Served at Wrong Path
- **Mistake:** Serving card at `/.well-known/agent.json` or `/agent-card.json`.
- **Spec Violation:** §14.3 (`A2A-CARD-001`).
- **Fix:** Mount card at exact well-known path: `/.well-known/agent-card.json`.

## 8. Divergent Behavior Across Transports
- **Mistake:** REST transport returns a Task, but JSON-RPC transport returns an error for the same request.
- **Spec Violation:** §5.1 (`A2A-BIND-001`). All declared protocol bindings MUST provide functionally equivalent representations and behavior.
- **Fix:** Wrap the same `RequestHandler` instance across all transport handlers.

## 9. Non-CamelCase JSON Keys
- **Mistake:** Emitting JSON fields with snake_case (`task_id`, `created_at`) or PascalCase (`TaskID`).
- **Spec Violation:** §5.5 (`A2A-BIND-002`). All JSON serializations MUST use camelCase field names (`taskId`, `createdAt`).
- **Fix:** Use ProtoJSON / SDK standard serializers.

## 10. Non-Standard Error Code Mapping
- **Mistake:** Returning generic HTTP 500 for `TaskNotFoundError` or custom error codes without `google.rpc.ErrorInfo`.
- **Spec Violation:** §5.4, §11.6 (`A2A-BIND-003`, `A2A-BIND-004`).
- **Fix:** Map errors according to spec §5.4 table and include `@type` in `details[]`.

## 11. Ignoring `returnImmediately` Flag
- **Mistake:** Always blocking until task completion, even when caller set `returnImmediately: true`.
- **Spec Violation:** §3.2.2 (`A2A-OPS-003`).
- **Fix:** Honor `returnImmediately` by returning immediately after task creation.

## 12. Unscoped Resource Access
- **Mistake:** `GetTask` or `ListTasks` returning tasks belonging to other authenticated users.
- **Spec Violation:** §13.1 (`A2A-SEC-002`).
- **Fix:** Filter task store queries by caller's authenticated user ID.

## 13. Overwriting `contextId` Mismatches
- **Mistake:** Client passes `contextId: "ctx-1"` with `taskId: "task-9"`, but task belongs to `ctx-2`, and server silently overwrites or generates a new context ID.
- **Spec Violation:** §3.4.1 (`A2A-TASK-003`).
- **Fix:** Reject request with an error on context ID mismatch.

## 14. Missing Stream Cleanup
- **Mistake:** Leaking background goroutines/coroutines when a client disconnects mid-stream.
- **Spec Violation:** Resource leak / Operability flaw.
- **Fix:** Check `ctx.Done()` or `yield` return boolean in execution loop to stop processing when caller cancels.
