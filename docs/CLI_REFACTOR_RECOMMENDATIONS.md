# CLI Refactor Recommendations: a2acli

This document tracks the specific improvements to the `a2acli` tool based on an assessment against modern CLI best practices and the `bd` reference implementation.

## Assessment & Strategy Table

| Category | Priority | Recommendation | Impact | Status |
| :--- | :--- | :--- | :--- | :--- |
| **Command Organization** | 🟢 High | Use `rootCmd.AddGroup` and `GroupID` to categorize commands. | Improved structure for agents/humans. | ✅ Done |
| **Descriptive Depth** | 🟢 High | Add `Long` descriptions to every command. | Better agent understanding. | ✅ Done |
| **Usage Examples** | 🟢 High | Add `Example` strings with 2-3 concrete cases. | Reduced guesswork. | ✅ Done |
| **Semantic Help** | 🟡 Medium | Implement basic colorized help via lipgloss. | Better visual hierarchy. | ✅ Done |
| **Error Guidance** | 🟡 Medium | Standardize "Hint:" messages in error outputs. | Faster user recovery. | ✅ Done |
| **Flag Consistency** | 🔵 Low | Audit all flags for consistent short-names. | Predictable API. | ✅ Done |

## Refactor Scorecard (Final)

| Category | Score | Observations |
| :--- | :---: | :--- |
| **Command Organization** | 🟢 | Grouped into Discovery, Messaging, and System (A2A Aligned). |
| **Descriptive Depth** | 🟢 | Added detailed `Long` descriptions to all commands. |
| **Usage Examples** | 🟢 | Added comprehensive `Example` blocks to all commands. |
| **Agent Friendliness** | 🟢 | Retained `--no-tui` and added short-names (`-n`, `-V`). |
| **Error Guidance** | 🟢 | Standardized "Hint:" messages via `fatalf` helper. |
| **Flag Consistency** | 🟢 | Audited and synchronized short-names across all commands. |

## Implementation Roadmap (Completed)

The following tasks were tracked in the `bd` issue tracker:

1.  **a2ac-cni**: Organized a2acli commands into logical groups (Discovery, Messaging, System).
2.  **a2ac-a8v**: Added 'Long' descriptions and 'Example' blocks to all core commands.
3.  **a2ac-5rv**: Implemented basic semantic help styling in `cmd/a2acli/help.go`.
4.  **a2ac-59j**: Standardized 'Hint' guidance for common error states.
5.  **a2ac-eh1**: Audited and synchronized flag short-names.
6.  **a2ac-udz**: Implemented dynamic transport selection (gRPC vs JSON-RPC).
