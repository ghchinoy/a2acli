# A2A Conformance Report

**Date:** 2026-09-29
**CLI Version:** v2.1.1-7-g6fcda79
**SDK Source:** `github.com/a2aproject/a2a-go`
**SDK Ref:** `v2.6.0`

## Conformance Status

- A2A v1.0.0: **PASSING** — against a2a-go's `e2e/tck` SUT
- A2A v0.3.0: **PASSING** — against the vendored `e2e/fixtures/v0_3_sut` SUT
- A2UI Extension v1.0: **NOT RUN** — against a2a-experiments' `cmd/a2ui`

`NOT RUN` means the suite skipped because its system-under-test or its
credentials were unavailable in the environment that generated this report.
It is not a pass.

### Test Results Summary

```text
=== RUN   TestCredentialFlagsOnTheWire
=== RUN   TestCredentialFlagsOnTheWire/BearerFlag
=== RUN   TestCredentialFlagsOnTheWire/APIKeyFlagDefaultHeader
=== RUN   TestCredentialFlagsOnTheWire/BearerEnv
=== RUN   TestCredentialFlagsOnTheWire/APIKeyEnv
=== RUN   TestCredentialFlagsOnTheWire/FlagOverridesEnv
--- PASS: TestCredentialFlagsOnTheWire (1.18s)
    --- PASS: TestCredentialFlagsOnTheWire/BearerFlag (0.01s)
    --- PASS: TestCredentialFlagsOnTheWire/APIKeyFlagDefaultHeader (0.01s)
    --- PASS: TestCredentialFlagsOnTheWire/BearerEnv (0.01s)
    --- PASS: TestCredentialFlagsOnTheWire/APIKeyEnv (0.01s)
    --- PASS: TestCredentialFlagsOnTheWire/FlagOverridesEnv (0.01s)
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
=== RUN   TestConformance/A2A-0.3.0/Describe
=== RUN   TestConformance/A2A-0.3.0/SendWait
=== RUN   TestConformance/A2UI-Extension-v1.0
    conformance_test.go:374: skipping A2UI extension e2e test: GOOGLE_CLOUD_PROJECT and GOOGLE_CLOUD_LOCATION environment variables must be set
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
--- PASS: TestConformance (14.24s)
    --- PASS: TestConformance/JSON-RPC (6.07s)
        --- PASS: TestConformance/JSON-RPC/Describe (0.01s)
        --- PASS: TestConformance/JSON-RPC/SendWait (2.03s)
        --- PASS: TestConformance/JSON-RPC/SendStdin (2.01s)
        --- PASS: TestConformance/JSON-RPC/ConformanceSmoke (2.01s)
    --- PASS: TestConformance/gRPC (4.03s)
        --- PASS: TestConformance/gRPC/SendWait (2.01s)
        --- PASS: TestConformance/gRPC/ForcegRPC (2.01s)
    --- PASS: TestConformance/A2A-0.3.0 (0.85s)
        --- PASS: TestConformance/A2A-0.3.0/Describe (0.01s)
        --- PASS: TestConformance/A2A-0.3.0/SendWait (0.01s)
    --- SKIP: TestConformance/A2UI-Extension-v1.0 (0.00s)
    --- PASS: TestConformance/A2A-Simple-MultiTransport (1.11s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/Discover (0.01s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/JSONRPC (0.01s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/REST (0.01s)
        --- PASS: TestConformance/A2A-Simple-MultiTransport/gRPC (0.01s)
    --- PASS: TestConformance/A2A-Simple-Multimodal (0.81s)
        --- PASS: TestConformance/A2A-Simple-Multimodal/ArtifactTypes (0.01s)
        --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates (0.04s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-completed (0.01s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-failed (0.01s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-input-required (0.01s)
            --- PASS: TestConformance/A2A-Simple-Multimodal/TaskStates/state-auth-required (0.01s)
    --- PASS: TestConformance/JourneySuites (1.14s)
        --- PASS: TestConformance/JourneySuites/PositionalURLDiscover (0.01s)
        --- PASS: TestConformance/JourneySuites/ZeroArgValidation (0.01s)
        --- PASS: TestConformance/JourneySuites/ContextContinuity (0.02s)
        --- PASS: TestConformance/JourneySuites/TerminalTaskStrict (0.02s)
        --- PASS: TestConformance/JourneySuites/ListTasksColumns (0.02s)
        --- PASS: TestConformance/JourneySuites/DirectoryGuard (0.01s)
=== RUN   TestTier1CLIContract
=== RUN   TestTier1CLIContract/UsageErrorsExit2
=== RUN   TestTier1CLIContract/UsageErrorsExit2/UnknownCommand
=== RUN   TestTier1CLIContract/UsageErrorsExit2/UnknownFlag
=== RUN   TestTier1CLIContract/DefaultSendIsSingleJSONDoc
=== RUN   TestTier1CLIContract/StreamSendIsJSONL
--- PASS: TestTier1CLIContract (1.43s)
    --- PASS: TestTier1CLIContract/UsageErrorsExit2 (0.03s)
        --- PASS: TestTier1CLIContract/UsageErrorsExit2/UnknownCommand (0.01s)
        --- PASS: TestTier1CLIContract/UsageErrorsExit2/UnknownFlag (0.01s)
    --- PASS: TestTier1CLIContract/DefaultSendIsSingleJSONDoc (0.01s)
    --- PASS: TestTier1CLIContract/StreamSendIsJSONL (0.01s)
PASS
ok  	github.com/ghchinoy/a2acli/e2e	16.844s
```

*(Auto-generated via make conformance-report)*
