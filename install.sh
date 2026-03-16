#!/bin/sh
# Research Loop — one-line installer
# Usage: curl -fsSL https://raw.githubusercontent.com/research-loop/research-loop/main/install.sh | sh
#    or: wget -qO- https://raw.githubusercontent.com/research-loop/research-loop/main/install.sh | sh
#
# Environment variables:
#   VERSION   — install a specific version (e.g. VERSION=v0.2.0)

set -eu

REPO="research-loop/research-loop"
BINARY="research-loop"
INSTALL_DIR="/usr/local/bin"

# --- Colors ----------------------------------------------------------------
BOLD="\033[1m"
GREEN="\033[32m"
YELLOW="\033[33m"
RED="\033[31m"
RESET="\033[0m"

step()  { printf "  ${BOLD}->  %s${RESET}\n" "$1"; }
ok()    { printf "  ${GREEN}ok${RESET}  %s\n" "$1"; }
warn()  { printf "  ${YELLOW}!!${RESET}  %s\n" "$1"; }
die()   { printf "  ${RED}ERR${RESET} %s\n" "$1"; exit 1; }

# --- Detect HTTP client ----------------------------------------------------
if command -v curl >/dev/null 2>&1; then
  fetch()    { curl -fsSL "$1"; }
  download() { curl -fsSL "$1" -o "$2"; }
elif command -v wget >/dev/null 2>&1; then
  fetch()    { wget -qO- "$1"; }
  download() { wget -qO "$2" "$1"; }
else
  die "curl or wget required"
fi

printf "\n"
printf "  ${BOLD}Research Loop${RESET} installer\n"
printf "\n"

# --- Detect OS + arch ------------------------------------------------------
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) die "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) die "Unsupported OS: $OS" ;;
esac

step "Detected: $OS/$ARCH"

# --- Resolve version -------------------------------------------------------
if [ -n "${VERSION:-}" ]; then
  TAG="$VERSION"
  step "Using requested version: $TAG"
else
  step "Fetching latest release..."
  TAG=$(fetch "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null \
    | grep '"tag_name"' | cut -d'"' -f4) || true
  if [ -z "$TAG" ]; then
    die "No releases found. Check https://github.com/$REPO/releases"
  fi
  ok "Latest version: $TAG"
fi

# --- Install dir -----------------------------------------------------------
if [ -w "$INSTALL_DIR" ]; then
  : # writable
elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
  : # sudo available without password
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
  warn "No sudo — installing to $INSTALL_DIR (add to PATH if needed)"
fi

# --- Download + verify -----------------------------------------------------
ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/$ARCHIVE"
CHECKSUMS_URL="https://github.com/$REPO/releases/download/$TAG/checksums.txt"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

step "Downloading $ARCHIVE..."
download "$DOWNLOAD_URL" "$TMP_DIR/$ARCHIVE" \
  || die "Download failed: $DOWNLOAD_URL"

step "Verifying checksum..."
download "$CHECKSUMS_URL" "$TMP_DIR/checksums.txt" \
  || die "Could not download checksums"

EXPECTED=$(grep "$ARCHIVE" "$TMP_DIR/checksums.txt" | awk '{print $1}')
if [ -z "$EXPECTED" ]; then
  die "Archive not found in checksums.txt"
fi

ACTUAL=$(sha256sum "$TMP_DIR/$ARCHIVE" 2>/dev/null || shasum -a 256 "$TMP_DIR/$ARCHIVE" 2>/dev/null)
ACTUAL=$(echo "$ACTUAL" | awk '{print $1}')

if [ "$EXPECTED" != "$ACTUAL" ]; then
  die "Checksum mismatch (expected $EXPECTED, got $ACTUAL)"
fi
ok "Checksum verified"

# --- Extract + install -----------------------------------------------------
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

step "Installing to $INSTALL_DIR/$BINARY..."
if [ -w "$INSTALL_DIR" ]; then
  cp "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
  sudo cp "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi
chmod +x "$INSTALL_DIR/$BINARY"
ok "Installed: $INSTALL_DIR/$BINARY"

# --- Verify on PATH -------------------------------------------------------
if command -v research-loop >/dev/null 2>&1; then
  ok "research-loop is on PATH"
else
  warn "Add $INSTALL_DIR to your PATH:"
  printf "\n    export PATH=\"\$PATH:%s\"\n\n" "$INSTALL_DIR"
fi

# --- Done ------------------------------------------------------------------
printf "\n"
printf "  ${GREEN}${BOLD}Installation complete.${RESET}\n"
printf "\n"
printf "  Get started:\n"
printf "\n"
printf "    ${BOLD}research-loop init${RESET}                     Initialize a workspace\n"
printf "    ${BOLD}research-loop dashboard --open${RESET}         Start dashboard\n"
printf "    ${BOLD}research-loop start <arxiv-url>${RESET}        Ingest a paper\n"
printf "\n"
