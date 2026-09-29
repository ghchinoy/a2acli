# A2A 0.3 compat SUT fixture

The System-Under-Test that `TestConformance/A2A-0.3.0` (`e2e/conformance_test.go`)
runs against. It is the only end-to-end exercise of a2acli's `--protocol 0.3.0`
compat transport (`a2aclient.WithCompatTransport`).

## Why this lives here

The test originally pointed at `$A2A_GO_SRC/e2e/compat/v0_3` in the a2a-go SDK.
Upstream deleted that directory on 2026-02-27 in commit `ff1509c` ("fix: update
compat to work against old lib version (#246)"), replacing the standalone binary
with in-process Go tests. The path existed on a2a-go's `main` for about seven days
and was never present at any released tag, so the case skipped from the day it was
written and never once ran in CI.

A2A 0.3 itself is **not** deprecated — it is normative in the released 1.0.1 spec
(§3.6.2: "Agents MUST interpret empty value as 0.3 version"), and a2acli still ships
`--protocol 0.3.0`. So the fixture is vendored here rather than dropped: depending on
an upstream *internal test* path was the root cause, and owning it locally is what
stops it recurring.

## Provenance

`main.go` is recovered verbatim from upstream `ff1509c^:e2e/compat/v0_3/main.go`.
Two deliberate changes were made when vendoring:

- `go.mod`: the stale `github.com/a2aproject/a2a-go v0.3.6` pin was bumped to
  `v0.3.15` (the current v0.x release); `go.sum` regenerated accordingly.
- `go.mod`: module path renamed to `github.com/ghchinoy/a2acli/e2e/fixtures/v0_3_sut`
  to reflect local ownership.

> Recovering it again requires `git log --full-history` — default history
> simplification hides this deletion.

## Nested module

This is a **separate Go module** on purpose: the legacy `github.com/a2aproject/a2a-go`
(v0.x) and `github.com/a2aproject/a2a-go/v2` cannot coexist in one module, and a2acli's
main module requires `/v2`. Nested modules are excluded from the parent's `./...`, so
`go build`, `go vet`, `go test`, and `golangci-lint` at the repo root do not see this
directory. Build it directly if you need to:

```bash
cd e2e/fixtures/v0_3_sut && go build ./...
```

## Contract the conformance test depends on

Do not break these — `test030Suite` relies on all four:

1. Started as `go run main.go server`, with this directory as the working directory.
2. Listens on an ephemeral port and prints **only that port number** to stdout.
3. Agent card `Name` is exactly `Compat Test Agent`.
4. Serves HTTP JSON-RPC and answers `message/send` with a **Message** (not a Task),
   which a2acli emits as the App-B `{"message":{…}}` wrapper carrying `messageId`.

The server handles exactly one `/invoke` request and then shuts down, so the test's
`Describe` subtest (agent-card fetch only) must run before `SendWait`.
