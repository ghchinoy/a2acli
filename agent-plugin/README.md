# a2acli — Agent Plugin

An [Agent Plugin](https://agent-plugins.org/) that teaches a plugin-aware AI agent
how to drive the [`a2acli`](https://github.com/ghchinoy/a2acli) command-line
client — discovering, messaging, and managing A2A (Agent2Agent) agents from the
command line.

It packages an `a2acli` Agent Skill as one installable unit for a plugin-aware
client. The skill and the binary can always be installed independently — the
plugin is a convenience, never a precondition.

## Contents

```text
agent-plugin/
├── plugin.json                 # manifest (Agent Plugins 1.0.0)
├── LICENSE                     # Apache-2.0
└── skills/
    └── a2acli/
        └── SKILL.md            # thin pointer to the canonical root skill
```

Skills-only: this plugin carries no MCP server, which is valid under Agent
Plugins (a skills-only plugin conforms).

## Requires the `a2acli` binary

Agent Plugins packages the agent-facing pieces (skills and MCP servers), not tool
binaries. The `a2acli` binary is **not** bundled and must be installed separately
(Homebrew, a prebuilt release, `install.sh`, or `go install`) — see the
[project README](https://github.com/ghchinoy/a2acli#installation). The skill
includes a preflight check (`a2acli version`).

## Relationship to the repository-root skills

This directory is the OFFICIAL-shaped, self-contained Agent Plugin package
(mirroring `a2aproject/a2a-cli`'s `agent-plugin/`). The repository also ships its
full set of Agent Skills from the root (`../plugin.json` + `../skills/`), which
remains the canonical, richer source (it carries `references/`, `scripts/`, and
additional skills for building and auditing A2A services).

**Single source of truth.** To keep two copies of the `a2acli` skill from
drifting, the `SKILL.md` bundled here is a deliberately thin pointer: it does
**not** restate the command index, flag tables, or usage rules. The authoritative,
maintained content lives once, in the repository-root skill at
[`../skills/a2acli/SKILL.md`](../skills/a2acli/SKILL.md). This plugin's `SKILL.md`
defers to that root skill and to `a2acli --help` for the current flag surface.
Update the root skill; leave this plugin copy as a pointer.

> Note on A2A SPEC §14.1 ("if a tool ships a skill, it MUST ship exactly one"):
> the repository currently ships three skills under `../skills/`. Whether to
> consolidate to one is an owner product decision, tracked in the alignment
> report (§2.9), and is intentionally **not** resolved by adding this package.

## Versioning

The plugin `version` in `plugin.json` tracks the `a2acli` release. The manifest
targets Agent Plugins spec **1.0.0**.
