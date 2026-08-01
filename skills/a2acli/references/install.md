# Obtaining & Installing `a2acli` (Binary & Skills)

This reference guides coding agents and developers through verifying, installing, and updating both the `a2acli` binary and its associated Agent Skills / Agent Plugin package.

---

## 1. Step 0: Verification First

Before attempting installation, verify if `a2acli` is already available on `PATH`:

```bash
a2acli version
```

- **If output shows version (e.g. `a2acli version v1.8.2`)**: The binary is installed and ready.
- **If command fails (`command not found`)**: Proceed with binary installation below.

---

## 2. Installing the `a2acli` Binary

Choose the appropriate channel for your host environment:

### Option A: Homebrew (macOS / Linux)
Recommended for macOS and Linux developer workstations:

```bash
brew install ghchinoy/tap/a2acli
```

### Option B: `go install` (Cross-Platform / Go Environments)
Requires Go 1.25+:

```bash
go install github.com/ghchinoy/a2acli/cmd/a2acli@latest
```

Ensure `$GOPATH/bin` or `~/go/bin` is present in your system `PATH`.

### Option C: GitHub Release Binaries (CI / Non-Go Environments)
Pre-compiled binaries for macOS, Linux, and Windows (amd64 / arm64) are attached to every GitHub release at `https://github.com/ghchinoy/a2acli/releases/latest`:

```bash
# Example Linux x86_64 download:
curl -sSL https://github.com/ghchinoy/a2acli/releases/latest/download/a2acli_Linux_x86_64.tar.gz | tar -xz
mv a2acli /usr/local/bin/
```

### Option D: Build from Source
From the cloned repository root:

```bash
make build    # Outputs binary to ./bin/a2acli
```

---

## 3. Installing the Agent Skills & Agent Plugin Package

This repository ships three [`agentskills.io`](https://agentskills.io/) compliant skills and is a conformant [`Agent Plugins Specification v1.0.0`](https://agent-plugins.org) package (`plugin.json`).

### Option A: Open Agent Skills CLI (`npx skills`)
Recommended for OpenCode, Claude Code, Cursor, Gemini CLI, and Copilot CLI:

```bash
# Discover skills in the repository
npx skills add ghchinoy/a2acli --list

# Install all skills globally across all projects
npx skills add ghchinoy/a2acli --all -g

# Or install a specific skill to current project (./.claude/skills/ or ./.opencode/skill/)
npx skills add ghchinoy/a2acli --skill a2acli
npx skills add ghchinoy/a2acli --skill a2a-expose
npx skills add ghchinoy/a2acli --skill a2a-conformance
```

### Option B: Agent Plugins v1.0 Conformant Client
Any Agent-Plugins-compliant client (e.g. `~/.agents/plugins/`) loads the root `plugin.json` directly from the repository root:

```text
my-workspace/
└── plugins/
    └── a2acli/        # Cloned or downloaded repo root
        ├── plugin.json
        └── skills/
            ├── a2acli/
            ├── a2a-expose/
            └── a2a-conformance/
```

---

## 4. Verification & Fail-Safe Rule

After installation:

1. Re-run `a2acli version` to confirm binary availability.
2. If `a2acli` cannot be installed or remains unavailable on `PATH`, the coding agent MUST **fail cleanly** with an explicit error message to the user rather than improvising unverified shell commands.
