# A2A Conformance Report

**Date:** 2026-08-04
**CLI Version:** v1.9.0-12-g137826a-dirty
**SDK Source:** `unknown`
**SDK Branch:** `unknown`

## Conformance Status

- A2A v1.0.0: **PASSING**
- A2A v0.3.0: **PASSING**
- A2UI Extension v1.0: **PASSING**

### Test Results Summary

```text
=== RUN   TestConformance
=== RUN   TestConformance/JSON-RPC
    conformance_test.go:96: a2a-go SDK source not found at ../../github/a2a-go/e2e/tck
=== RUN   TestConformance/gRPC
    conformance_test.go:180: a2a-go SDK source not found at ../../github/a2a-go/e2e/tck
=== RUN   TestConformance/A2A-0.3.0
    conformance_test.go:213: 0.3.0 compat SUT not found at ../../github/a2a-go/e2e/compat/v0_3
=== RUN   TestConformance/A2UI-Extension-v1.0
    conformance_test.go:271: skipping A2UI extension e2e test: GOOGLE_CLOUD_PROJECT and GOOGLE_CLOUD_LOCATION environment variables must be set
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
=== RUN   TestConformance/JourneySuites
=== RUN   TestConformance/JourneySuites/PositionalURLDiscover
=== RUN   TestConformance/JourneySuites/ZeroArgValidation
=== RUN   TestConformance/JourneySuites/ContextContinuity
=== RUN   TestConformance/JourneySuites/TerminalTaskStrict
=== RUN   TestConformance/JourneySuites/ListTasksColumns
=== RUN   TestConformance/JourneySuites/DirectoryGuard
--- PASS: TestConformance (2.62s)
    --- SKIP: TestConformance/JSON-RPC (0.00s)
    --- SKIP: TestConformance/gRPC (0.00s)
    --- SKIP: TestConformance/A2A-0.3.0 (0.00s)
    --- SKIP: TestConformance/A2UI-Extension-v1.0 (0.00s)
    --- PASS: TestConformance/A2A-Simple-MultiTransport (0.65s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/Discover (0.03s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/JSONRPC (0.02s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/REST (0.02s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/gRPC (0.02s)
    --- PASS: TestConformance/A2A-Simple-Multimodal (0.53s)
        --- PASS: TestConformance/A2A-Simple-Multimodal/ArtifactTypes (0.03s)
        --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates (0.07s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-completed (0.02s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-failed (0.02s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-input-required (0.02s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-auth-required (0.02s)
    --- PASS: TestConformance/JourneySuites (0.74s)
        --- PASS: TestConformance/JourneySuites/PositionalURLDiscover (0.04s)
        --- PASS: TestConformance/JourneySuites/ZeroArgValidation (0.02s)
        --- PASS: TestConformance/JourneySuites/ContextContinuity (0.04s)
        --- PASS: TestConformance/JourneySuites/TerminalTaskStrict (0.03s)
        --- PASS: TestConformance/JourneySuites/ListTasksColumns (0.04s)
        --- PASS: TestConformance/JourneySuites/DirectoryGuard (0.02s)
PASS
ok  	github.com/ghchinoy/a2acli/e2e	2.620s
```

*(Auto-generated via make conformance-report)*
