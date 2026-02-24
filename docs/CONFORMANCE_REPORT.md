# A2A Conformance Report

**Date:** 2026-02-24
**CLI Version:** v0.1.10-2-gf84a311-dirty
**SDK Source:** `github.com/ghchinoy/a2a-go`
**SDK Branch:** `fix/tck-v1-initial-event`

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
=== RUN   TestConformance/A2A-0.3.0/Describe
=== RUN   TestConformance/A2A-0.3.0/SendWait
--- PASS: TestConformance (10.13s)
    --- PASS: TestConformance/JSON-RPC (3.12s)
        --- PASS: TestConformance/JSON-RPC/Describe (0.65s)
        --- PASS: TestConformance/JSON-RPC/SendWait (2.07s)
    --- PASS: TestConformance/gRPC (4.13s)
        --- PASS: TestConformance/gRPC/SendWait (2.06s)
        --- PASS: TestConformance/gRPC/ForcegRPC (2.07s)
    --- PASS: TestConformance/A2A-0.3.0 (2.09s)
        --- PASS: TestConformance/A2A-0.3.0/Describe (0.07s)
        --- PASS: TestConformance/A2A-0.3.0/SendWait (0.02s)
PASS
ok  	github.com/ghchinoy/a2acli/e2e	(cached)
```

*(Auto-generated via make conformance-report)*
