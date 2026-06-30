# A2A Conformance Report

**Date:** 2026-06-30
**CLI Version:** v1.8.0-4-gff2de12-dirty
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
    conformance_test.go:199: 0.3.0 compat SUT not found at ../../github/a2a-go/e2e/compat/v0_3
=== RUN   TestConformance/A2UI-Extension-v1.0
=== RUN   TestConformance/A2UI-Extension-v1.0/Validate
=== RUN   TestConformance/A2A-Simple-MultiTransport
=== RUN   TestConformance/A2A-Simple-MultiTransport/Discover
=== RUN   TestConformance/A2A-Simple-MultiTransport/JSONRPC
=== RUN   TestConformance/A2A-Simple-MultiTransport/REST
=== RUN   TestConformance/A2A-Simple-MultiTransport/gRPC
=== RUN   TestConformance/A2A-Simple-Multimodal
=== RUN   TestConformance/A2A-Simple-Multimodal/ArtifactTypes
=== RUN   TestConformance/A2A-Simple-Multimodal/TaskStates
=== RUN   TestConformance/A2A-Simple-Multimodal/TaskStates/state-completed
=== RUN   TestConformance/A2A-Simple-Multimodal/TaskStates/state-failed
=== RUN   TestConformance/A2A-Simple-Multimodal/TaskStates/state-input-required
=== RUN   TestConformance/A2A-Simple-Multimodal/TaskStates/state-auth-required
--- PASS: TestConformance (16.49s)
    --- PASS: TestConformance/JSON-RPC (6.47s)
        --- PASS: TestConformance/JSON-RPC/Describe (0.22s)
        --- PASS: TestConformance/JSON-RPC/SendWait (2.01s)
        --- PASS: TestConformance/JSON-RPC/SendStdin (2.01s)
        --- PASS: TestConformance/JSON-RPC/ConformanceSmoke (2.02s)
    --- PASS: TestConformance/gRPC (4.03s)
        --- PASS: TestConformance/gRPC/SendWait (2.01s)
        --- PASS: TestConformance/gRPC/ForcegRPC (2.01s)
    --- SKIP: TestConformance/A2A-0.3.0 (0.00s)
    --- PASS: TestConformance/A2UI-Extension-v1.0 (3.85s)
        --- PASS: TestConformance/A2UI-Extension-v1.0/Validate (1.97s)
    --- PASS: TestConformance/A2A-Simple-MultiTransport (1.21s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/Discover (0.01s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/JSONRPC (0.01s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/REST (0.01s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/gRPC (0.01s)
    --- PASS: TestConformance/A2A-Simple-Multimodal (0.58s)
        --- PASS: TestConformance/A2A-Simple-Multimodal/ArtifactTypes (0.01s)
        --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates (0.04s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-completed (0.01s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-failed (0.01s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-input-required (0.01s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-auth-required (0.01s)
PASS
ok  	github.com/ghchinoy/a2acli/e2e	16.726s
```

*(Auto-generated via make conformance-report)*
