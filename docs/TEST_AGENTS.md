# A2A Test Agent Landscape

A survey of the A2A agents available for exercising and validating `a2acli`, what
each can test, its value (utility vs conformance), gaps worth filling, and a list
of test services still worth building to cover the full A2A pattern space.

> Generated as a housekeeping pass. All surveyed agents speak **A2A protocol 1.0**
> (via `a2a-go/v2`). None currently serve gRPC; only read-aloud plans REST/HTTP+JSON.

## Capability Matrix

| Capability | apex a2a_a2ui | apex a2a_server | a2a-simple | syntaxis | eldamo (candir) | read-aloud (planned) |
|---|---|---|---|---|---|---|
| SDK version | v2.2.1 | v2.2.1 | v2.3.1 | v2.2.0 | v2.2.0 | TBD |
| Protocol | 1.0 | 1.0 | 1.0 | 1.0 | 1.0 | 1.0 |
| **JSONRPC** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ (planned) |
| **gRPC** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **REST/HTTP+JSON** | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ (planned) |
| **Streaming** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ (planned) |
| **Auth** | none | none | Bearer/JWT | none | OAuth2+PKCE | none |
| **Stateful Tasks** | ✅ | ❌ msg-only | ✅ mixed | ✅ | ✅ mixed | ✅ (planned) |
| **Multi-turn** | surface map | ctx history | task refs | FSM sessions | task store | — |
| **Text artifact** | ✅ | ❌ | ✅ | ✅ | ✅ | — |
| **Data/JSON artifact** | ✅ | ❌ | ✅ | ❌ | ❌ | — |
| **Raw/binary artifact** | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ (fallback) |
| **URL/FileURL artifact** | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ (GCS) |
| **Push notifications** | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Extensions** | ✅ A2UI v1.0 | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Extended agent card** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **ListTasks (history)** | ❌ | ❌ | ✅ | ❌ | ✅ Firestore | — |
| **Multi-skill** | 1 | 2 | 7 | 8 | 4 | 1 |

## Per-Agent Assessment

### apex `a2a_a2ui` — A2UI Showcase
- **Exercises in a2acli:** `a2ui validate` (the flagship A2UI extension conformance), `discover` (extension declaration), `send` with Data artifacts.
- **Utility value:** Low as a standalone agent (it's a demo). High as the *only* A2UI extension producer.
- **Conformance value:** **High and unique** — the sole test target for the A2UI v1.0 extension validator. Drives the e2e `A2UI-Extension-v1.0` test.
- **Missing / valuable to add:** inline catalog support (`acceptsInlineCatalogs: true`); client→server round-trips (for A2UI Phase B, `a2ac-k6e`).

### apex `a2a_server` — basic GenAI chat
- **Exercises in a2acli:** `send` (message-only path), `discover`, context continuation.
- **Utility value:** Low — a minimal sample.
- **Conformance value:** Low–medium — useful as the "message-only, no Task" reference (distinct from stateful-Task agents). Validates a2acli handles agents that never create Tasks.
- **Missing / valuable to add:** nothing critical; it intentionally stays minimal.

### a2a-simple — A2A Experiments
- **Exercises in a2acli:** the widest surface — `send` (text/data/raw/url artifacts), `push-config` CRUD (only push-capable agent), `multimodal_echo` (the `--parts/--json/--attach/--data` round-trip target), `--token` auth gating (`admin_echo`), `get`/`subscribe`/`list tasks`, cross-task `--ref`.
- **Utility value:** Medium — it's a test/demo harness, not a product.
- **Conformance value:** **Highest overall.** Richest artifact coverage, the only push-notification target, the only multimodal echo, bearer-auth gating. The de-facto a2acli regression workhorse.
- **Missing / valuable to add:** gRPC or REST transport (would make it the only multi-transport target); an extended-agent-card variant; A2UI or another extension.

### syntaxis — publication engine
- **Exercises in a2acli:** multi-turn FSM conversation (`project_assistant` → the natural `a2ac-srx` REPL target), stateful tasks, streaming status logs, `--skill` targeting across 8 skills, file-as-source workflows.
- **Utility value:** **Highest** — a real product that does genuinely useful work (PDF/Typst generation). Demonstrates a2acli driving a substantial agent.
- **Conformance value:** Medium — strong for multi-turn/session and streaming-status patterns; thin on artifact variety (Text only).
- **Missing / valuable to add:** return generated PDFs as Raw/FileURL artifacts (currently returns filesystem paths as text — a real gap); DataPart for structured review suggestions (currently JSON-in-text).

### eldamo / candir.mithlond.com — Mithlond Elvish Agent
- **Exercises in a2acli:** **OAuth 2.1 auth-code + PKCE** (`auth login`, the only OAuth target), per-skill scopes display in `discover`, 401 actionable hints, persistent task store (`get`/`list` survive restarts), streaming token chunks, CNAME/CIMD flows.
- **Utility value:** **High** — a real, deployed, useful agent (Tolkien linguistics) on Cloud Run.
- **Conformance value:** **High and unique** — the only OAuth2/PKCE target, only Firestore-backed persistent store, only per-skill scope enforcement. Drives the entire auth feature set.
- **Missing / valuable to add:** audio artifacts (TTS of generated names would add Raw/URL coverage); extended agent card (richer card for authenticated callers — would be the `a2ac-o2i` test target).

### read-aloud / Fabulae — (planned)
- **Will exercise in a2acli:** binary artifact save (`a2ac-mfd` — the MP3 Raw/FileURL path, already implemented in anticipation), **REST/HTTP+JSON transport** (the only planned non-JSONRPC target), `audio/mpeg` output modes, `--out-dir` for media.
- **Utility value:** **High** (planned) — real product (text→audio).
- **Conformance value:** **High and unique** (planned) — the only REST transport target and the only binary-audio-artifact producer. Will validate `a2ac-mfd` end-to-end and a2acli's REST transport path against a real server.
- **Status:** A2A not yet built; tracked under `read-aloud-ijo`. a2acli's consumer-side support (`a2ac-mfd`) already shipped.

## Coverage Gaps Across All Agents

These A2A capabilities have **no live test target** today:

| Gap | Impact on a2acli | Closest plan |
|---|---|---|
| **gRPC transport** | a2acli's gRPC path is only tested against the TCK SUT, never a sister agent | none — would need a new/retrofitted server |
| **REST/HTTP+JSON transport** | a2acli's `--transport rest` is TCK-only | read-aloud (planned) |
| **Extended agent card** | `a2ac-o2i` has no test target; feature held | none |
| **A2A extension (non-A2UI)** | generic extension activation untested beyond A2UI | none |
| **Push notification *delivery*** | a2acli tests config CRUD, but never observes an actual webhook callback | none — needs a server that POSTs + a receiver |
| **input-required / auth-required mid-task states** | a2acli's handling of these task states is untested | none |

## Proposed New Test Services

Services worth creating to exercise the full A2A pattern space — independent of
whether a2acli currently supports them (building the target often reveals the
client gap):

1. **`a2a-grpc-echo`** *(high value)* — a minimal echo agent served over **gRPC**
   (and ideally all three transports from one binary). Closes the single biggest
   coverage gap: no sister agent serves gRPC. Would let the e2e suite test
   transport auto-selection and `--transport grpc` against a real agent, not just
   the TCK SUT.

2. **`a2a-multimodal`** *(high value)* — an agent that deterministically returns
   **every artifact type** (Text, Data, Raw bytes, FileURL) and **every task
   state** (working, input-required, auth-required, completed, failed, canceled)
   on demand via skill or keyword. The "kitchen sink" for client rendering and
   state-handling. (apex's A2UI kitchen-sink is the inspiration; this generalizes
   it beyond A2UI.)

3. **`a2a-pushnotify` receiver+sender pair** *(medium value)* — a server that
   actually **delivers** push notifications to a callback, plus a tiny local
   webhook receiver, so a2acli can validate end-to-end push delivery (not just
   config CRUD). Would let a2acli add a `--watch-push` or callback-listen mode.

4. **`a2a-extended-card`** *(medium value)* — an agent advertising
   `extendedAgentCard: true` that returns a richer card to authenticated callers.
   The concrete test target that unblocks `a2ac-o2i` (`discover --extended`).

5. **`a2a-rest`** *(medium value, or fold into #1)* — a REST/HTTP+JSON-bound
   agent, if read-aloud's timeline slips. Validates `--transport rest` against a
   real server before read-aloud ships.

6. **`a2a-interrupt`** *(lower value)* — an agent that drives `input-required`
   and `auth-required` mid-task transitions, to test a2acli's multi-turn
   resumption and the auth-required → `auth login` → resume loop.

### Recommended priority

`a2a-grpc-echo` (#1) and `a2a-multimodal` (#2) give the most coverage per unit of
effort — together they close gRPC, REST (if multi-transport), all artifact types,
and all task states. The others are valuable but narrower, and #4/#5 may be
satisfied by eldamo and read-aloud respectively as those projects evolve.

## How a2acli Features Map to Test Agents

| a2acli feature | Primary test agent | Notes |
|---|---|---|
| `discover`, `send`, `get`, `cancel`, `subscribe` | any | base protocol |
| `send --skill` | a2a-simple, syntaxis, eldamo | multi-skill agents |
| `send --parts/--json/--attach/--data` | a2a-simple (`multimodal_echo`), serve --echo | round-trip validation |
| `push-config *` | a2a-simple | only push-capable agent |
| `auth login` (OAuth2/PKCE) | eldamo/candir | only OAuth target |
| `--token` bearer auth | a2a-simple (`admin_echo`) | bearer gating |
| `a2ui validate` | apex a2a_a2ui | only A2UI producer |
| `conformance` | any | smoke checks |
| binary artifact save (`--out-dir`) | read-aloud (planned), a2a-simple | Raw/URL parts |
| `list tasks` | a2a-simple, eldamo | task-store-backed |
| `discover --extended` (`a2ac-o2i`) | **none yet** | needs `a2a-extended-card` |
| gRPC transport | **TCK SUT only** | needs `a2a-grpc-echo` |
| REST transport | **TCK SUT only** | read-aloud (planned) |
