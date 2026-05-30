## confire Makefile
## Author: Efe <efe@efebehar.dev>
##
## Usage:
##   make dev      build dev binary (symbols kept, fast)
##   make release  build hardened release binary (garble + stripped)
##   make test     run unit tests
##   make e2e      run end-to-end tests (requires built binary)
##   make all      dev + test

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

.PHONY: all dev release test e2e clean install

all: dev test

## dev: fast build with symbols (for local development)
dev:
	go build -trimpath -ldflags "$(BASE_FLAGS)" -o $(BINARY) .

## release: hardened binary — garble obfuscation + stripped symbols
## Use this for all distributed binaries.
release:
	CGO_ENABLED=1 $(GARBLE) $(GARBLE_FLAGS) build \
		-trimpath \
		-ldflags "$(REL_FLAGS)" \
		-o $(BINARY) .
	@echo "Release binary: $$(ls -lh $(BINARY) | awk '{print $$5}')"
	@echo "Symbols:        $$(nm $(BINARY) 2>/dev/null | wc -l | tr -d ' ') (should be 0 for release)"

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
