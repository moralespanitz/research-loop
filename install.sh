#!/bin/sh
# Research Loop — one-line installer
# Usage: curl -fsSL https://research-loop.dev/install | sh
#
# Installs the research-loop binary to /usr/local/bin (or ~/.local/bin if no sudo).
# Set VERSION to pin a specific release: VERSION=v1.2.3 sh install.sh

set -e

REPO="research-loop/research-loop"
BINARY="research-loop"
INSTALL_DIR="/usr/local/bin"

# ─── Colors ──────────────────────────────────────────────────────────────────
BOLD="\033[1m"
GREEN="\033[32m"
YELLOW="\033[33m"
RED="\033[31m"
RESET="\033[0m"

step()  { printf "  ${BOLD}%s${RESET} %s\n" "→" "$1"; }
ok()    { printf "  ${GREEN}✓${RESET}  %s\n" "$1"; }
warn()  { printf "  ${YELLOW}⚠${RESET}  %s\n" "$1"; }
die()   { printf "  ${RED}✗${RESET}  %s\n" "$1"; exit 1; }

printf "\n  🔬  ${BOLD}Research Loop${RESET} installer\n\n"

# ─── Detect OS + arch ────────────────────────────────────────────────────────
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

# ─── Resolve version ─────────────────────────────────────────────────────────
if [ -n "${VERSION:-}" ]; then
  TAG="$VERSION"
  step "Pinned version: $TAG"
else
  step "Fetching latest release…"
  TAG=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' | cut -d'"' -f4)
  [ -n "$TAG" ] || die "Could not determine latest release. Set VERSION=vX.Y.Z to pin a specific version."
  step "Latest: $TAG"
fi

# ─── Install dir ─────────────────────────────────────────────────────────────
if [ -w "$INSTALL_DIR" ] || sudo -n true >/dev/null 2>&1; then
  : # can write to /usr/local/bin
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
  warn "No sudo — installing to $INSTALL_DIR (add to PATH if needed)"
fi

# ─── Download ────────────────────────────────────────────────────────────────
ASSET="${BINARY}_${OS}_${ARCH}"
BASE_URL="https://github.com/$REPO/releases/download/$TAG"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

step "Downloading $ASSET ($TAG)…"
curl -fsSL "$BASE_URL/$ASSET" -o "$TMP_DIR/$BINARY" \
  || die "Download failed: $BASE_URL/$ASSET"

# ─── Checksum ────────────────────────────────────────────────────────────────
step "Verifying checksum…"
curl -fsSL "$BASE_URL/checksums.txt" -o "$TMP_DIR/checksums.txt" \
  || die "Could not fetch checksums.txt from $BASE_URL"

EXPECTED=$(grep "[[:space:]]${ASSET}$" "$TMP_DIR/checksums.txt" | awk '{print $1}')
[ -n "$EXPECTED" ] || die "No checksum entry found for $ASSET in checksums.txt"

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL=$(sha256sum "$TMP_DIR/$BINARY" | cut -d' ' -f1)
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL=$(shasum -a 256 "$TMP_DIR/$BINARY" | cut -d' ' -f1)
else
  warn "No sha256 tool found — skipping checksum verification"
  ACTUAL="$EXPECTED"
fi

[ "$ACTUAL" = "$EXPECTED" ] || die "Checksum mismatch (got $ACTUAL, expected $EXPECTED)"
ok "Checksum verified"

# ─── Install ─────────────────────────────────────────────────────────────────
chmod +x "$TMP_DIR/$BINARY"
step "Installing to $INSTALL_DIR/$BINARY…"

if [ -w "$INSTALL_DIR" ]; then
  cp "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
  sudo cp "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

ok "Installed: $INSTALL_DIR/$BINARY"

# ─── Verify ──────────────────────────────────────────────────────────────────
if command -v research-loop >/dev/null 2>&1; then
  ok "research-loop is on PATH"
else
  warn "Add $INSTALL_DIR to your PATH:"
  printf "\n    export PATH=\"\$PATH:%s\"\n\n  Then reload your shell.\n\n" "$INSTALL_DIR"
fi

# ─── Done ────────────────────────────────────────────────────────────────────
printf "\n  ${GREEN}${BOLD}Installation complete.${RESET}\n\n"
printf "  Get started:\n\n"
printf "    ${BOLD}research-loop init${RESET}                     Initialize a workspace\n"
printf "    ${BOLD}research-loop dashboard --open${RESET}         Start dashboard at localhost:4321\n"
printf "    ${BOLD}research-loop start <arxiv-url>${RESET}        Ingest a paper\n\n"
printf "  Connect to Claude Code:\n\n"
printf "    ${BOLD}claude mcp add research-loop -- \$(which research-loop) mcp serve${RESET}\n\n"
printf "  Or just open your project — .mcp.json is already configured.\n\n"
