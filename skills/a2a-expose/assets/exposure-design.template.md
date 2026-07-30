# A2A Exposure Design Proposal

**Proposed Agent Name:** `{{AGENT_NAME}}`  
**Target Codebase / Service:** `{{SERVICE_NAME}}`  
**Target SDK / Language:** `{{SDK_LANGUAGE}}` (`a2a-go` v2.4.0 | `a2a-sdk` >=1.0)  

---

## 1. Skill Modeling & Intent Granularity

*(Consolidate underlying API endpoints into high-level user intent skills.)*

| Skill ID | Display Name | Intent Description | Mapped Underlying API(s) |
|---|---|---|---|
| `{{SKILL_1_ID}}` | `{{SKILL_1_NAME}}` | `{{SKILL_1_DESC}}` | `{{SKILL_1_MAPPED_APIS}}` |

---

## 2. Protocol & Transport Choices

- **Primary Transport:** `{{PRIMARY_TRANSPORT}}` (`JSONRPC` at `/invoke`)
- **Additional Transports:** `{{ADDITIONAL_TRANSPORTS}}` (`HTTP+JSON` at `/`, `GRPC` at `host:port`)
- **Capabilities Declared:**
  - `streaming`: `{{CAPABILITY_STREAMING}}` (`true` | `false`)
  - `pushNotifications`: `{{CAPABILITY_PUSH}}` (`true` | `false`)
  - `extendedAgentCard`: `{{CAPABILITY_EXTENDED_CARD}}` (`true` | `false`)

---

## 3. Data & Artifact Delivery Model

- **Primary Interaction Mode:** `{{INTERACTION_MODE}}` (`Task` | `Message`)
- **Output Artifacts Produced:**
  - Artifact 1: `{{ARTIFACT_1_NAME}}` (`{{ARTIFACT_1_MIME}}`)
- **Task Store Implementation:** `{{TASK_STORE_TYPE}}` (`InMemory` | `PostgreSQL` | `SQLite` | `Custom`)

---

## 4. Authentication & Security Mapping

- **Security Scheme:** `{{SECURITY_SCHEME_NAME}}` (`bearerAuth` | `apiKeyAuth` | `oauth2` | `none`)
- **Auth Enforcement:** `{{AUTH_ENFORCEMENT_LOCATION}}` (Middleware | Interceptor | Per-Skill)

---

## 5. Proposed Agent Card JSON Preview

```json
{{PROPOSED_AGENT_CARD_JSON}}
```

---

**Approval Request:** Please review and approve this design before implementation begins.
