#!/usr/bin/env bash
# Seed the local R2 bucket used by `wrangler dev` in distribution/.
# Writes to distribution/.wrangler/state — same persistence wrangler dev reads.

set -euo pipefail

VERSION="${1:-dev-local}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST_DIR="${ROOT}/dist"
DIST_PKG="${ROOT}/distribution"
BUCKET="${CONFIRE_RELEASES_BUCKET:-confire-releases}"
DIST_URL="${CONFIRE_DIST_URL:-http://127.0.0.1:8789}"
PERSIST_DIR="${DIST_PKG}/.wrangler/state"
WRANGLER=(pnpm exec wrangler)

OS="$(uname -s)"
ARCH="$(uname -m)"
case "$OS" in
  Darwin) GOOS="darwin" ;;
  Linux)  GOOS="linux"  ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac
case "$ARCH" in
  x86_64|amd64)  GOARCH="amd64" ;;
  arm64|aarch64) GOARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

SUFFIX="${GOOS}_${GOARCH}"
BINARY="confire_${SUFFIX}"

printf "Building %s for local distribution seed...\n" "$BINARY"
mkdir -p "$DIST_DIR" "$PERSIST_DIR"
(
  cd "$ROOT/cli"
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
    go build -trimpath -ldflags "-s -w" -o "$DIST_DIR/$BINARY" .
)
(cd "$DIST_DIR" && sha256sum "$BINARY" > confire_checksums.txt)

PUBLISHED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
cat > "$DIST_DIR/latest.json" <<EOF
{
  "version": "${VERSION}",
  "publishedAt": "${PUBLISHED_AT}",
  "releaseUrl": "${DIST_URL}/${VERSION}",
  "assets": {
    "${BINARY}": "${DIST_URL}/${VERSION}/${BINARY}",
    "confire_checksums.txt": "${DIST_URL}/${VERSION}/confire_checksums.txt"
  }
}
EOF

put_local() {
  local key="$1"
  local file="$2"
  echo "  → r2://${BUCKET}/${key} (local, persist ${PERSIST_DIR})"
  (cd "$DIST_PKG" && "${WRANGLER[@]}" r2 object put "${BUCKET}/${key}" \
    --file="$file" --local --persist-to "$PERSIST_DIR")
}

(
  put_local "${VERSION}/${BINARY}" "$DIST_DIR/$BINARY"
  put_local "${VERSION}/confire_checksums.txt" "$DIST_DIR/confire_checksums.txt"
  put_local "latest.json" "$DIST_DIR/latest.json"
  put_local "${VERSION}/latest.json" "$DIST_DIR/latest.json"
  put_local "install.sh" "$ROOT/install.sh"
)

printf "Done. Start distribution dev server, then install:\n"
printf "  cd distribution && pnpm dev\n"
printf "  ./install_local.sh --from-distribution --version %s\n" "$VERSION"
