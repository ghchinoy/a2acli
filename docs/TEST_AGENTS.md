# A2A Test Agent Landscape

A survey of the A2A agents available for exercising and validating `a2acli`, what
each can test, its value (utility vs conformance), gaps worth filling, and a list
of test services still worth building to cover the full A2A pattern space.

> Generated as a housekeeping pass. All surveyed agents speak **A2A protocol 1.0**
> (via `a2a-go/v2`). gRPC + REST/HTTP+JSON multi-transport ships in `a2a-simple`'s
> `grpc-echo` fixture (`a2a-simple-4e1`), covered by e2e (`a2ac-k9i`).

## Capability Matrix

| Capability | apex a2a_a2ui | apex a2a_server | a2a-simple | syntaxis | eldamo (candir) | read-aloud (planned) |
|---|---|---|---|---|---|---|
| SDK version | v2.2.1 | v2.2.1 | v2.3.1 | v2.2.0 | v2.2.0 | TBD |
| Protocol | 1.0 | 1.0 | 1.0 | 1.0 | 1.0 | 1.0 |
| **JSONRPC** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ (planned) |
| **gRPC** | ❌ | ❌ | ✅ (`grpc-echo`) | ❌ | ❌ | ❌ |
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
| **Extended agent card** | ❌ | ❌ | ⏳ planned | ❌ | ✅ | ❌ |
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
- **Shipped:** multi-service restructure (`a2a-simple-gnv`) and the `grpc-echo` multi-transport fixture (`a2a-simple-4e1`) — a single process serving JSON-RPC + REST/HTTP+JSON (`:9002`) and gRPC (`:9003`), exercised by a2acli e2e (`a2ac-k9i`).
- **In progress:** multimodal kitchen-sink (`a2a-simple-lin`) and extended-card (`a2a-simple-2z1`) fixture servers.
- **Missing / valuable to add:** A2UI or another extension producer.

### syntaxis — publication engine
- **Exercises in a2acli:** multi-turn FSM conversation (`project_assistant` → the natural `a2ac-srx` REPL target), stateful tasks, streaming status logs, `--skill` targeting across 8 skills, file-as-source workflows.
- **Utility value:** **Highest** — a real product that does genuinely useful work (PDF/Typst generation). Demonstrates a2acli driving a substantial agent.
- **Conformance value:** Medium — strong for multi-turn/session and streaming-status patterns; thin on artifact variety (Text only).
- **Missing / valuable to add:** return generated PDFs as Raw/FileURL artifacts (currently returns filesystem paths as text — a real gap); DataPart for structured review suggestions (currently JSON-in-text).

### eldamo / candir.mithlond.com — Mithlond Elvish Agent
- **Exercises in a2acli:** **OAuth 2.1 auth-code + PKCE** (`auth login`, the only OAuth target), per-skill scopes display in `discover`, **`discover --extended`** (the only live extended-card target), 401 actionable hints, persistent task store (`get`/`list` survive restarts, `list --status` filtering), streaming token chunks, CNAME/CIMD flows.
- **Utility value:** **High** — a real, deployed, useful agent (Tolkien linguistics) on Cloud Run.
- **Conformance value:** **High and unique** — the only OAuth2/PKCE target, only Firestore-backed persistent store, only per-skill scope enforcement, **only live `extendedAgentCard` target** (`eldamo-server-lde`, shipped). Drives the entire auth feature set and validated `discover --extended` (`a2ac-o2i`).
- **Missing / valuable to add:** audio artifacts (TTS of generated names would add Raw/URL coverage, `eldamo-server-zhw`).

### read-aloud / Fabulae — (planned)
- **Will exercise in a2acli:** binary artifact save (`a2ac-mfd` — the MP3 Raw/FileURL path, already implemented in anticipation), **REST/HTTP+JSON transport** (the only planned non-JSONRPC target), `audio/mpeg` output modes, `--out-dir` for media.
- **Utility value:** **High** (planned) — real product (text→audio).
- **Conformance value:** **High and unique** (planned) — the only REST transport target and the only binary-audio-artifact producer. Will validate `a2ac-mfd` end-to-end and a2acli's REST transport path against a real server.
- **Status:** A2A not yet built; tracked under `read-aloud-ijo`. a2acli's consumer-side support (`a2ac-mfd`) already shipped.

## Coverage Gaps Across All Agents

These A2A capabilities have **no live test target** today:

| Gap | Impact on a2acli | Closest plan |
|---|---|---|
| ~~gRPC transport~~ | **Closed** — `a2a-simple` `grpc-echo` now exercises gRPC against a sister agent in e2e | `a2a-simple-4e1` ✅ (`a2ac-k9i`) |
| ~~REST/HTTP+JSON transport~~ | **Closed** — `grpc-echo` serves REST/HTTP+JSON alongside gRPC; e2e-covered | `a2a-simple-4e1` ✅ (`a2ac-k9i`) |
| **A2A extension (non-A2UI)** | generic extension activation untested beyond A2UI | none |
| **Push notification *delivery*** | a2acli tests config CRUD, but never observes an actual webhook callback | none — needs a server that POSTs + a receiver |
| **input-required / auth-required mid-task states** | a2acli's handling of these task states is untested | `a2a-simple-lin` (multimodal kitchen-sink drives all states) |

> **Closed gap:** Extended agent card — eldamo/candir now advertises `extendedAgentCard: true`
> and serves a richer card to authenticated callers; `discover --extended` is live-validated against it.

## Proposed New Test Services

Services worth creating to exercise the full A2A pattern space — independent of
whether a2acli currently supports them (building the target often reveals the
client gap):

1. ~~**`a2a-grpc-echo`**~~ *(SHIPPED — `a2a-simple-4e1`)* — `a2a-simple`'s `grpc-echo`
   serves all three transports (gRPC, JSON-RPC, REST/HTTP+JSON) from one binary and
   echoes message parts as named artifacts. a2acli e2e (`a2ac-k9i`) now tests transport
   auto-selection and `--transport grpc/rest/jsonrpc` against it, not just the TCK SUT.

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

4. **`a2a-extended-card`** *(now partially satisfied)* — eldamo/candir now serves
   a live `extendedAgentCard`, which validated `discover --extended` (`a2ac-o2i`).
   A deterministic fixture (`a2a-simple-2z1`) is still worth building for CI
   (eldamo requires real OAuth + network).

5. **`a2a-rest`** *(medium value, or fold into #1)* — a REST/HTTP+JSON-bound
   agent, if read-aloud's timeline slips. Validates `--transport rest` against a
   real server before read-aloud ships.

6. **`a2a-interrupt`** *(lower value)* — an agent that drives `input-required`
   and `auth-required` mid-task transitions, to test a2acli's multi-turn
   resumption and the auth-required → `auth login` → resume loop.

### Recommended priority

`a2a-grpc-echo` (#1) is **done** (`a2a-simple-4e1`), closing the gRPC and
REST/HTTP+JSON gaps. `a2a-multimodal` (#2, `a2a-simple-lin`) is next and closes
all artifact types and task states. The others are valuable but narrower, and
#4/#5 may be satisfied by eldamo and read-aloud respectively as those projects evolve.

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
| `list tasks` (+ `--status`/`--context`) | a2a-simple, eldamo | task-store-backed |
| `discover --extended` | eldamo/candir (live) | `a2a-simple-2z1` for deterministic CI |
| gRPC transport | a2a-simple `grpc-echo` (e2e) + TCK SUT | `a2a-simple-4e1` ✅ |
| REST transport | a2a-simple `grpc-echo` (e2e) + TCK SUT | `a2a-simple-4e1` ✅ |
