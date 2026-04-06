# A2A Conformance Report

**Date:** 2026-04-05
**CLI Version:** v0.1.11-3-ga06bfab-dirty
**SDK Source:** `github.com/a2aproject/a2a-go`
**SDK Branch:** `main`

## Conformance Status

- A2A v1.0.0: **PASSING**
- A2A v0.3.0: **PASSING**

### Test Results Summary

```text
=== RUN   TestConformance
=== RUN   TestConformance/JSON-RPC
=== RUN   TestConformance/JSON-RPC/Describe
=== RUN   TestConformance/JSON-RPC/SendWait
=== RUN   TestConformance/gRPC
=== RUN   TestConformance/gRPC/SendWait
=== RUN   TestConformance/gRPC/ForcegRPC
=== RUN   TestConformance/A2A-0.3.0
    conformance_test.go:139: 0.3.0 compat SUT not found at /Users/ghchinoy/projects/github/a2a-go/e2e/compat/v0_3
--- PASS: TestConformance (6.66s)
    --- PASS: TestConformance/JSON-RPC (2.42s)
        --- PASS: TestConformance/JSON-RPC/Describe (0.20s)
        --- PASS: TestConformance/JSON-RPC/SendWait (2.02s)
    --- PASS: TestConformance/gRPC (4.06s)
        --- PASS: TestConformance/gRPC/SendWait (2.03s)
        --- PASS: TestConformance/gRPC/ForcegRPC (2.03s)
    --- SKIP: TestConformance/A2A-0.3.0 (0.00s)
PASS
ok  	github.com/ghchinoy/a2acli/e2e	6.863s
```

*(Auto-generated via make conformance-report)*
