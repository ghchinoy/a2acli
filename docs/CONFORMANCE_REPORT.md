# A2A Conformance Report

**Date:** 2026-02-23  
**CLI Version:** v1.0.0  
**SDK Source:** `github.com/a2aproject/a2a-go`  
**SDK Branch:** `release/spec-v1` (with local fix)  

## Conformance Status: **PASSING**

The `a2acli` has been verified against the A2A Technology Compatibility Kit (TCK) provided by the `a2a-go` SDK.

### Test Results Summary

```text
=== RUN   TestConformance
=== RUN   TestConformance/JSON-RPC
=== RUN   TestConformance/JSON-RPC/Describe
=== RUN   TestConformance/JSON-RPC/SendWait
=== RUN   TestConformance/gRPC
=== RUN   TestConformance/gRPC/SendWait
=== RUN   TestConformance/gRPC/ForcegRPC
--- PASS: TestConformance (10.18s)
    --- PASS: TestConformance/JSON-RPC (4.90s)
        --- PASS: TestConformance/JSON-RPC/Describe (0.38s)
        --- PASS: TestConformance/JSON-RPC/SendWait (2.09s)
    --- PASS: TestConformance/gRPC (4.14s)
        --- PASS: TestConformance/gRPC/SendWait (2.07s)
        --- PASS: TestConformance/gRPC/ForcegRPC (2.07s)
PASS
```

## Dependencies & Blockers

A critical issue was identified in the upstream SDK's TCK implementation during the Spec v1.0 migration.

*   **Issue:** [a2a-go #231](https://github.com/a2aproject/a2a-go/issues/231) - TCK `sut_agent_executor.go` violates V1 Spec.
*   **PR:** [a2a-go #235](https://github.com/a2aproject/a2a-go/pull/235) (Fix initial event sequence).

### Rationale for Fix

Under the A2A Spec v1.0, the `taskupdate.Manager` enforces that the first event emitted by an agent during task execution MUST be a full `Task` or `Message` object. This ensures the task state is properly initialized before any status or artifact updates are processed.

The upstream SUT was incorrectly emitting a `TaskStatusUpdateEvent` as the first event, causing v1.0 clients to reject the response with `invalid agent response`.

### Verification Environment

This report was generated using a local `go.mod` replace directive:
`replace github.com/a2aproject/a2a-go => /Users/ghchinoy/dev/github/a2a-go`

The local SDK source was checked out to `release/spec-v1` with the patch from `fix/tck-v1-initial-event` applied.
