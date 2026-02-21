#!/usr/bin/env bash
set -e

# A2A CLI Installation Script
# This script downloads and installs the latest a2acli binary.

REPO="ghchinoy/a2acli"
BINARY="a2acli"

# Helper for colorful output
echo_info() {
    echo -e "\033[1;34m==>\033[0m $1"
}
echo_err() {
    echo -e "\033[1;31mError:\033[0m $1" >&2
}

# Determine OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64)  ARCH="arm64" ;;
    aarch64) ARCH="arm64" ;;
    *) 
        echo_err "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# GitHub's latest release API
LATEST_URL="https://api.github.com/repos/${REPO}/releases/latest"

echo_info "Fetching latest release information for ${OS}/${ARCH}..."

# We need curl or wget
if command -v curl >/dev/null 2>&1; then
    RELEASE_DATA=$(curl -sL "$LATEST_URL")
elif command -v wget >/dev/null 2>&1; then
    RELEASE_DATA=$(wget -qO- "$LATEST_URL")
else
    echo_err "curl or wget is required to download a2acli."
    exit 1
fi

# Note: Since there are no published binary artifacts yet in GitHub Releases,
# this script assumes a standard naming convention for future releases:
# e.g., a2acli_darwin_arm64.tar.gz
TARBALL="${BINARY}_${OS}_${ARCH}.tar.gz"

DOWNLOAD_URL=$(echo "$RELEASE_DATA" | grep "browser_download_url" | grep "$TARBALL" | cut -d '"' -f 4 | head -n 1)

if [ -z "$DOWNLOAD_URL" ]; then
    echo_err "Could not find a pre-compiled binary for ${OS} ${ARCH}."
    echo_err "Please install via Go: go install github.com/${REPO}/cmd/a2acli@latest"
    exit 1
fi

echo_info "Downloading $DOWNLOAD_URL..."
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

if command -v curl >/dev/null 2>&1; then
    curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/$TARBALL"
else
    wget -qO "$TMP_DIR/$TARBALL" "$DOWNLOAD_URL"
fi

echo_info "Extracting..."
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"

INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    echo_info "Sudo required to install to $INSTALL_DIR"
    SUDO="sudo"
else
    SUDO=""
fi

echo_info "Installing to $INSTALL_DIR/$BINARY..."
$SUDO mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
$SUDO chmod +x "$INSTALL_DIR/$BINARY"

echo_info "Installation complete! Run 'a2acli --help' to get started."
