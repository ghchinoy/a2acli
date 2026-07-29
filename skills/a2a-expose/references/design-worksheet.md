# A2A Exposure Design Worksheet & Decision Tree

Before writing any code or configuring server handlers, perform an explicit A2A exposure design pass. Do NOT map REST endpoints 1:1 to A2A skills. Answer the following 10 architectural modeling questions.

---

## The 10 Modeling Questions

### 1. Skill Granularity (Intent vs. API Endpoint)
- **Question:** What high-level user intents does your service fulfill?
- **Rule:** A skill represents an *intent an orchestrating LLM or user asks for*, not a URL path.
- **Guideline:** If your API has 15 endpoints (e.g. `POST /users`, `GET /users/:id`, `PUT /users/:id/preferences`), do NOT create 15 skills. Create 1 or 2 skills (e.g. `manage_user_profile`) with descriptive intent blurbs.

### 2. Interaction Model: Simple Response (`Message`) vs Async Job (`Task`)
- **Question:** Does the operation complete synchronously in under ~1-2 seconds without producing persistent outputs or requiring multi-turn follow-ups?
- **Rule:**
  - Return a bare **`Message`** ONLY if: execution is synchronous, sub-second, stateless, and produces no persistent files/artifacts.
  - Return a **`Task`** if: execution takes >2 seconds, streams updates, produces files/reports, or requires multi-turn dialogue (`INPUT_REQUIRED` / `AUTH_REQUIRED`).

### 3. Output Delivery: Message Parts vs Task Artifacts
- **Question:** What does the user or agent keep after execution finishes?
- **Rule:** Per spec §3.7, outputs SHOULD be delivered as **`Artifacts`** attached to the Task, NOT embedded in status message text. Messages MUST NOT be treated as a reliable delivery mechanism for critical outputs.
- **Guideline:** Summaries, generated documents, images, JSON datasets, and files belong in `Artifacts`.

### 4. Real-Time Streaming Support
- **Question:** Does your agent emit incremental progress updates, status changes, or streaming text chunks?
- **Rule:**
  - If YES: Declare `capabilities.streaming = true` in `AgentCard` AND implement `SendStreamingMessage` / `SubscribeToTask`.
  - If NO: Omit `capabilities.streaming` (or set `false`). Attempts to stream MUST return `UnsupportedOperationError` (§3.3.4).

### 5. Conversation Context & Multi-Turn State
- **Question:** Does the agent support follow-up questions or iterative refinement on an existing task?
- **Rule:**
  - Use `contextId` to group related tasks/messages in a conversation.
  - Use `referenceTaskIds` to accept completed tasks as context.
  - Choose task store: `InMemory` for single-instance/dev, DB-backed (`PostgreSQL`/`SQLite`/Custom) for production persistence.

### 6. Authentication & Security Scheme Mapping
- **Question:** How does your existing service authenticate callers?
- **Rule:** Declare matching scheme in `AgentCard.securitySchemes`:
  - API Key in header/query → `APIKeySecurityScheme`
  - Bearer JWT / Basic → `HTTPAuthSecurityScheme`
  - OAuth 2.0 / 2.1 → `OAuth2SecurityScheme` + flows (`authorizationCode`, `clientCredentials`)
  - Mid-task user authorization (e.g., OAuth consent mid-execution) → In-task auth (§7.6) with `TASK_STATE_AUTH_REQUIRED`.

### 7. Asynchronous Push Notifications
- **Question:** Will tasks run for minutes or hours after the client disconnects?
- **Rule:**
  - If YES: Declare `capabilities.pushNotifications = true`, AND wire both a `ConfigStore` and a `PushSender` webhook dispatcher.
  - If NO: Omit capability. Attempts to configure push MUST return `PushNotificationNotSupportedError`.

### 8. Transport Bindings & Functional Equivalence
- **Question:** Which protocols will clients use to connect?
- **Rule:**
  - Primary default: `JSONRPC` over HTTP at `/invoke` or `/`.
  - Optional additions: `HTTP+JSON` (REST) for browser/curl friendliness; `GRPC` for high-throughput internal microservices.
  - **CRITICAL (Spec §5.1):** All declared transports MUST provide identical functionality and semantic equivalence.

### 9. Input & Output Modalities
- **Question:** What MIME types does your agent accept and produce?
- **Rule:** Declare `defaultInputModes` (e.g. `["text"]`) and `defaultOutputModes` (e.g. `["text", "file"]`) at card root, and override per-skill in `skills[].inputModes` / `outputModes`.

### 10. Agent Discoverability & Metadata
- **Question:** Will an LLM orchestrator understand when to invoke your skills?
- **Rule:** Each skill MUST have an intent-rich `description`, relevant `tags`, and concrete `examples` showing user prompts.

---

## Artifact Output

After completing this worksheet, fill out `assets/exposure-design.template.md` and present it to the user for sign-off before implementing the code.
