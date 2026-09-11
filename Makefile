VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

PKG := github.com/DRMCGT/Squawk/internal/version

LDFLAGS := -s -w \
	-X $(PKG).Version=$(VERSION) \
	-X $(PKG).Commit=$(COMMIT) \
	-X $(PKG).BuildDate=$(BUILD_DATE)

.PHONY: build test vet fmt check release install clean

build:
	mkdir -p dist
	go build -trimpath -ldflags "$(LDFLAGS)" -o dist/squawk ./cmd/squawk

install:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(HOME)/.local/bin/squawk ./cmd/squawk
	@echo "installed $(HOME)/.local/bin/squawk"

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

check:
	go vet ./...
	test -z "$$(gofmt -l .)"

# Builds per-platform binaries, packages each in a tarball containing a single
# `squawk` binary, and writes SHA256SUMS.txt for install.sh to verify against.
release:
	mkdir -p dist
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do \
		os=$${target%/*}; arch=$${target#*/}; \
		go build -trimpath -ldflags "$(LDFLAGS)" \
			-o "dist/squawk-$$os-$$arch" ./cmd/squawk; \
		mkdir -p "dist/pkg-$$os-$$arch"; \
		cp "dist/squawk-$$os-$$arch" "dist/pkg-$$os-$$arch/squawk"; \
		tar -czf "dist/squawk-$$os-$$arch.tar.gz" -C "dist/pkg-$$os-$$arch" squawk; \
		rm -rf "dist/pkg-$$os-$$arch" "dist/squawk-$$os-$$arch"; \
	done
	cd dist && sha256sum squawk-*.tar.gz > SHA256SUMS.txt

clean:
	rm -rf dist