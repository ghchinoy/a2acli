#!/usr/bin/env bash
set -e

# A2A CLI Installation Script
# This script autonomously discovers, downloads, and installs the latest a2acli binary
# for your specific OS and architecture from GitHub Releases.
#
# Usage:
#   curl -sL https://raw.githubusercontent.com/ghchinoy/a2acli/main/scripts/install.sh | bash

REPO="ghchinoy/a2acli"
BINARY="a2acli"
INSTALL_DIR="/usr/local/bin"

# UI Helpers for a professional CLI experience
echo_info() {
    printf "\033[1;34m==>\033[0m %s\n" "$1"
}
echo_success() {
    printf "\033[1;32m==>\033[0m %s\n" "$1"
}
echo_err() {
    printf "\033[1;31mError:\033[0m %s\n" "$1" >&2
}

# 1. Environment Discovery
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) 
        echo_err "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# 2. Dependency Check
if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
    echo_err "curl or wget is required to download a2acli."
    exit 1
fi

if ! command -v tar >/dev/null 2>&1; then
    echo_err "tar is required to extract the binary."
    exit 1
fi

# 3. Fetch Latest Release Metadata
LATEST_URL="https://api.github.com/repos/${REPO}/releases/latest"
echo_info "Checking GitHub for the latest release of ${REPO}..."

if command -v curl >/dev/null 2>&1; then
    RELEASE_DATA=$(curl -sL "$LATEST_URL")
else
    RELEASE_DATA=$(wget -qO- "$LATEST_URL")
fi

TAG=$(echo "$RELEASE_DATA" | grep '"tag_name":' | cut -d '"' -f 4)
echo_info "Found version: ${TAG}"

# 4. Construct Artifact Name (Aligned with .goreleaser.yaml)
# GoReleaser uses TitleCase for OS (Darwin, Linux) and x86_64 for amd64
OS_TITLE="$(tr '[:lower:]' '[:upper:]' <<< ${OS:0:1})${OS:1}"
if [ "$ARCH" = "amd64" ]; then
    ARCH_MAP="x86_64"
else
    ARCH_MAP="$ARCH"
fi

TARBALL="${BINARY}_${OS_TITLE}_${ARCH_MAP}.tar.gz"
DOWNLOAD_URL=$(echo "$RELEASE_DATA" | grep "browser_download_url" | grep "$TARBALL" | cut -d '"' -f 4 | head -n 1)

if [ -z "$DOWNLOAD_URL" ]; then
    echo_err "Could not find a pre-compiled binary for ${OS_TITLE} (${ARCH_MAP})."
    echo_err "Please install via Go source instead: go install github.com/${REPO}/cmd/a2acli@latest"
    exit 1
fi

# 5. Download and Extract
echo_info "Downloading ${TARBALL}..."
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

if command -v curl >/dev/null 2>&1; then
    curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/$TARBALL"
else
    wget -qO "$TMP_DIR/$TARBALL" "$DOWNLOAD_URL"
fi

echo_info "Extracting binary..."
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"

# 6. Final Installation
if [ ! -w "$INSTALL_DIR" ]; then
    echo_info "Sudo privileges required to install to ${INSTALL_DIR}"
    SUDO="sudo"
else
    SUDO=""
fi

echo_info "Installing ${BINARY} to ${INSTALL_DIR}..."
$SUDO mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
$SUDO chmod +x "$INSTALL_DIR/$BINARY"

# 7. Verification & Handoff
if command -v "$BINARY" >/dev/null 2>&1; then
    echo_success "Installation successful!"
    echo "Run 'a2acli --help' to explore discovery, messaging, and tasks."
else
    echo_err "Installation completed, but '${BINARY}' is not in your PATH."
    echo_err "Please add ${INSTALL_DIR} to your PATH or run the binary directly from there."
fi
