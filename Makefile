VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

PKG := github.com/squawk-dev/squawk/internal/version

LDFLAGS := -s -w \
	-X $(PKG).Version=$(VERSION) \
	-X $(PKG).Commit=$(COMMIT) \
	-X $(PKG).BuildDate=$(BUILD_DATE)

.PHONY: build test vet fmt check release clean

build:
	mkdir -p dist
	go build -trimpath -ldflags "$(LDFLAGS)" -o dist/squawk ./cmd/squawk

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

check:
	go vet ./...
	test -z "$$(gofmt -l .)"

release:
	mkdir -p dist
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		os=$${target%/*}; arch=$${target#*/}; \
		ext=""; if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		GOOS=$$os GOARCH=$$arch go build -trimpath -ldflags "$(LDFLAGS)" \
			-o "dist/squawk-$$os-$$arch$$ext" ./cmd/squawk; \
	done

clean:
	rm -rf dist