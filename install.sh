#!/usr/bin/env sh
# Confire installer
# Usage: curl -fsSL https://raw.githubusercontent.com/confire-ai/mono/main/install.sh | sh
# Or:    curl -fsSL https://raw.githubusercontent.com/confire-ai/mono/main/install.sh | sh -s -- --version v0.3.0

set -e

REPO="confire-ai/mono"
BINARY="confire"
INSTALL_DIR="${CONFIRE_INSTALL_DIR:-}"

# ── parse flags ───────────────────────────────────────────────────────────────
VERSION=""
while [ $# -gt 0 ]; do
  case "$1" in
    --version|-v) VERSION="$2"; shift 2 ;;
    --dir|-d)     INSTALL_DIR="$2"; shift 2 ;;
    --help|-h)
      echo "Usage: install.sh [--version <tag>] [--dir <install-dir>]"
      echo "  --version   Install a specific version (default: latest)"
      echo "  --dir       Install directory (default: /usr/local/bin or ~/.local/bin)"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

# ── detect OS and arch ────────────────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin) OS_NAME="darwin" ;;
  Linux)  OS_NAME="linux"  ;;
  *)
    echo "Unsupported OS: $OS" >&2
    echo "Confire supports macOS and Linux. Windows support is coming." >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64)    ARCH_NAME="amd64" ;;
  arm64|aarch64)   ARCH_NAME="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    echo "Confire supports x86_64 and arm64." >&2
    exit 1
    ;;
esac

PLATFORM="${OS_NAME}_${ARCH_NAME}"

# ── resolve version ───────────────────────────────────────────────────────────
if [ -z "$VERSION" ]; then
  # Fetch the latest release tag from GitHub API (no auth needed for public repos).
  LATEST_URL="https://api.github.com/repos/${REPO}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    VERSION="$(curl -fsSL "$LATEST_URL" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  elif command -v wget >/dev/null 2>&1; then
    VERSION="$(wget -qO- "$LATEST_URL" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  else
    echo "curl or wget is required." >&2
    exit 1
  fi
  if [ -z "$VERSION" ]; then
    echo "Could not determine latest version. Pass --version explicitly." >&2
    exit 1
  fi
fi

# ── pick install dir ──────────────────────────────────────────────────────────
if [ -z "$INSTALL_DIR" ]; then
  # Prefer /usr/local/bin if writable, otherwise ~/.local/bin
  if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  elif [ -w "/usr/bin" ]; then
    INSTALL_DIR="/usr/bin"
  else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
  fi
fi

# ── download ──────────────────────────────────────────────────────────────────
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
BINARY_NAME="${BINARY}_${PLATFORM}"
BINARY_URL="${BASE_URL}/${BINARY_NAME}"
CHECKSUM_URL="${BASE_URL}/confire_checksums.txt"
DEST="${INSTALL_DIR}/${BINARY}"
TMP_DIR="$(mktemp -d)"
TMP_BINARY="${TMP_DIR}/${BINARY}"
TMP_CHECKSUMS="${TMP_DIR}/checksums.txt"

printf "Installing confire %s for %s/%s...\n" "$VERSION" "$OS_NAME" "$ARCH_NAME"
printf "  from: %s\n" "$BINARY_URL"
printf "  to:   %s\n" "$DEST"

download() {
  URL="$1"
  OUT="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --progress-bar "$URL" -o "$OUT"
  else
    wget -q --show-progress "$URL" -O "$OUT"
  fi
}

download "$BINARY_URL"   "$TMP_BINARY"
download "$CHECKSUM_URL" "$TMP_CHECKSUMS"

# ── verify checksum ───────────────────────────────────────────────────────────
printf "Verifying checksum... "
EXPECTED="$(grep "${BINARY_NAME}" "$TMP_CHECKSUMS" | awk '{print $1}')"
if [ -z "$EXPECTED" ]; then
  echo "WARN: no checksum entry found for ${BINARY_NAME} — skipping verification." >&2
else
  if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL="$(sha256sum "$TMP_BINARY" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    ACTUAL="$(shasum -a 256 "$TMP_BINARY" | awk '{print $1}')"
  else
    ACTUAL=""
    echo "WARN: sha256sum/shasum not found — skipping checksum verification." >&2
  fi

  if [ -n "$ACTUAL" ]; then
    if [ "$ACTUAL" = "$EXPECTED" ]; then
      echo "OK"
    else
      echo "FAILED" >&2
      echo "  expected: $EXPECTED" >&2
      echo "  got:      $ACTUAL" >&2
      rm -rf "$TMP_DIR"
      exit 1
    fi
  fi
fi

# ── install ───────────────────────────────────────────────────────────────────
chmod +x "$TMP_BINARY"

# Atomic replace: move temp binary into place.
# Use sudo only if the install dir is not writable by current user.
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_BINARY" "$DEST"
else
  echo "Installing to $INSTALL_DIR requires sudo:"
  sudo mv "$TMP_BINARY" "$DEST"
fi

rm -rf "$TMP_DIR"

# ── verify install ────────────────────────────────────────────────────────────
if ! command -v confire >/dev/null 2>&1; then
  # Not in PATH — check if it was a non-standard install dir.
  if [ -x "$DEST" ]; then
    echo ""
    echo "confire installed to $DEST"
    echo ""
    echo "Add it to your PATH:"
    echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
    echo ""
    echo "Or add the above line to ~/.zshrc / ~/.bashrc and restart your shell."
  else
    echo "Installation failed — binary not found at $DEST." >&2
    exit 1
  fi
else
  INSTALLED_VERSION="$(confire version 2>/dev/null | head -1 || echo 'unknown')"
  echo ""
  echo "confire installed: $INSTALLED_VERSION"
  echo ""
  echo "Get started:"
  echo "  confire setup     install hooks into your AI agent"
  echo "  confire login     connect your account"
  echo "  confire status    show current state"
  echo ""
fi
