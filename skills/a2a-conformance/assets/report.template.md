# A2A Protocol Conformance Evaluation Report

**Target Service:** `{{SERVICE_URL}}`  
**Agent Name:** `{{AGENT_NAME}}`  
**Agent Version:** `{{AGENT_VERSION}}`  
**Evaluator:** `a2a-conformance skill`  
**Evaluation Date:** `{{DATE}}`  

---

## Executive Summary

- **Overall Verdict:** **`{{VERDICT}}`** (`CONFORMANT` | `CONFORMANT WITH WARNINGS` | `NOT CONFORMANT`)
- **MUST Checks Passed:** `{{MUST_PASSED}} / {{MUST_TOTAL}}`
- **SHOULD Checks Passed:** `{{SHOULD_PASSED}} / {{SHOULD_TOTAL}}`
- **Supported Transports Tested:** `{{TRANSPORTS}}`

---

## 1. Compliance Summary by Area

| Area | MUST Pass / Total | SHOULD Pass / Total | Area Status |
|---|---|---|---|
| **Agent Card & Discovery (`CARD`)** | `{{CARD_MUST_P}} / {{CARD_MUST_T}}` | `{{CARD_SHOULD_P}} / {{CARD_SHOULD_T}}` | `{{CARD_STATUS}}` |
| **Core Operations (`OPS`)** | `{{OPS_MUST_P}} / {{OPS_MUST_T}}` | `{{OPS_SHOULD_P}} / {{OPS_SHOULD_T}}` | `{{OPS_STATUS}}` |
| **Task Lifecycle (`TASK`)** | `{{TASK_MUST_P}} / {{TASK_MUST_T}}` | `{{TASK_SHOULD_P}} / {{TASK_SHOULD_T}}` | `{{TASK_STATUS}}` |
| **Streaming (`STREAM`)** | `{{STREAM_MUST_P}} / {{STREAM_MUST_T}}` | `{{STREAM_SHOULD_P}} / {{STREAM_SHOULD_T}}` | `{{STREAM_STATUS}}` |
| **Push Notifications (`PUSH`)** | `{{PUSH_MUST_P}} / {{PUSH_MUST_T}}` | `{{PUSH_SHOULD_P}} / {{PUSH_SHOULD_T}}` | `{{PUSH_STATUS}}` |
| **Protocol Bindings (`BIND`)** | `{{BIND_MUST_P}} / {{BIND_MUST_T}}` | `{{BIND_SHOULD_P}} / {{BIND_SHOULD_T}}` | `{{BIND_STATUS}}` |
| **Versioning (`VER`)** | `{{VER_MUST_P}} / {{VER_MUST_T}}` | `{{VER_SHOULD_P}} / {{VER_SHOULD_T}}` | `{{VER_STATUS}}` |
| **Security & Auth (`SEC`)** | `{{SEC_MUST_P}} / {{SEC_MUST_T}}` | `{{SEC_SHOULD_P}} / {{SEC_SHOULD_T}}` | `{{SEC_STATUS}}` |
| **Extensions (`EXT`)** | `{{EXT_MUST_P}} / {{EXT_MUST_T}}` | `{{EXT_SHOULD_P}} / {{EXT_SHOULD_T}}` | `{{EXT_STATUS}}` |

---

## 2. Detailed Findings

### ❌ Blocking MUST Failures (if any)
*(If none, state "None — 100% of MUST requirements met.")*

| Check ID | Requirement | Finding / Failure Rationale | Spec § |
|---|---|---|---|
| `{{FAIL_ID}}` | `{{FAIL_REQ}}` | `{{FAIL_RATIONALE}}` | `{{FAIL_SPEC}}` |

---

### ⚠️ SHOULD Warnings & Deviations
*(Non-blocking conformance gaps or missing recommendations.)*

| Check ID | Requirement | Observation | Remediation Suggestion |
|---|---|---|---|
| `{{WARN_ID}}` | `{{WARN_REQ}}` | `{{WARN_OBS}}` | `{{WARN_SUGGESTION}}` |

---

## 3. Quality & Design Observations (Non-Normative)

- **Skill Modeling:** `{{SKILL_MODELING_OBSERVATION}}`
- **Discoverability:** `{{DISCOVERABILITY_OBSERVATION}}`
- **Transport Equivalence:** `{{TRANSPORT_EQUIVALENCE_OBSERVATION}}`

---

## 4. Top 3 Highest-Impact Action Items

1. **`{{FIX_1_TITLE}}`**: `{{FIX_1_DESC}}`
2. **`{{FIX_2_TITLE}}`**: `{{FIX_2_DESC}}`
3. **`{{FIX_3_TITLE}}`**: `{{FIX_3_DESC}}`
