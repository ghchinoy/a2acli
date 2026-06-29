# Releasing and Publishing `a2acli`

This project uses [GoReleaser](https://goreleaser.com/) via GitHub Actions to automate building, packaging, and publishing across all distribution channels.

## What happens on release

When you push a Git tag starting with `v` (e.g. `v1.2.0`), the release workflow (`.github/workflows/release.yaml`) runs GoReleaser, which:

1. Compiles binaries for macOS, Linux, and Windows on amd64 and arm64.
2. Injects the Git tag, commit SHA, and build date into `a2acli version` via linker flags.
3. Packages binaries into `.tar.gz` (Linux/macOS) and `.zip` (Windows) archives.
4. Generates `.deb` and `.rpm` packages for Linux.
5. Publishes all artifacts to a new GitHub Release with an auto-generated changelog.
6. Pushes an updated Homebrew formula to [`ghchinoy/homebrew-tap`](https://github.com/ghchinoy/homebrew-tap).
7. Opens a PR against [`microsoft/winget-pkgs`](https://github.com/microsoft/winget-pkgs) with updated manifests.

## Required secrets

| Secret | Used for |
|---|---|
| `GITHUB_TOKEN` | GitHub Release publishing (automatic) |
| `HOMEBREW_TAP_GITHUB_TOKEN` | Pushing formula to `ghchinoy/homebrew-tap` |
| `WINGET_GITHUB_TOKEN` | Pushing manifests to `ghchinoy/winget-pkgs` fork |

## SDK version alignment

a2acli and its sister projects (a2a-simple, syntaxis, …) all depend on
`github.com/a2aproject/a2a-go/v2`. Keep versions in sync across repos — the SDK
has had breaking API changes between minor versions (e.g. v2.2.0 → v2.3.1 changed
the push-notification config API: `CreateTaskPushConfigRequest` was replaced by
passing `*a2a.PushConfig` directly with `TaskID` embedded). A version skew between
a2acli and a test-target server will produce confusing errors at the transport or
JSON-RPC layer, not obvious compile errors.

Check before releasing: `grep a2a-go go.mod` and compare with sister projects.

## CIMD document stability

`docs/metadata.json` is served via GitHub Pages at
`https://ghchinoy.github.io/a2acli/metadata.json` and is a2acli's OAuth 2.1
client identity document (`client_id`). **Do not change the `client_id` URL** —
it is registered with OAuth consent SPAs that display the client name and logo.
If you need to change its content (e.g. add a redirect URI), update the JSON and
redeploy via a normal commit; the URL itself must remain stable across releases.

## Step-by-step release guide

### 1. Ensure `main` is clean and up to date

```bash
git checkout main
git pull
git status   # must be clean
```

### 2. Run conformance tests

```bash
make test-e2e A2A_GO_SRC=/path/to/a2a-go
```

### 3. Update the conformance report

```bash
make conformance-report A2A_GO_SRC=/path/to/a2a-go
git add docs/CONFORMANCE_REPORT.md
git commit -m "docs: update conformance report for vX.Y.Z"
git push
```

### 4. Tag and push

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

GoReleaser triggers automatically. Watch progress at the [Actions tab](https://github.com/ghchinoy/a2acli/actions).

### 5. Verify distribution channels

After the workflow completes (~2 min), verify each channel:

| Channel | Check |
|---|---|
| GitHub Release | New release appears at `/releases` with all artifacts |
| Homebrew | `brew update && brew upgrade a2acli` works |
| Linux deb/rpm | `.deb` and `.rpm` files present on the release page |
| winget | PR opened at `microsoft/winget-pkgs` (auto-merged within hours for updates) |

## Local testing (snapshot mode)

Test the GoReleaser config locally without publishing:

```bash
# Requires goreleaser installed locally
brew install goreleaser

goreleaser release --snapshot --clean
```

Artifacts are written to `dist/` for inspection.
