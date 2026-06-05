#!/usr/bin/env sh
# Confire installer — get.confire.dev edition
#
# Usage:
#   curl -fsSL https://get.confire.dev/install.sh | sh
#   curl -fsSL https://get.confire.dev/install.sh | sh -s -- --version v1.0.0
#
# Options:
#   --version <tag>   Install a specific version (default: latest)
#   --dir <path>      Override install directory

set -e

BINARY="confire"
GET_URL="${CONFIRE_GET_URL:-https://get.confire.dev}"
RELEASES_URL="${CONFIRE_RELEASES_URL:-https://releases.confire.dev}"
PLATFORM_URL="${CONFIRE_PLATFORM_URL:-https://confire.dev}"

VERSION=""
INSTALL_DIR="${CONFIRE_INSTALL_DIR:-}"

while [ $# -gt 0 ]; do
  case "$1" in
    --version|-v) VERSION="$2"; shift 2 ;;
    --dir|-d)     INSTALL_DIR="$2"; shift 2 ;;
    --help|-h)
      printf "Usage: install.sh [--version <tag>] [--dir <install-dir>]\n"
      printf "  --version   Specific version to install (default: latest)\n"
      printf "  --dir       Override install directory\n"
      exit 0
      ;;
    *) printf "Unknown option: %s\n" "$1" >&2; exit 1 ;;
  esac
done

# ── License notice ──────────────────────────────────────────────────────────
printf "\n"
printf "Confire is proprietary software. By installing, you agree to the license terms.\n"
printf "See: %s/terms\n\n" "$PLATFORM_URL"

# ── Platform detection ───────────────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin) OS_NAME="darwin" ;;
  Linux)  OS_NAME="linux"  ;;
  *)
    printf "Unsupported OS: %s\n" "$OS" >&2
    printf "Confire supports macOS and Linux.\n" >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64)  ARCH_NAME="amd64" ;;
  arm64|aarch64) ARCH_NAME="arm64" ;;
  *)
    printf "Unsupported architecture: %s\n" "$ARCH" >&2
    printf "Confire supports x86_64 (amd64) and arm64.\n" >&2
    exit 1
    ;;
esac

# ── Resolve version ──────────────────────────────────────────────────────────
if [ -z "$VERSION" ]; then
  LATEST_URL="${GET_URL}/latest.json"
  if command -v curl >/dev/null 2>&1; then
    VERSION="$(curl -fsSL "$LATEST_URL" | sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)"
  elif command -v wget >/dev/null 2>&1; then
    VERSION="$(wget -qO- "$LATEST_URL" | sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)"
  else
    printf "curl or wget is required.\n" >&2
    exit 1
  fi
  if [ -z "$VERSION" ]; then
    printf "Could not determine latest version. Pass --version explicitly.\n" >&2
    exit 1
  fi
fi

# ── Install directory ────────────────────────────────────────────────────────
if [ -z "$INSTALL_DIR" ]; then
  if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
  fi
fi

# ── Artifact paths ───────────────────────────────────────────────────────────
# confire_v1.0.0_darwin_arm64.tar.gz
ARCHIVE_NAME="${BINARY}_${VERSION}_${OS_NAME}_${ARCH_NAME}.tar.gz"
BASE_URL="${RELEASES_URL}/${VERSION}"
ARCHIVE_URL="${BASE_URL}/${ARCHIVE_NAME}"
CHECKSUM_URL="${BASE_URL}/checksums.txt"
BUNDLE_URL="${BASE_URL}/checksums.txt.bundle"
DEST="${INSTALL_DIR}/${BINARY}"
TMP_DIR="$(mktemp -d)"

printf "Installing Confire %s (%s/%s)\n" "$VERSION" "$OS_NAME" "$ARCH_NAME"
printf "  from: %s\n" "$ARCHIVE_URL"
printf "  to:   %s\n\n" "$DEST"

# ── Download helpers ─────────────────────────────────────────────────────────
_download() {
  URL="$1"; OUT="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --progress-bar "$URL" -o "$OUT"
  else
    wget -q --show-progress "$URL" -O "$OUT"
  fi
}

_download_optional() {
  URL="$1"; OUT="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$URL" -o "$OUT" 2>/dev/null || true
  else
    wget -q "$URL" -O "$OUT" 2>/dev/null || true
  fi
}

# ── Download archive and checksums ───────────────────────────────────────────
_download "$ARCHIVE_URL"   "${TMP_DIR}/${ARCHIVE_NAME}"
_download "$CHECKSUM_URL"  "${TMP_DIR}/checksums.txt"
_download_optional "$BUNDLE_URL" "${TMP_DIR}/checksums.txt.bundle"

# ── Verify checksum ───────────────────────────────────────────────────────────
printf "Verifying checksum... "
EXPECTED="$(grep "${ARCHIVE_NAME}" "${TMP_DIR}/checksums.txt" | awk '{print $1}')"
if [ -z "$EXPECTED" ]; then
  printf "ERROR: no checksum entry found for %s in checksums.txt\n" "$ARCHIVE_NAME" >&2
  printf "Aborting installation — do not install unverified binaries.\n" >&2
  rm -rf "$TMP_DIR"
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "${TMP_DIR}/${ARCHIVE_NAME}" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL="$(shasum -a 256 "${TMP_DIR}/${ARCHIVE_NAME}" | awk '{print $1}')"
else
  printf "ERROR: sha256sum or shasum not found.\n" >&2
  printf "Aborting — checksum verification is required.\n" >&2
  rm -rf "$TMP_DIR"
  exit 1
fi

if [ "$ACTUAL" != "$EXPECTED" ]; then
  printf "FAILED\n" >&2
  printf "  expected: %s\n" "$EXPECTED" >&2
  printf "  got:      %s\n" "$ACTUAL" >&2
  rm -rf "$TMP_DIR"
  exit 1
fi
printf "OK\n"

# ── Verify cosign bundle (optional) ──────────────────────────────────────────
if [ -s "${TMP_DIR}/checksums.txt.bundle" ] && command -v cosign >/dev/null 2>&1; then
  printf "Verifying cosign bundle... "
  if cosign verify-blob \
       --bundle "${TMP_DIR}/checksums.txt.bundle" \
       --certificate-identity-regexp "https://github.com/confire-ai/mono" \
       --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
       "${TMP_DIR}/checksums.txt" >/dev/null 2>&1; then
    printf "OK\n"
  else
    printf "FAILED\n" >&2
    printf "cosign bundle verification failed — aborting.\n" >&2
    rm -rf "$TMP_DIR"
    exit 1
  fi
else
  if [ ! -s "${TMP_DIR}/checksums.txt.bundle" ]; then
    printf "Note: checksums.txt.bundle not available; skipping cosign verification.\n"
  fi
fi

# ── Extract binary ────────────────────────────────────────────────────────────
tar -xzf "${TMP_DIR}/${ARCHIVE_NAME}" -C "$TMP_DIR"
TMP_BINARY="${TMP_DIR}/${BINARY}"
if [ ! -f "$TMP_BINARY" ]; then
  printf "ERROR: binary not found in archive after extraction.\n" >&2
  rm -rf "$TMP_DIR"
  exit 1
fi
chmod +x "$TMP_BINARY"

# ── Install ───────────────────────────────────────────────────────────────────
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_BINARY" "$DEST"
else
  printf "Installing to %s requires elevated privileges:\n" "$INSTALL_DIR"
  sudo mv "$TMP_BINARY" "$DEST"
fi

rm -rf "$TMP_DIR"

# ── Post-install ──────────────────────────────────────────────────────────────
if ! command -v confire >/dev/null 2>&1; then
  if [ -x "$DEST" ]; then
    printf "\nconfire installed to %s\n" "$DEST"
    printf "\nAdd it to your PATH:\n"
    printf "  export PATH=\"\$PATH:%s\"\n\n" "$INSTALL_DIR"
  else
    printf "Installation failed — binary not found at %s.\n" "$DEST" >&2
    exit 1
  fi
else
  INSTALLED_VERSION="$(confire version 2>/dev/null | head -1 || echo 'unknown')"
  printf "\n%s\n\n" "$INSTALLED_VERSION"
  confire setup || true
  confire start
fi
