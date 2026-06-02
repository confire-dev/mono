#!/usr/bin/env sh
# Local dev installer — builds confire from source (default) or pulls from the
# local distribution worker, then installs a wrapper pointed at local services.
#
# Usage:
#   ./install_local.sh                    # build from cli/ and install confire-local
#   ./install_local.sh --from-distribution --seed   # build dist/ + install (recommended)
#   ./install_local.sh --from-distribution --http   # test HTTP via distribution worker
#   ./install_local.sh --dir ~/.local/bin
#
# Prerequisites (run in separate terminals):
#   pnpm worker:dev          → API worker  http://localhost:8787
#   cd platform && pnpm dev  → platform    http://localhost:4321
#   cd distribution && pnpm dev  → only for --from-distribution --http (port 8789)
#
# The installed `confire-local` binary talks to local worker + platform via env vars
# and persists worker_url in ~/.confire/config.json.

set -e

ROOT="$(cd "$(dirname "$0")" && pwd)"

WORKER_URL="${CONFIRE_WORKER_URL:-http://localhost:8787}"
PLATFORM_URL="${CONFIRE_PLATFORM_URL:-http://localhost:4321}"
DIST_URL="${CONFIRE_DIST_URL:-http://localhost:8789}"
INSTALL_DIR="${CONFIRE_INSTALL_DIR:-}"
WRAPPER_NAME="${CONFIRE_LOCAL_BIN:-confire-local}"
VERSION=""
MODE="build"
SEED=0
USE_HTTP=0

while [ $# -gt 0 ]; do
  case "$1" in
    --from-distribution|-D) MODE="distribution"; shift ;;
    --http)                 USE_HTTP=1; shift ;;
    --seed)                 SEED=1; shift ;;
    --version|-v)           VERSION="$2"; shift 2 ;;
    --dir|-d)                 INSTALL_DIR="$2"; shift 2 ;;
    --worker-url)             WORKER_URL="$2"; shift 2 ;;
    --platform-url)           PLATFORM_URL="$2"; shift 2 ;;
    --dist-url)               DIST_URL="$2"; shift 2 ;;
    --help|-h)
      cat <<'EOF'
Usage: install_local.sh [options]

Install a local dev build of confire that targets local worker + platform URLs.

Options:
  --from-distribution, -D  Install from dist/ artifacts (see --seed)
  --http                   With -D: download via distribution worker HTTP instead of dist/
  --seed                   With -D: run scripts/seed-local-distribution.sh first
  --version, -v <tag>      Version tag for distribution mode (default: dev-local)
  --dir, -d <path>         Install wrapper directory (default: ~/.local/bin)
  --worker-url <url>       Worker API base (default: http://localhost:8787)
  --platform-url <url>     Platform base (default: http://localhost:4321)
  --dist-url <url>         Distribution worker base (default: http://localhost:8789)

Environment:
  CONFIRE_WORKER_URL, CONFIRE_PLATFORM_URL, CONFIRE_DIST_URL, CONFIRE_INSTALL_DIR

After install:
  confire-local setup
  confire-local login
EOF
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

OS="$(uname -s)"
ARCH="$(uname -m)"
case "$OS" in
  Darwin) OS_NAME="darwin" ;;
  Linux)  OS_NAME="linux"  ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac
case "$ARCH" in
  x86_64|amd64)  ARCH_NAME="amd64" ;;
  arm64|aarch64) ARCH_NAME="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac
PLATFORM="${OS_NAME}_${ARCH_NAME}"

if [ -z "$INSTALL_DIR" ]; then
  INSTALL_DIR="${HOME}/.local/bin"
fi
mkdir -p "$INSTALL_DIR"

CONFIRE_BIN_DIR="${HOME}/.confire/bin"
mkdir -p "$CONFIRE_BIN_DIR"
REAL_BINARY="${CONFIRE_BIN_DIR}/confire-dev"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if [ "$MODE" = "build" ]; then
  printf "Building dev binary from cli/...\n"
  # Build directly to destination — avoids macOS provenance xattr that security
  # tools set on cp'd files, which causes immediate SIGKILL on some systems.
  (cd "$ROOT/cli" && make dev-install DEST="$REAL_BINARY")
else
  if [ "$SEED" = "1" ]; then
    "$ROOT/scripts/seed-local-distribution.sh" "${VERSION:-dev-local}"
  fi

  TAG="${VERSION:-dev-local}"
  BINARY_NAME="confire_${PLATFORM}"
  DIST_BIN="${ROOT}/dist/${BINARY_NAME}"
  DIST_SUM="${ROOT}/dist/confire_checksums.txt"
  TMP_BINARY="${TMP_DIR}/confire"
  TMP_CHECKSUMS="${TMP_DIR}/checksums.txt"

  verify_checksum() {
    EXPECTED="$(grep "${BINARY_NAME}" "$TMP_CHECKSUMS" | awk '{print $1}')"
    [ -z "$EXPECTED" ] && return 0
    if command -v sha256sum >/dev/null 2>&1; then
      ACTUAL="$(sha256sum "$TMP_BINARY" | awk '{print $1}')"
    else
      ACTUAL="$(shasum -a 256 "$TMP_BINARY" | awk '{print $1}')"
    fi
    if [ "$ACTUAL" != "$EXPECTED" ]; then
      echo "Checksum mismatch for ${BINARY_NAME}" >&2
      exit 1
    fi
  }

  distribution_hint() {
    echo "" >&2
    echo "Local distribution setup:" >&2
    echo "  ./scripts/seed-local-distribution.sh ${TAG}" >&2
    echo "  ./install_local.sh --from-distribution --version ${TAG}" >&2
    echo "" >&2
    echo "Or in one step:" >&2
    echo "  ./install_local.sh --from-distribution --seed --version ${TAG}" >&2
  }

  if [ "$USE_HTTP" = "1" ]; then
    BINARY_URL="${DIST_URL}/${TAG}/${BINARY_NAME}"
    CHECKSUM_URL="${DIST_URL}/${TAG}/confire_checksums.txt"
    printf "Downloading %s via local distribution HTTP...\n" "$TAG"
    printf "  from: %s\n" "$BINARY_URL"

    download() {
      URL="$1"
      OUT="$2"
      if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$URL" -o "$OUT"
      else
        wget -q "$URL" -O "$OUT"
      fi
    }

    if ! curl -fsSL "${DIST_URL}/latest.json" >/dev/null 2>&1; then
      echo "Distribution server not reachable at ${DIST_URL}." >&2
      echo "Start: cd distribution && pnpm dev" >&2
      distribution_hint
      exit 1
    fi
    download "$BINARY_URL" "$TMP_BINARY"
    download "$CHECKSUM_URL" "$TMP_CHECKSUMS"
  elif [ -f "$DIST_BIN" ] && [ -f "$DIST_SUM" ]; then
    printf "Installing %s from dist/...\n" "$TAG"
    printf "  binary: %s\n" "$DIST_BIN"
    cp "$DIST_BIN" "$TMP_BINARY"
    cp "$DIST_SUM" "$TMP_CHECKSUMS"
  else
    echo "No dist/ artifacts for ${BINARY_NAME}." >&2
    distribution_hint
    exit 1
  fi

  verify_checksum
  cp "$TMP_BINARY" "$REAL_BINARY"
fi

chmod +x "$REAL_BINARY"

WRAPPER="${INSTALL_DIR}/${WRAPPER_NAME}"
cat > "$WRAPPER" <<EOF
#!/usr/bin/env sh
# Installed by install_local.sh — local dev endpoints
export CONFIRE_API="${WORKER_URL}"
export CONFIRE_PLATFORM_URL="${PLATFORM_URL}"
exec "${REAL_BINARY}" "\$@"
EOF
chmod +x "$WRAPPER"

# Persist worker URL for commands that read ~/.confire/config.json
if command -v python3 >/dev/null 2>&1; then
  python3 - "$WORKER_URL" <<'PY'
import json, os, sys
path = os.path.expanduser("~/.confire/config.json")
cfg = {}
if os.path.exists(path):
    with open(path, encoding="utf-8") as f:
        cfg = json.load(f)
cfg["worker_url"] = sys.argv[1]
os.makedirs(os.path.dirname(path), exist_ok=True)
with open(path, "w", encoding="utf-8") as f:
    json.dump(cfg, f, indent=2)
    f.write("\n")
os.chmod(path, 0o600)
PY
fi

INSTALLED_VERSION="$("$REAL_BINARY" version 2>/dev/null | head -1 || echo dev)"

printf "\nInstalled local dev CLI:\n"
printf "  wrapper:  %s\n" "$WRAPPER"
printf "  binary:   %s\n" "$REAL_BINARY"
printf "  version:  %s\n" "$INSTALLED_VERSION"
printf "  worker:   %s\n" "$WORKER_URL"
printf "  platform: %s\n" "$PLATFORM_URL"
printf "\nMake sure local services are running:\n"
printf "  pnpm worker:dev\n"
printf "  cd platform && pnpm dev\n"
if [ "$MODE" = "distribution" ] && [ "$USE_HTTP" = "1" ]; then
  printf "  cd distribution && pnpm dev   # %s\n" "$DIST_URL"
fi
if ! command -v "$WRAPPER_NAME" >/dev/null 2>&1; then
  printf "\nAdd to PATH:\n"
  printf "  export PATH=\"\$PATH:%s\"\n" "$INSTALL_DIR"
fi
printf "\n"

# Run setup interactively (shows scope + agent selector).
"$WRAPPER" setup || true

# Always start the daemon, even if setup was cancelled.
"$WRAPPER" start
