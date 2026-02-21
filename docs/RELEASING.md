# Releasing and Publishing `a2acli`

This project uses [GoReleaser](https://goreleaser.com/) via GitHub Actions to completely automate the building, packaging, and publishing of `a2acli` binaries.

## How it works

When you push a Git tag starting with `v` (e.g., `v1.0.0`), a GitHub Action is triggered (defined in `.github/workflows/release.yaml`).

This workflow spins up a runner that:
1. Compiles the Go code across multiple Operating Systems (macOS, Linux, Windows) and Architectures (amd64, arm64).
2. Uses Go Linker Flags (`-ldflags`) to dynamically inject the Git Tag, Commit SHA, and Build Date into the `a2acli version` command. (You do not need to manually edit `version.go`).
3. Packages the binaries into `.tar.gz` and `.zip` archives.
4. Auto-generates a Changelog based on the git commit history since the last tag.
5. Publishes all artifacts to a new release on the GitHub Repository Releases page.

Once the release is published, users can instantly use the curl installation script:
```bash
curl -sL https://raw.githubusercontent.com/ghchinoy/a2acli/main/scripts/install.sh | bash
```

## Step-by-Step Release Guide

Follow these steps to publish a new version of `a2acli`:

### 1. Ensure `main` is up to date and clean
Make sure all your code is pushed and your working directory is clean.
```bash
git checkout main
git pull
git status
```

### 2. Create a Semantic Version Tag
Create an annotated tag (or lightweight tag) with the new version number. It **must** start with a lowercase `v`.
```bash
git tag v1.0.0
```

### 3. Push the Tag to GitHub
Pushing the tag triggers the GoReleaser GitHub Action.
```bash
git push origin v1.0.0
```

### 4. Verify the Release
Navigate to your GitHub repository's **Releases** page or click the **Actions** tab to watch the GoReleaser workflow run.

Within ~2 minutes, the workflow will finish and the new binaries will be available for download.

## Local Testing (Optional)

If you want to test the release process locally *without* publishing to GitHub, you can run GoReleaser in snapshot mode:

```bash
# Requires goreleaser to be installed locally (brew install goreleaser)
goreleaser release --snapshot --clean
```

This will build all cross-platform binaries and place them in the local `dist/` directory for your inspection.