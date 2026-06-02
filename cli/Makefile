## confire Makefile
## Author: Efe <efe@efebehar.dev>
##
## Usage:
##   make dev           build dev binary (symbols kept, fast)
##   make release       build hardened binary for current platform (garble + stripped)
##   make dist          cross-compile all release targets into dist/
##   make test          run unit tests
##   make e2e           run end-to-end tests (requires built binary)
##   make all           dev + test

BINARY     := confire
PKG        := github.com/confire-dev/confire
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE       ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
AUTHOR     := Efe <efe@efebehar.dev>

# ldflags applied to ALL builds
BASE_FLAGS := -X '$(PKG)/cmd.buildVersion=$(VERSION)' \
              -X '$(PKG)/cmd.buildCommit=$(COMMIT)'   \
              -X '$(PKG)/cmd.buildDate=$(DATE)'        \
              -X '$(PKG)/cmd.buildAuthor=$(AUTHOR)'

# Release flags: strip symbol table, DWARF debug info, source paths
REL_FLAGS  := $(BASE_FLAGS) -s -w

GARBLE     := $(shell go env GOPATH)/bin/garble
GARBLE_FLAGS := -literals -seed=random -tiny

DIST_DIR   := dist

# Cross-compile targets (CGO_ENABLED=0 — modernc/sqlite is pure Go)
TARGETS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64  \
	linux/arm64

.PHONY: all dev release dist test e2e clean install checksums

all: dev test

## dev: fast build with symbols (for local development)
dev:
	go build -trimpath -ldflags "$(BASE_FLAGS)" -o $(BINARY) .

## dev-install: build dev binary directly to DEST (avoids macOS provenance xattr on copied files)
dev-install:
	go build -trimpath -ldflags "$(BASE_FLAGS)" -o $(DEST) .

## release: hardened binary for the current platform (garble + stripped)
release:
	CGO_ENABLED=1 $(GARBLE) $(GARBLE_FLAGS) build \
		-trimpath \
		-ldflags "$(REL_FLAGS)" \
		-o $(BINARY) .
	@echo "Release binary: $$(ls -lh $(BINARY) | awk '{print $$5}')"
	@echo "Symbols:        $$(nm $(BINARY) 2>/dev/null | wc -l | tr -d ' ') (should be 0)"

## dist: cross-compile all targets into dist/ with plain go build + stripped flags
##       (garble does not support cross-compilation; use release: for the local platform)
dist: clean-dist
	@mkdir -p $(DIST_DIR)
	@$(foreach TARGET,$(TARGETS), \
		$(eval OS   := $(word 1,$(subst /, ,$(TARGET)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(TARGET)))) \
		echo "  building $(OS)/$(ARCH)..." && \
		CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) \
			go build -trimpath -ldflags "$(REL_FLAGS)" \
			-o "$(DIST_DIR)/$(BINARY)_$(OS)_$(ARCH)" . ; \
	)
	@$(MAKE) --no-print-directory checksums

## checksums: generate SHA-256 checksums for all dist/ binaries
checksums:
	@cd $(DIST_DIR) && sha256sum $(BINARY)_* > $(BINARY)_checksums.txt
	@echo "Checksums:"
	@cat $(DIST_DIR)/$(BINARY)_checksums.txt

## test: run unit + integration tests (no subprocess, uses httptest)
test:
	go test ./... -count=1 -timeout 30s

## e2e: run full end-to-end tests (builds binary first)
e2e: dev
	go test ./e2e/... -v -count=1 -timeout 60s -binary=$(shell pwd)/$(BINARY)

## install: install to GOPATH/bin
install:
	go install -trimpath -ldflags "$(BASE_FLAGS)" .

## clean
clean:
	rm -f $(BINARY)
	go clean -testcache

clean-dist:
	rm -rf $(DIST_DIR)
