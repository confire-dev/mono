#!/usr/bin/env bash
# Upload release artifacts to R2 for get.confire.dev / releases.confire.dev.
# Called from .github/workflows/release.yml (requires wrangler + Cloudflare secrets).
#
# Artifact format: confire_<version>_<os>_<arch>.tar.gz
# R2 key structure: <version>/confire_<version>_<os>_<arch>.tar.gz

set -euo pipefail

VERSION="${1:?usage: publish-release-r2.sh <tag e.g. v1.0.0> <dist-dir>}"
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

# Versioned archives + checksums
for ARCHIVE in "${DIST_DIR}"/confire_*.tar.gz; do
  BASENAME="$(basename "${ARCHIVE}")"
  put "${VERSION}/${BASENAME}" "${ARCHIVE}" "application/octet-stream"
done
put "${VERSION}/checksums.txt" "${DIST_DIR}/checksums.txt" "text/plain; charset=utf-8"

# latest.json
cat > "${DIST_DIR}/latest.json" <<EOF
{
  "version": "${VERSION}",
  "publishedAt": "${PUBLISHED_AT}",
  "releaseUrl": "https://github.com/confire-ai/mono/releases/tag/${VERSION}",
  "assets": {
    "confire_${VERSION}_darwin_amd64.tar.gz": "${RELEASES_URL}/${VERSION}/confire_${VERSION}_darwin_amd64.tar.gz",
    "confire_${VERSION}_darwin_arm64.tar.gz": "${RELEASES_URL}/${VERSION}/confire_${VERSION}_darwin_arm64.tar.gz",
    "confire_${VERSION}_linux_amd64.tar.gz":  "${RELEASES_URL}/${VERSION}/confire_${VERSION}_linux_amd64.tar.gz",
    "confire_${VERSION}_linux_arm64.tar.gz":  "${RELEASES_URL}/${VERSION}/confire_${VERSION}_linux_arm64.tar.gz",
    "checksums.txt": "${RELEASES_URL}/${VERSION}/checksums.txt"
  }
}
EOF

put "latest.json" "${DIST_DIR}/latest.json" "application/json; charset=utf-8"
put "${VERSION}/latest.json" "${DIST_DIR}/latest.json" "application/json; charset=utf-8"

# Installer (always latest)
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
put "install.sh" "${ROOT}/install.sh" "text/x-shellscript; charset=utf-8"

echo "Done. Verify:"
echo "  curl -fsSL https://get.confire.dev/latest.json"
echo "  curl -fsSL ${RELEASES_URL}/${VERSION}/checksums.txt"
