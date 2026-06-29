# A2A Conformance Report

**Date:** 2026-06-28
**CLI Version:** v1.4.8-1-g33e542d-dirty
**SDK Source:** `github.com/a2aproject/a2a-go`
**SDK Branch:** `main`

## Conformance Status

- A2A v1.0.0: **PASSING**
- A2A v0.3.0: **PASSING**
- A2UI Extension v1.0: **PASSING**

### Test Results Summary

```text
=== RUN   TestConformance
=== RUN   TestConformance/JSON-RPC
=== RUN   TestConformance/JSON-RPC/Describe
=== RUN   TestConformance/JSON-RPC/SendWait
=== RUN   TestConformance/JSON-RPC/SendStdin
=== RUN   TestConformance/JSON-RPC/ConformanceSmoke
=== RUN   TestConformance/gRPC
=== RUN   TestConformance/gRPC/SendWait
=== RUN   TestConformance/gRPC/ForcegRPC
=== RUN   TestConformance/A2A-0.3.0
    conformance_test.go:186: 0.3.0 compat SUT not found at /Users/ghchinoy/projects/github/a2a-go/e2e/compat/v0_3
=== RUN   TestConformance/A2UI-Extension-v1.0
=== RUN   TestConformance/A2UI-Extension-v1.0/Validate
--- PASS: TestConformance (13.00s)
    --- PASS: TestConformance/JSON-RPC (6.24s)
        --- PASS: TestConformance/JSON-RPC/Describe (0.18s)
        --- PASS: TestConformance/JSON-RPC/SendWait (2.01s)
        --- PASS: TestConformance/JSON-RPC/SendStdin (2.03s)
        --- PASS: TestConformance/JSON-RPC/ConformanceSmoke (2.02s)
    --- PASS: TestConformance/gRPC (4.04s)
        --- PASS: TestConformance/gRPC/SendWait (2.02s)
        --- PASS: TestConformance/gRPC/ForcegRPC (2.02s)
    --- SKIP: TestConformance/A2A-0.3.0 (0.00s)
    --- PASS: TestConformance/A2UI-Extension-v1.0 (2.54s)
        --- PASS: TestConformance/A2UI-Extension-v1.0/Validate (2.14s)
PASS
ok  	github.com/ghchinoy/a2acli/e2e	(cached)
```

*(Auto-generated via make conformance-report)*
