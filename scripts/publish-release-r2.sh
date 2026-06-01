#!/usr/bin/env bash
# Upload release artifacts to R2 for get.confire.dev / releases.confire.dev.
# Called from .github/workflows/release.yml (requires wrangler + Cloudflare secrets).

set -euo pipefail

VERSION="${1:?usage: publish-release-r2.sh <tag e.g. v0.3.0> <dist-dir>}"
DIST_DIR="${2:?usage: publish-release-r2.sh <tag> <dist-dir>}"
BUCKET="${CONFIRE_RELEASES_BUCKET:-confire-releases}"
RELEASES_URL="${CONFIRE_RELEASES_URL:-https://releases.confire.dev}"

if [ ! -d "$DIST_DIR" ]; then
  echo "dist dir not found: $DIST_DIR" >&2
  exit 1
fi

PUBLISHED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

put() {
  local key="$1"
  local file="$2"
  local type="${3:-application/octet-stream}"
  echo "  → r2://${BUCKET}/${key}"
  wrangler r2 object put "${BUCKET}/${key}" --file="$file" --content-type="$type" --remote
}

echo "Publishing ${VERSION} to R2 bucket ${BUCKET}..."

# Versioned binaries + checksums
for bin in confire_darwin_amd64 confire_darwin_arm64 confire_linux_amd64 confire_linux_arm64; do
  put "${VERSION}/${bin}" "${DIST_DIR}/${bin}" "application/octet-stream"
done
put "${VERSION}/confire_checksums.txt" "${DIST_DIR}/confire_checksums.txt" "text/plain; charset=utf-8"

# latest.json manifest (curl installer + platform API)
cat > "${DIST_DIR}/latest.json" <<EOF
{
  "version": "${VERSION}",
  "publishedAt": "${PUBLISHED_AT}",
  "releaseUrl": "https://github.com/confire-ai/mono/releases/tag/${VERSION}",
  "assets": {
    "confire_darwin_amd64": "${RELEASES_URL}/${VERSION}/confire_darwin_amd64",
    "confire_darwin_arm64": "${RELEASES_URL}/${VERSION}/confire_darwin_arm64",
    "confire_linux_amd64": "${RELEASES_URL}/${VERSION}/confire_linux_amd64",
    "confire_linux_arm64": "${RELEASES_URL}/${VERSION}/confire_linux_arm64",
    "confire_checksums.txt": "${RELEASES_URL}/${VERSION}/confire_checksums.txt"
  }
}
EOF

put "latest.json" "${DIST_DIR}/latest.json" "application/json; charset=utf-8"
put "${VERSION}/latest.json" "${DIST_DIR}/latest.json" "application/json; charset=utf-8"

# Installer script (always latest from repo root)
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
put "install.sh" "${ROOT}/install.sh" "text/x-shellscript; charset=utf-8"

echo "Done. Verify:"
echo "  curl -fsSL https://get.confire.dev/latest.json"
echo "  curl -fsSL https://releases.confire.dev/${VERSION}/confire_checksums.txt"
