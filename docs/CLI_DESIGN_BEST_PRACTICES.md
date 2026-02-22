# CLI Design Best Practices for Humans and AI Agents

This document outlines the principles for building modern, discoverable, and self-documenting command-line interfaces (CLIs). These practices ensure that tools are equally usable by human developers and automated coding agents.

## 1. Structured Discoverability

The entry point of a CLI should clearly map out its capabilities.

*   **Command Grouping**: Use groups (e.g., Cobra's `GroupID`) to categorize commands (e.g., `Task Management`, `Information`, `Configuration`). This prevents a "wall of text" in the root help output.
*   **Semantic Help**: Implement visual hierarchy in help text. Section headers (Examples, Flags, Commands) should be distinct from the content.
*   **Entry Point Guidance**: Explicitly mark common entry points in help descriptions (e.g., `"(start here)"` or `"(typical first step)"`).

## 2. The Three Pillars of Command Documentation

Every command in the CLI should populate these three fields:

| Field | Purpose | Best Practice |
| :--- | :--- | :--- |
| **Short** | Quick scan | One-line summary (5-10 words) starting with an action verb. |
| **Long** | Deep understanding | Detailed explanation of *what* the command does, *why* it's used, and how it differs from similar commands. |
| **Example** | Immediate utility | 3-5 concrete copy-pasteable examples showing common flag combinations. |

## 3. Agent-First Interoperability

To be "agentic," a CLI must be parseable and predictable.

*   **Deterministic Output**: Every command that produces data must support a `--json` or `--no-tui` flag. The output should be valid JSON/NDJSON that an agent can parse into internal structures.
*   **Environment Overrides**: Support `NO_COLOR` and `[APP]_NO_TUI` environment variables to allow agents to bypass interactive elements automatically.
*   **Non-Interactive Fallbacks**: Commands that use TUIs (like Bubble Tea) should have a "plain" mode that emits standard text or JSON to `stdout`.

## 4. Robust Configuration & Context

A CLI should understand its environment without requiring excessive flags.

*   **XDG Compliance**: Follow the XDG Base Directory Specification for config files (`~/.config/app/config.yaml`).
*   **Smart Fallbacks**: Flags should have a priority order: `Command Line Flag > Environment Variable > Config File > Default Value`.
*   **Auditability**: Every write operation should track the "Actor" (e.g., derived from `git config user.name` or `$USER`).

## 5. Proactive Error Guidance

Don't just report an error; provide a path to resolution.

*   **Contextual Hints**: When a command fails due to a missing prerequisite, include a "Hint:" line suggesting the fix (e.g., `"Hint: run 'app init' to create a database"`).
*   **Fail Fast**: Validate configuration and connectivity in a `PersistentPreRun` hook before executing heavy logic.

## 6. Flag & Argument Consistency

*   **Predictable Patterns**: If `-o` means `--out-dir` in one command, it should mean the same thing across the entire application.
*   **Sane Defaults**: Defaults should favor the "safe" or "most common" path.
*   **Positional vs. Named**: Use positional arguments for required IDs/entities and flags for optional modifiers.

## 7. Visual Information Design (Tufte-Inspired)

CLIs should prioritize high-density information display while minimizing cognitive load.

*   **Maximize Data-Ink Ratio**: Minimize non-essential "ink" (decorative borders, unnecessary colors). Reserve color for elements that demand attention or convey state.
*   **Whitespace over Color**: Use positioning and whitespace as the primary tool for conveying hierarchy. Reserve color for exceptional states (errors, warnings) or high-value scan targets.
*   **Semantic Color Tokens**: Use meaning-based tokens (e.g., `Accent`, `Muted`, `Pass`, `Warn`, `Fail`) rather than raw color names. This ensures semantic consistency across the application.
*   **Standard Semantic Palette (Ayu-based)**: Use the following hex codes (aligned with the `ayu` theme) to ensure cross-tool consistency and high contrast in both terminal modes.

| Token | Visual Intent | Light Hex | Dark Hex | Role |
| :--- | :--- | :--- | :--- | :--- |
| `Accent` | Landmarks | `#399ee6` | `#59c2ff` | Headers, Group Titles, Section Labels |
| `Command` | Scan Targets | `#5c6166` | `#bfbdb6` | Command names, Flags |
| `Pass` | Success | `#86b300` | `#c2d94c` | Completed tasks, Success states |
| `Warn` | Transient | `#f2ae49` | `#ffb454` | Active tasks, Warnings, Pending states |
| `Fail` | Error | `#f07171` | `#f07178` | Failed tasks, Errors, Rejected states |
| `Muted` | De-emphasis | `#828c99` | `#6c7680` | Metadata, Types, Defaults, Previews |
| `ID` | Identifiers | `#46ba94` | `#95e6cb` | Unique IDs (TaskIDs, SkillIDs) |

*   **Perceptual Optimization**: Ensure output is optimized for both **Light and Dark** terminal backgrounds. Use adaptive colors that maintain contrast and semantic meaning regardless of the user's terminal theme.

## 8. Functional Color Usage

Color should serve a specific functional purpose, not an aesthetic one.

*   **Scan Targets**: Use an `Accent` or `Command` color for navigation landmarks (headers) and actionable items (command and flag names).
*   **State Indicators**: Use standard semantic colors for states: `Pass` (Green), `Warn` (Yellow/Orange), `Fail` (Red).
*   **De-emphasis**: Use `Muted` (Grey) for low-priority information like type annotations, default values, and supplemental metadata.
*   **Avoid Over-Coloring**: Do not color descriptions, help examples, or every item in a list. Over-coloring leads to "rainbow output" where nothing stands out.

## 9. Versioning and Lifecycle

*   **Self-Tracking**: The CLI should know its own version and optionally notify the user if a significant version jump has occurred since the last run.
*   **Clean Exit**: Ensure graceful handling of `SIGINT` (Ctrl+C) to prevent data corruption, especially when writing to local databases or files.
