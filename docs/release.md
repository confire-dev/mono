# Confire Release Guide

## Build-time version injection

All version metadata is injected at build time using Go's `-ldflags`. The
variables live in `cmd/version.go` (package `github.com/confire-dev/confire/cmd`).

### Manual build (single platform)

```sh
go build \
  -ldflags "\
    -X 'github.com/confire-dev/confire/cmd.buildVersion=v1.0.0' \
    -X 'github.com/confire-dev/confire/cmd.buildCommit=$(git rev-parse --short HEAD)' \
    -X 'github.com/confire-dev/confire/cmd.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' \
    -X 'github.com/confire-dev/confire/cmd.buildBuiltBy=manual'" \
  -trimpath \
  -o dist/confire \
  ./cli
```

Run from the repo root. Adjust `-o dist/confire` for the target OS/arch.

### Makefile targets (run from `cli/`)

| Target | Description |
|---|---|
| `make dev` | Fast local build, symbols retained |
| `make dist` | Cross-compile linux/darwin × amd64/arm64 into `dist/` |
| `make release` | Hardened single-platform build (garble + stripped) |
| `make snapshot` | GoReleaser local preview (no publish) |
| `make test` | Unit + integration tests |
| `make clean` | Remove built binaries |

### GoReleaser (preferred for releases)

```sh
cd cli
goreleaser release --snapshot --clean   # local preview
goreleaser release --clean              # production (requires git tag)
```

Install: `go install github.com/goreleaser/goreleaser/v2@latest`

Config: `cli/.goreleaser.yaml`

---

## Artifact naming convention

Release archives follow this pattern:

```
confire_<version>_<os>_<arch>.tar.gz
```

Examples:

```
confire_v1.0.0_darwin_arm64.tar.gz
confire_v1.0.0_darwin_amd64.tar.gz
confire_v1.0.0_linux_arm64.tar.gz
confire_v1.0.0_linux_amd64.tar.gz
```

Each release also produces:

```
checksums.txt          SHA-256 of all archives
checksums.txt.bundle   cosign-signed bundle of checksums.txt
```

---

## Public release artifact layout

Artifacts are served from `get.confire.dev` (Cloudflare R2 via `distribution/`):

```
https://get.confire.dev/install.sh
https://get.confire.dev/latest.json

https://get.confire.dev/releases/v1.0.0/checksums.txt
https://get.confire.dev/releases/v1.0.0/checksums.txt.bundle
https://get.confire.dev/releases/v1.0.0/confire_v1.0.0_darwin_arm64.tar.gz
https://get.confire.dev/releases/v1.0.0/confire_v1.0.0_darwin_amd64.tar.gz
https://get.confire.dev/releases/v1.0.0/confire_v1.0.0_linux_arm64.tar.gz
https://get.confire.dev/releases/v1.0.0/confire_v1.0.0_linux_amd64.tar.gz
```

`latest.json` must be updated on each release:

```json
{ "version": "v1.0.0" }
```

Upload script: `scripts/publish-release-r2.sh <tag> <dist-dir>`

---

## Homebrew tap

The Homebrew formula lives in two places:

| File | Purpose |
|---|---|
| `homebrew/confire.rb` | Live formula (current artifact URLs, pushed to tap) |
| `packaging/homebrew/Formula/confire.rb` | Template with tar.gz URLs and updated format |

After each release:
1. Update `version` and `sha256` values in `homebrew/confire.rb`.
2. Push the updated formula to `github.com/confire-ai/homebrew-confire`.

The tap repo can be **public** even though the source repo is private. Homebrew
only downloads the pre-built binary archives from `get.confire.dev`, never the
source code.

Install command for users:

```sh
brew tap confire-ai/confire
brew install confire
```

---

## Signing with cosign (optional)

If `cosign` is installed, GoReleaser will sign `checksums.txt` and produce
`checksums.txt.bundle`. The installer (`scripts/install.sh`) verifies this
bundle if both `cosign` and the bundle are present.

```sh
# Install cosign
go install github.com/sigstore/cosign/v2/cmd/cosign@latest

# Manual verification
cosign verify-blob \
  --bundle checksums.txt.bundle \
  --certificate-identity-regexp "https://github.com/confire-ai/mono" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  checksums.txt
```

---

## Legal review checklist

Before the first public release, ensure the following are reviewed:

- [ ] `LICENSE` reviewed by legal counsel
- [ ] `TERMS.md` reviewed by legal counsel
- [ ] Third-party OSS dependency list generated and added to `NOTICE`
- [ ] SBOM generated (e.g. `syft`) and published to `get.confire.dev/sbom.json`
- [ ] Security contact (`security@confire.dev` or `https://confire.dev/security`) is live
- [ ] cosign signing integrated into CI/CD pipeline
- [ ] `latest.json` update automated in release workflow

---

## Release infrastructure — file map

Files that make up the release and distribution setup:

| File | Purpose |
|---|---|
| `LICENSE` | Proprietary license — no copy/redistribute/reverse-engineer |
| `NOTICE` | Copyright notice, OSS dependency placeholder, security contact |
| `TERMS.md` | Developer-readable terms of use |
| `cli/cmd/version.go` | Build-time vars (`buildVersion`, `buildCommit`, `buildDate`, `buildBuiltBy`); `confire version` output |
| `cli/Makefile` | `make dev`, `dist`, `release`, `snapshot`, `test`, `clean` |
| `cli/.goreleaser.yaml` | Cross-platform release builds, checksums, cosign signing |
| `scripts/install.sh` | POSIX installer — tar.gz format, mandatory SHA-256, optional cosign, license notice |
| `homebrew/confire.rb` | Live Homebrew formula (pushed to `confire-ai/homebrew-confire` on release) |
| `packaging/homebrew/Formula/confire.rb` | Formula template using `get.confire.dev` tar.gz URLs |
| `docs/release.md` | This file |

## Dev environment URLs

| Service | URL |
|---|---|
| Worker (dev) | `https://api.dev.confire.dev` |
| Platform (dev) | `https://dev.confire.dev` |
| Worker (prod) | `https://api.confire.dev` |
| Platform (prod) | `https://confire.dev` |
| Distribution | `https://get.confire.dev` (no dev/prod split) |

The `confire-dev` binary bakes in `api.dev.confire.dev` and `dev.confire.dev` at build time via ldflags. Build it with:

```sh
cd cli && make build-dev-env
```
