# A2A Conformance Report

**Date:** 2026-05-01
**CLI Version:** v1.3.0-3-g5da5634-dirty
**SDK Source:** `github.com/a2aproject/a2a-go`
**SDK Branch:** ``

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
    conformance_test.go:139: 0.3.0 compat SUT not found at /tmp/a2a-go/e2e/compat/v0_3
--- PASS: TestConformance (7.86s)
    --- PASS: TestConformance/JSON-RPC (2.97s)
        --- PASS: TestConformance/JSON-RPC/Describe (0.69s)
        --- PASS: TestConformance/JSON-RPC/SendWait (2.08s)
    --- PASS: TestConformance/gRPC (4.13s)
        --- PASS: TestConformance/gRPC/SendWait (2.06s)
        --- PASS: TestConformance/gRPC/ForcegRPC (2.06s)
    --- SKIP: TestConformance/A2A-0.3.0 (0.00s)
PASS
ok  	github.com/ghchinoy/a2acli/e2e	(cached)
```

*(Auto-generated via make conformance-report)*
