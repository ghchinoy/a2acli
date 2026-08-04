# A2A CLI Journey & UX Test Plan

**Document Version:** 1.0  
**Target:** `a2acli` v1.10.0+  
**Scope:** End-to-end user-experience validation, conversation continuity, error handling, argument validation, and output contract verification.

---

## 1. Purpose & Relationship to Existing Test Suites

This test plan defines a **UX Journey Suite** for `a2acli`. It complements the existing protocol-conformance suite in `e2e/conformance_test.go` (692 lines, 7 sub-suites).

- **Conformance Suite (`e2e/conformance_test.go`):** Validates wire-level protocol correctness against A2A v1.0 / v0.3.0 specs (JSON-RPC, gRPC, REST, A2UI schemas, transport auto-selection). Answers: *"Does the CLI generate valid protocol frames?"*
- **UX Journey Suite (This Plan):** Validates the developer experience, multi-turn conversation workflows, command-line ergonomics, exit codes, and output clarity against real-world test agents. Answers: *"Can a developer accomplish a multi-turn task without hitting CLI footguns or needing to patch a server?"*

---

## 2. Empirically Derived Testing Methodology

The testing methodology incorporates an 8-point empirical evaluation framework:

1. **Comparative A/B Baseline:** Evaluated `a2acli` against legacy CLI tools side-by-side using identical prompts and agents.
2. **Server-Side Ground Truth Verification:** Inspected server logs during execution to verify actual wire parameters (`task_id`, `context_id`, `referenceTaskIds`) sent by the client.
3. **Wire-Level Payload Logging:** Enabled server payload logging (`-payload`) to observe complete request/response structures.
4. **Latency Injection for Stream Observation:** Utilized delayed response handlers (`time.sleep`) to make streaming vs blocking (`--wait`) semantics and spinner states observable.
5. **Wall-Clock Timing:** Used timing benchmarks on every command to evaluate client overhead and execution duration.
6. **Complexity Escalation:** Tested in progression: minimal echo -> delayed sleep echo -> stateful Gemini LLM.
7. **Horizontal Concurrency Scaling:** Opened multiple concurrent terminal sessions to test session isolation, stream multiplexing, and server load handling.
8. **Clean-Room & Restart Resilience:** Cleared caches (`go clean -modcache`), rebuilt binaries, and restarted agent processes mid-conversation to test stale ID handling.

---

## 3. Baseline User Journey Evaluation Cases

The table below formalizes the baseline user journey evaluation cases:

| Case ID | User Journey Step | Command / Prompt | Expected Behavior | Category |
|---|---|---|---|---|
| **UC01.1** | Install & Discovery | `a2a discover http://127.0.0.1:9999` | Accepts positional URL argument without error | Discovery Grammar |
| **UC01.2** | Discovery with Flag | `a2a discover --service-url http://127.0.0.1:9999` | Displays agent name, version, skills | Discovery Output |
| **UC01.3** | Extended Agent Card | `a2a discover -u http://127.0.0.1:9999 --extended` | Resolves extended card without client-side auth pre-flight | Discovery Auth |
| **UC01.4** | Single-turn Streaming | `time a2a send -u http://127.0.0.1:9999 "Hello"` | Streams response and displays preview artifact | Streaming UX |
| **UC01.5** | Single-turn Blocking | `time a2a send -u http://127.0.0.1:9999 "Hello" --wait` | Returns blocking result with Task ID footer | Blocking UX |
| **UC02.1** | Continuing Task | `a2a send -u http://127.0.0.1:9999 "Hello" --task <completed-id>` | Warns on terminal task state and fails non-zero on error | Task State Handling |
| **UC02.2** | Task History Check | `a2a list tasks -u http://127.0.0.1:9999` | Displays formatted table with CONTEXT ID column | History Output |
| **UC02.3** | Referencing Task | `a2a send -u http://127.0.0.1:9999 "Write haiku" --ref <task-id>` | References prior task artifacts for cross-task chaining | Artifact Chaining |
| **UC03.1** | Multi-Terminal Scale | `a2a send -u http://127.0.0.1:9999 "Hello"` (x9) | Executes concurrent streaming sessions cleanly | Concurrency |
| **UC03.2** | Gemini LLM Interaction | Multi-terminal send to Gemini-backed agent | Handles multi-turn streaming LLM output | Model Integration |
| **UC03.3** | Multi-turn Conversation | `a2a send --context <context-id> "Followup"` | Maintains conversation thread across multi-turn messages | Thread Continuity |
| **Extra** | JSON Output Mode | `a2a send -u http://127.0.0.1:9999 "Hello" -o json` | Outputs NDJSON without local directory collision | Flag Grammar |

---

## 4. Target Agent Inventory & Environment Setup

### 4.1 Target Agent Matrix

| Target Name | Repository Path | Transport(s) | Default Port | Auth | Tasks? | Multi-turn Support | Credentials Needed | Execution Tier |
|---|---|---|---|---|---|---|---|---|
| **`grpc-echo`** | `/workspace/a2a-experiments/cmd/grpc-echo` | JSON-RPC, REST, gRPC (9003) | 9002 | Bearer (`secret-token`) | Always | Task-based (`--task`) | None | **Tier 0** (CI) |
| **`multimodal`** | `/workspace/a2a-experiments/cmd/multimodal` | JSON-RPC, REST | 9004 | Bearer | Always | Live task (`state-input-required`) | None (`./scripts/gen-assets.sh`) | **Tier 0** (CI) |
| **`server`** | `/workspace/a2a-experiments/cmd/server` | JSON-RPC | 9001 | Bearer | Mixed | State Bridge (`--task` & `--ref`) | `GEMINI_API_KEY` | **Tier 1** (Local) |
| **`a2ui`** | `/workspace/a2a-experiments/cmd/a2ui` | JSON-RPC | 9005 | Bearer | Always | Context map (`--context`) | Vertex AI (`GOOGLE_CLOUD_PROJECT` + `LOCATION`) | **Tier 1** (Local) |
| **py `helloworld`** | `/workspace/a2a-dart/test/support/agents/helloworld` | JSON-RPC | 9999 | Extended Card | Never | None | None | **Tier 0** (CI) |
| **py `a2a-samples`** | `https://github.com/a2aproject/a2a-samples` | JSON-RPC | 9999 | Extended Card | Always | `tasks/list`, `contextId` | None | **Tier 0** (Network) |

### 4.2 Critical Environment & Tooling Prerequisites

1. **Rebuild Binary Target Requirement:** Pre-compiled binaries checked into `/workspace/a2a-experiments/bin/` are Darwin arm64. They must be rebuilt for Linux arm64 before testing:
   ```bash
   cd /workspace/a2a-experiments && make build
   ```
2. **Asset Generation Requirement:** `cmd/multimodal` requires synthetic test assets before execution:
   ```bash
   cd /workspace/a2a-experiments && ./scripts/gen-assets.sh
   ```
3. **Port Collision Prevention:**
   - `a2a-experiments/cmd/server` and `apex/tools/sample_servers/a2a_server` both default to port `9001`.
   - `a2a-experiments/cmd/grpc-echo` and `apex/tools/sample_servers/a2a_a2ui` both default to port `9002`.
   - Use `-port <unique_port>` when launching services concurrently.
4. **Task Store Authentication & Ownership Rules:**
   - `tasks/list` requires an authenticated identity. Pass `--auth "Bearer secret-token"` on `a2a-experiments` agents. Unauthenticated calls return `UNAUTHENTICATED`.
   - Tasks are scoped to the creating user. Running `send` anonymously and then calling `get` or `list` with a token will return `TaskNotFound`. Keep authentication flags consistent within a test run.
5. **Payload Logging:** All `a2a-experiments` servers support `-payload` to log full request/response wire payloads to stderr, replacing hand-patched executor print statements.

---

## 5. Expanded Journey Test Suites (TP-1 through TP-7)

### Suite TP-1: Onboarding, Installation & Binary Environment

#### Test Case TP-1.1: `go install` Version Fallback Verification
- **CLI Domain:** Installation & Versioning
- **Target:** Clean Go environment
- **Tier:** Tier 0
- **Preconditions:** Go 1.22+ installed; `a2acli` not present in `GOBIN`.
- **Command:**
  ```bash
  go install github.com/ghchinoy/a2acli/cmd/a2acli@latest
  $(go env GOPATH)/bin/a2acli version
  ```
- **Assertions:**
  - Standard output MUST NOT contain `version dev`, `commit: none`, or `built at: unknown`.
  - Version line MUST match module release tag or VCS commit info (e.g. `v1.10.0` or commit hash).
  - Exit code MUST be `0`.

#### Test Case TP-1.2: PATH & Alias Guidance Verification
- **CLI Domain:** Shell Environment & Setup
- **Target:** Documentation
- **Tier:** Tier 0
- **Command:** `grep -i "GOPATH" /workspace/a2acli/README.md`
- **Assertions:**
  - Documentation MUST provide copy-pasteable export lines for `export PATH="$PATH:$(go env GOPATH)/bin"`.
  - Documentation MUST mention `alias a2a=a2acli` shorthand configuration.

---

### Suite TP-2: Agent Discovery, Grammar & Argument Validation

#### Test Case TP-2.1: Ergonomic Positional URL on `discover`
- **CLI Domain:** Discovery Positional Arguments
- **Target:** `cmd/grpc-echo` (port 9002)
- **Tier:** Tier 0
- **Command:**
  ```bash
  /workspace/a2a-experiments/bin/grpc-echo -port 9002 &
  ECHO_PID=$!
  sleep 1
  /workspace/a2acli/bin/a2acli discover http://127.0.0.1:9002
  kill $ECHO_PID
  ```
- **Assertions:**
  - Client MUST resolve card from `http://127.0.0.1:9002/.well-known/agent-card.json`.
  - Client MUST NOT attempt connection to default `127.0.0.1:9001`.
  - Output MUST display `Agent: Multi-Transport Echo Server`.
  - Exit code MUST be `0`.

#### Test Case TP-2.2: Strict Argument Validation on Zero-Arg Commands
- **CLI Domain:** Argument Validation
- **Target:** Local CLI
- **Tier:** Tier 0
- **Command:** `/workspace/a2acli/bin/a2acli version invalid_extra_arg 2>&1`
- **Assertions:**
  - Output MUST contain error message indicating unexpected positional argument.
  - Process MUST exit with non-zero status code (`1`).

#### Test Case TP-2.3: Unauthenticated Extended Card Resolution
- **CLI Domain:** Extended Card Resolution
- **Target:** Python `helloworld` sample (port 9999)
- **Tier:** Tier 0
- **Preconditions:** `uv` installed; Python 3.10+ available.
- **Command:**
  ```bash
  cd /workspace/a2a-dart/test/support/agents/helloworld && uv run . &
  PY_PID=$!
  sleep 2
  /workspace/a2acli/bin/a2acli discover http://127.0.0.1:9999 --extended
  kill $PY_PID
  ```
- **Assertions:**
  - Client MUST NOT abort client-side with `authentication required`.
  - Client MUST issue RPC request and display extended agent card details (`Hello World Agent - Extended Edition`).
  - Exit code MUST be `0`.

---

### Suite TP-3: Single-Turn Execution & Output Formatting

#### Test Case TP-3.1: Suppressing Selection Chrome in Standard Mode
- **CLI Domain:** Clean Output Formatting
- **Target:** `cmd/grpc-echo` (port 9002)
- **Tier:** Tier 0
- **Command:**
  ```bash
  /workspace/a2a-experiments/bin/grpc-echo -port 9002 &
  ECHO_PID=$!
  sleep 1
  /workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9002 "Hello World"
  kill $ECHO_PID
  ```
- **Assertions:**
  - `stdout` MUST NOT contain `Auto-selected transport: JSONRPC` unless `--verbose` is passed.
  - TUI / output MUST render cleanly without preceding console noise.

#### Test Case TP-3.2: Complete Output Contract in Text Mode (`--output text`)
- **CLI Domain:** Output Summary Footers
- **Target:** `cmd/grpc-echo` (port 9002)
- **Tier:** Tier 0
- **Command:**
  ```bash
  /workspace/a2a-experiments/bin/grpc-echo -port 9002 &
  ECHO_PID=$!
  sleep 1
  /workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9002 "Hello World" --output text
  kill $ECHO_PID
  ```
- **Assertions:**
  - Output MUST display message artifact content (`You said: Hello World`).
  - Output MUST terminate with standardized summary footer containing `Task ID:` and `Context ID:`.
  - Exit code MUST be `0`.

---

### Suite TP-4: Conversation Continuity & Context Management

#### Test Case TP-4.1: Initiating and Continuing Multi-Turn Conversation (`--context`)
- **CLI Domain:** Conversation Thread Continuity
- **Target:** `cmd/multimodal` (port 9004)
- **Tier:** Tier 0
- **Command:**
  ```bash
  /workspace/a2a-experiments/bin/multimodal -port 9004 &
  MM_PID=$!
  sleep 1
  # Turn 1: Initiate conversation
  OUT1=$(/workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9004 "state-input-required" --wait)
  CTX_ID=$(echo "$OUT1" | grep "Context ID:" | awk '{print $3}')
  
  # Turn 2: Continue using --context
  OUT2=$(/workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9004 "Verification code: 123456" --context "$CTX_ID" --wait)
  kill $MM_PID
  ```
- **Assertions:**
  - Turn 1 output MUST print explicit `Context ID: <uuid>`.
  - Turn 2 execution MUST succeed and include `Context ID: <uuid>` matching Turn 1.
  - Wire logs MUST confirm `msg.ContextID` was transmitted in request payload.

#### Test Case TP-4.2: Terminal Task State Handling (`--task` Validation)
- **CLI Domain:** Task State Validation
- **Target:** `cmd/grpc-echo` (port 9002)
- **Tier:** Tier 0
- **Command:**
  ```bash
  /workspace/a2a-experiments/bin/grpc-echo -port 9002 &
  ECHO_PID=$!
  sleep 1
  # Step 1: Complete a task
  OUT1=$(/workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9002 "First request" --wait)
  TASK_ID=$(echo "$OUT1" | grep "Task ID:" | awk '{print $3}')
  
  # Step 2: Attempt continuing completed task in strict mode vs default mode
  ERR_WARN=$(/workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9002 "Second request" --task "$TASK_ID" 2>&1)
  
  /workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9002 "Second request" --task "$TASK_ID" --strict 2>&1
  STRICT_EXIT=$?
  kill $ECHO_PID
  ```
- **Assertions:**
  - Default mode MUST output warning to `stderr` indicating task is in terminal state `TASK_STATE_COMPLETED` and suggest using `--context`.
  - `--strict` mode MUST exit with non-zero status code (`1`) and output error message `ErrCodeFailedPrecondition`.

#### Test Case TP-4.3: Task Listing Table Formatting & Context Visibility
- **CLI Domain:** Task History Table Output
- **Target:** `cmd/grpc-echo` (port 9002)
- **Tier:** Tier 0
- **Command:**
  ```bash
  /workspace/a2a-experiments/bin/grpc-echo -port 9002 &
  ECHO_PID=$!
  sleep 1
  /workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9002 "Test message" --auth "Bearer secret-token" --wait
  LIST_OUT=$(/workspace/a2acli/bin/a2acli list tasks -u http://127.0.0.1:9002 --auth "Bearer secret-token")
  kill $ECHO_PID
  ```
- **Assertions:**
  - Header line MUST contain `TASK ID`, `CONTEXT ID`, `STATUS`, `CREATED AT`.
  - `STATUS` column values MUST be stripped of prefix (e.g. `COMPLETED` instead of `TASK_STATE_COMPLETED`).
  - Table headers and separator lines align properly without overflow.

#### Test Case TP-4.4: Disambiguating `--ref` Artifact Reference
- **CLI Domain:** Artifact Reference Chaining
- **Target:** `cmd/server` (port 9001)
- **Tier:** Tier 1
- **Command:**
  ```bash
  /workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9001 "Research Go" --skill ai_researcher --wait
  # Capture TASK_ID
  /workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9001 "Summarize report" --skill summarize --ref "$TASK_ID" --wait
  ```
- **Assertions:**
  - `send --help` MUST explicitly state `--ref` attaches completed task outputs as reference context, whereas `--context` maintains active conversation state.
  - Continuation footers MUST NOT list `--ref` as a conversation continuation mechanism.

---

### Suite TP-5: Error Handling, Exit Codes & Robustness

#### Test Case TP-5.1: Streaming Failure Non-Zero Exit Code
- **CLI Domain:** Streaming Error Codes
- **Target:** `cmd/multimodal` (port 9004)
- **Tier:** Tier 0
- **Command:**
  ```bash
  /workspace/a2a-experiments/bin/multimodal -port 9004 &
  MM_PID=$!
  sleep 1
  /workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9004 "state-failed"
  EXIT_CODE=$?
  kill $MM_PID
  ```
- **Assertions:**
  - Process MUST terminate with non-zero exit code (`1`).
  - `stderr` (or TUI error block) MUST display failure state details.

#### Test Case TP-5.2: Zero-Event Stream Failure Detection
- **CLI Domain:** Empty Stream Detection
- **Target:** Invalid endpoint / unresponsive server
- **Tier:** Tier 0
- **Command:**
  ```bash
  /workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9998 "Test" 2>&1
  EXIT_CODE=$?
  ```
- **Assertions:**
  - Process MUST NOT exit cleanly with code `0`.
  - Process MUST exit with non-zero code (`1`) and display connection/stream initialization error.

---

### Suite TP-6: Concurrency, Session Isolation & Scale

#### Test Case TP-6.1: High-Concurrency Session Isolation
- **CLI Domain:** Concurrency & Session Isolation
- **Target:** `cmd/grpc-echo` (port 9002)
- **Tier:** Tier 0
- **Command:**
  ```bash
  /workspace/a2a-experiments/bin/grpc-echo -port 9002 &
  ECHO_PID=$!
  sleep 1
  for i in {1..9}; do
    /workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9002 "Concurrent message $i" --wait > /tmp/out_$i.txt &
  done
  wait
  kill $ECHO_PID
  ```
- **Assertions:**
  - All 9 invocations MUST complete with exit code `0`.
  - Each output file `/tmp/out_i.txt` MUST contain distinct `Task ID` and `Context ID` pairs.

---

### Suite TP-7: Output Contracts & Scripting Safety

#### Test Case TP-7.1: Flag Collision Protection (`-o` vs `-d`)
- **CLI Domain:** Flag Shorthand Collision Guard
- **Target:** Local CLI
- **Tier:** Tier 0
- **Command:**
  ```bash
  /workspace/a2acli/bin/a2acli send -u http://127.0.0.1:9002 "test" -d json 2>&1
  EXIT_CODE=$?
  ```
- **Assertions:**
  - Command MUST fail with error: `invalid directory name "json": did you mean -o or --output json?`.
  - Process MUST NOT create a local filesystem directory named `json/`.
  - Process MUST exit with non-zero code (`1`).

---

## 6. Capability & Test Matrix

| Requirement / Feature | Short Description | Test Case(s) | Primary Target | Tier |
|---|---|---|---|---|
| **Multi-turn Context** | Persistent `--context` flag on `send` | **TP-4.1** | `cmd/multimodal` | Tier 0 |
| **Context Surfacing** | Display `Context ID` in all output footers | **TP-3.2, TP-4.1** | `cmd/grpc-echo` | Tier 0 |
| **Terminal Task Guard** | Pre-flight check and warning for `--task` | **TP-4.2** | `cmd/grpc-echo` | Tier 0 |
| **Artifact Chaining** | `--ref` for referencing prior task artifacts | **TP-4.4** | `cmd/server` | Tier 1 |
| **Task History Table** | `list tasks` CONTEXT ID column & formatting | **TP-4.3** | `cmd/grpc-echo` | Tier 0 |
| **Positional URL** | Positional URL argument on `discover` | **TP-2.1** | `cmd/grpc-echo` | Tier 0 |
| **Strict Arg Check** | `cobra.NoArgs` on zero-arg subcommands | **TP-2.2** | Local CLI | Tier 0 |
| **Flag Shorthands** | `-o` for output, `-d` for out-dir | **TP-7.1** | Local CLI | Tier 0 |
| **Streaming Exit Code** | Non-zero exit code on stream failure | **TP-5.1** | `cmd/multimodal` | Tier 0 |
| **Zero-Event Stream** | Non-zero exit code on empty stream | **TP-5.2** | Unresponsive Port | Tier 0 |
| **Extended Card** | Unconditional `discover --extended` RPC | **TP-2.3** | py `helloworld` | Tier 0 |
| **Compact Output** | Single-frame `--output compact` mode | **TP-3.2** | `cmd/grpc-echo` | Tier 0 |
| **Text Mode Summary** | Summary footer in `--output text` mode | **TP-3.2** | `cmd/grpc-echo` | Tier 0 |
| **Transport Banner** | Clean TUI output without banner noise | **TP-3.1** | `cmd/grpc-echo` | Tier 0 |
| **Table Alignment** | Formatted `STATUS` column in `list tasks` | **TP-4.3** | `cmd/grpc-echo` | Tier 0 |
| **Version Fallback** | `ReadBuildInfo` version fallback | **TP-1.1** | Go Toolchain | Tier 0 |
| **PATH Setup Docs** | GOPATH/bin and alias guidance | **TP-1.2** | Documentation | Tier 0 |

---

## 7. Automation Strategy & CI Integration

1. **Automated E2E Suite (`e2e/conformance_test.go`):**
   - All Tier 0 journey cases (TP-2.1, TP-3.1, TP-3.2, TP-4.1, TP-4.2, TP-4.3, TP-5.1, TP-7.1) are integrated into `e2e/conformance_test.go` as `TestConformance/JourneySuites`.
   - Executable via Makefile target:
     ```bash
     make test-journey
     ```
2. **Local/Manual Utility Integration (Tier 1 & Tier 2):**
   - Tier 1/2 tests (Gemini/Vertex AI dependent) gate with `t.Skipf` when `GOOGLE_CLOUD_PROJECT` or `GOOGLE_CLOUD_LOCATION` environment variables are absent, preserving fast, offline CI builds.
