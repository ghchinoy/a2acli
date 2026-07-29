# Evaluation & Reporting Framework

This document defines the evaluation rules, verdict categories, and report format for A2A server conformance reviews.

---

## 1. Conformance Verdict Matrix

A2A protocol evaluation uses an RFC 2119 requirement-level verdict model:

| Overall Verdict | Condition | Action Required |
|---|---|---|
| **`CONFORMANT`** | 100% of applicable **MUST** checks pass.<br>0 **MUST** failures.<br>≥80% of applicable **SHOULD** checks pass. | Ready for production / catalog listing. |
| **`CONFORMANT WITH WARNINGS`** | 100% of applicable **MUST** checks pass.<br>0 **MUST** failures.<br><80% of **SHOULD** checks pass, or quality recommendations flagged. | Compliant with core spec, but has operability or DX gaps. |
| **`NOT CONFORMANT`** | **1 or more MUST checks FAIL.** | Non-compliant. MUST resolve failing MUST items before deployment. |

---

## 2. Requirement Severity Summary Table

When reporting results, summarize findings by area and severity:

| Requirement Area | Total Checks | MUST Pass / Fail | SHOULD Pass / Fail | Status |
|---|---|---|---|---|
| **1. Agent Card & Discovery (`CARD`)** | 8 | _ / _ | _ / _ | PASS / WARN / FAIL |
| **2. Core Operations (`OPS`)** | 11 | _ / _ | _ / _ | PASS / WARN / FAIL |
| **3. Task Lifecycle (`TASK`)** | 6 | _ / _ | _ / _ | PASS / WARN / FAIL |
| **4. Streaming (`STREAM`)** | 5 | _ / _ | _ / _ | PASS / WARN / FAIL |
| **5. Push Notifications (`PUSH`)** | 5 | _ / _ | _ / _ | PASS / WARN / FAIL |
| **6. Protocol Bindings (`BIND`)** | 5 | _ / _ | _ / _ | PASS / WARN / FAIL |
| **7. Versioning (`VER`)** | 4 | _ / _ | _ / _ | PASS / WARN / FAIL |
| **8. Security & Auth (`SEC`)** | 5 | _ / _ | _ / _ | PASS / WARN / FAIL |
| **9. Extensions (`EXT`)** | 3 | _ / _ | _ / _ | PASS / WARN / FAIL |
| **10. Quality & Design (`QUAL`)** | 5 | N/A | _ / _ | INFO / WARN |

---

## 3. Worked-Example Report Table

Below is an example evaluation table section from a completed conformance report:

### Evaluated Endpoint: `http://127.0.0.1:9001`
**Date:** 2026-07-29 | **Target Server:** `a2a-simple / gRPC-echo`

| Check ID | Requirement | Severity | Result | Finding / Rationale |
|---|---|---|---|---|
| `A2A-CARD-001` | Card served at `/.well-known/agent-card.json` | **MUST** | **PASS** | Successfully fetched agent card JSON. |
| `A2A-CARD-002` | Card contains required top-level fields | **MUST** | **PASS** | `name`, `version`, `supportedInterfaces`, `skills` present. |
| `A2A-CARD-006` | Card HTTP endpoint includes `Cache-Control` | **SHOULD** | **WARN** | `Cache-Control` header missing from response. |
| `A2A-OPS-001` | `SendMessage` returns immediately | **MUST** | **PASS** | Returns valid Task object in 42ms. |
| `A2A-OPS-009` | `CancelTask` returns error on completed task | **MUST** | **PASS** | Returned `TaskNotCancelableError` (HTTP 400). |
| `A2A-BIND-001` | Multi-transport behavioral equivalence | **MUST** | **PASS** | Equivalent responses across REST, JSON-RPC, and gRPC. |
| `A2A-VER-003` | Rejects unsupported protocol version | **MUST** | **PASS** | Returned `VersionNotSupportedError` for `A2A-Version=99.0`. |
| `A2A-QUAL-001` | Intent-based skills modeling | **Best Practice** | **PASS** | Skills represent distinct high-level user intents. |

**Final Verdict:** **`CONFORMANT WITH WARNINGS`** (0 MUST failures, 1 SHOULD warning).

---

## 4. Top Remediation Priorities Format

Always conclude the report with the **Top 3 Highest-Impact Fixes**, formatted as copy-pasteable code/config snippets or specific Go/Python SDK options.
