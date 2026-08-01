# A2A Agent Skills & Agent Plugin

This directory contains [`agentskills.io`](https://agentskills.io/) compliant agent skills designed for AI coding agents (Claude Code, OpenCode, Cursor, Gemini CLI, Copilot CLI). The repository is a conformant [`Agent Plugins Specification v1.0.0`](https://agent-plugins.org) plugin package (`plugin.json`).

## Available Skills

- [**`a2acli`**](a2acli/SKILL.md) — Drive `a2acli` to discover agents, send tasks, stream results, and manage authentication.
- [**`a2a-expose`**](a2a-expose/SKILL.md) — Add an A2A Protocol v1.0 exposure layer to an existing API or codebase using a design-first workflow.
- [**`a2a-conformance`**](a2a-conformance/SKILL.md) — Evaluate any A2A agent for spec compliance using static review, `a2acli` probes, and official TCK orchestration.

## Quick Install

```bash
# Install individual skill using Open Agent Skills CLI
npx skills add ghchinoy/a2acli --skill a2acli
npx skills add ghchinoy/a2acli --skill a2a-expose
npx skills add ghchinoy/a2acli --skill a2a-conformance

# Install all skills
npx skills add ghchinoy/a2acli --all
```

For complete documentation, scenario walkthroughs, prompt libraries, and authoring guidelines, see the [**Agent Skills Guide (`docs/SKILLS.md`)**](../docs/SKILLS.md).
