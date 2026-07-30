# A2A Conformance Report

**Date:** 2026-07-29
**CLI Version:** v1.8.1-6-g30c45e3-dirty
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
    conformance_test.go:205: 0.3.0 compat SUT not found at ../../github/a2a-go/e2e/compat/v0_3
=== RUN   TestConformance/A2UI-Extension-v1.0
    conformance_test.go:266: GOOGLE_CLOUD_PROJECT not set — skipping A2UI test (requires Vertex AI credentials)
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
--- PASS: TestConformance (12.29s)
    --- PASS: TestConformance/JSON-RPC (6.47s)
        --- PASS: TestConformance/JSON-RPC/Describe (0.19s)
        --- PASS: TestConformance/JSON-RPC/SendWait (2.02s)
        --- PASS: TestConformance/JSON-RPC/SendStdin (2.03s)
        --- PASS: TestConformance/JSON-RPC/ConformanceSmoke (2.03s)
    --- PASS: TestConformance/gRPC (4.06s)
        --- PASS: TestConformance/gRPC/SendWait (2.03s)
        --- PASS: TestConformance/gRPC/ForcegRPC (2.02s)
    --- SKIP: TestConformance/A2A-0.3.0 (0.00s)
    --- SKIP: TestConformance/A2UI-Extension-v1.0 (0.00s)
    --- PASS: TestConformance/A2A-Simple-MultiTransport (0.88s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/Discover (0.02s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/JSONRPC (0.02s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/REST (0.01s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/gRPC (0.01s)
    --- PASS: TestConformance/A2A-Simple-Multimodal (0.72s)
        --- PASS: TestConformance/A2A-Simple-Multimodal/ArtifactTypes (0.02s)
        --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates (0.03s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-completed (0.01s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-failed (0.01s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-input-required (0.01s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-auth-required (0.01s)
PASS
ok  	github.com/ghchinoy/a2acli/e2e	12.456s
```

*(Auto-generated via make conformance-report)*
