# Obtaining & Installing `a2acli` (Binary & Skills)

This reference guides coding agents and developers through verifying, installing, and updating both the `a2acli` binary and its associated Agent Skills / Agent Plugin package.

---

## 1. Step 0: Verification First

Before attempting installation, verify if `a2acli` is already available on `PATH`:

```bash
a2acli version
```

- **If output shows version (e.g. `a2acli version v1.10.0`)**: The binary is installed and ready.
- **If command fails (`command not found`)**: Proceed with binary installation below.

---

## 2. Installing the `a2acli` Binary

Choose the appropriate channel for your host environment:

### macOS and Linux — Homebrew
Recommended for macOS and Linux developer workstations:

```bash
brew tap ghchinoy/tap
brew trust ghchinoy/tap
brew install a2acli
```

### Linux — Install Script
Recommended for automated environments or general Linux systems:

```bash
curl -sL https://raw.githubusercontent.com/ghchinoy/a2acli/main/scripts/install.sh | bash
```

### Linux — apt (Debian / Ubuntu)
Download the `.deb` package from the [latest release](https://github.com/ghchinoy/a2acli/releases/latest):

```bash
sudo dpkg -i a2acli_*.deb
```

### Linux — rpm (Fedora / RHEL)
Download the `.rpm` package from the [latest release](https://github.com/ghchinoy/a2acli/releases/latest):

```bash
sudo rpm -i a2acli_*.rpm
```

### Windows — winget
Recommended for Windows environments:

```powershell
winget install ghchinoy.a2acli
```

### Cross-Platform — `go install`
Requires Go 1.25+:

```bash
go install github.com/ghchinoy/a2acli/cmd/a2acli@latest
```

Ensure `$(go env GOPATH)/bin` or `~/go/bin` is present in your system `PATH`.

### From Source
Build directly from the repository root:

```bash
git clone https://github.com/ghchinoy/a2acli.git
cd a2acli
make build    # Binary is written to ./bin/a2acli
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
