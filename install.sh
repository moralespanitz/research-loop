#!/bin/bash
# Research Loop — one-line installer
# Usage: curl -fsSL https://research-loop.dev/install | sh
#
# Installs the research-loop binary to /usr/local/bin (or ~/bin if no sudo).

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

echo ""
echo "  🔬  ${BOLD}Research Loop${RESET} installer"
echo ""

# ─── Detect OS + arch ────────────────────────────────────────────────────────
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) die "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) die "Unsupported OS: $OS. Please build from source: go build ./cmd/research-loop" ;;
esac

step "Detected: $OS/$ARCH"

# ─── Check for Go (build from source path) ───────────────────────────────────
if command -v go &>/dev/null; then
  GO_VERSION=$(go version | awk '{print $3}' | tr -d 'go')
  ok "Go $GO_VERSION found — will build from source"
  BUILD_FROM_SOURCE=1
else
  BUILD_FROM_SOURCE=0
fi

# ─── Install dir ─────────────────────────────────────────────────────────────
if [ -w "$INSTALL_DIR" ] || sudo -n true 2>/dev/null; then
  : # can write to /usr/local/bin
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
  warn "No sudo — installing to $INSTALL_DIR (add to PATH if needed)"
fi

# ─── Build or download ───────────────────────────────────────────────────────
if [ "$BUILD_FROM_SOURCE" = "1" ]; then
  step "Building from source (github.com/$REPO)…"
  TMP_DIR=$(mktemp -d)
  trap "rm -rf $TMP_DIR" EXIT

  if command -v git &>/dev/null; then
    git clone --depth 1 "https://github.com/$REPO.git" "$TMP_DIR/src" 2>/dev/null
    cd "$TMP_DIR/src"
    go build -o "$TMP_DIR/$BINARY" ./cmd/research-loop
  else
    die "git not found. Install git and try again."
  fi

  BINARY_PATH="$TMP_DIR/$BINARY"
else
  # ── Download pre-built binary ───────────────────────────────────────────────
  # (Releases not yet published — this path is for future use)
  step "Downloading pre-built binary…"
  TAG=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | grep '"tag_name"' | cut -d'"' -f4)
  if [ -z "$TAG" ]; then
    die "Could not find a release. Please install Go and re-run: the installer will build from source."
  fi

  DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/${BINARY}-${OS}-${ARCH}"
  TMP_FILE=$(mktemp)
  trap "rm -f $TMP_FILE" EXIT

  curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE" || die "Download failed: $DOWNLOAD_URL"
  chmod +x "$TMP_FILE"
  BINARY_PATH="$TMP_FILE"
fi

# ─── Install ─────────────────────────────────────────────────────────────────
step "Installing to $INSTALL_DIR/$BINARY…"

if [ -w "$INSTALL_DIR" ]; then
  cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY"
else
  sudo cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"
ok "Installed: $INSTALL_DIR/$BINARY"

# ─── Verify ──────────────────────────────────────────────────────────────────
if command -v research-loop &>/dev/null; then
  ok "research-loop is on PATH"
else
  warn "Add $INSTALL_DIR to your PATH:"
  echo ""
  echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
  echo ""
  echo "  Then reload your shell."
fi

# ─── Done ────────────────────────────────────────────────────────────────────
echo ""
echo "  ${GREEN}${BOLD}Installation complete.${RESET}"
echo ""
echo "  Get started:"
echo ""
echo "    ${BOLD}research-loop init${RESET}                     Initialize a workspace"
echo "    ${BOLD}research-loop dashboard --open${RESET}         Start dashboard at localhost:4321"
echo "    ${BOLD}research-loop start <arxiv-url>${RESET}        Ingest a paper"
echo ""
echo "  Connect to Claude Code:"
echo ""
echo "    ${BOLD}claude mcp add research-loop -- \$(which research-loop) mcp serve${RESET}"
echo ""
echo "  Or just open your project — .mcp.json is already configured."
echo ""
