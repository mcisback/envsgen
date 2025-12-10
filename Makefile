APP       := envsgen
PKG       := main
BUILD_DIR := builds

# Injected metadata
VERSION    := $(shell git describe --tags --always --dirty)
COMMIT     := $(shell git rev-parse HEAD)
BUILDDATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GOVERSION  := $(shell go version | sed 's/go version //')

LDFLAGS := -ldflags "\
	-X $(PKG).Version=$(VERSION) \
	-X $(PKG).Commit=$(COMMIT) \
	-X $(PKG).BuildDate=$(BUILDDATE) \
	-X '$(PKG).GoVersion=$(GOVERSION)' \
"

.PHONY: build clean run release linux darwin silicon windows

## Default build (local OS/ARCH)
build:
	@echo "⚙️  Building $(APP)…"
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP) .

## Run locally (no build artifact)
run:
	go run $(LDFLAGS) .

## Clean build artifacts
clean:
	@echo "🧹  Cleaning…"
	rm -rf $(BUILD_DIR)

## Cross-compile for Linux amd64
linux:
	@echo "🐧  Building Linux amd64 binary…"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP)-linux-amd64 .

## Cross-compile for macOS amd64 (Intel)
darwin:
	@echo "🍎  Building macOS amd64 binary…"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP)-macos-amd64 .

## Cross-compile for macOS arm64 (Apple Silicon: M-Series)
silicon:
	@echo "🍏  Building macOS arm64 (Apple Silicon) binary…"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP)-macos-arm64 .

## Cross-compile for Windows amd64
windows:
	@echo "🪟  Building Windows amd64 binary…"
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP)-windows-amd64.exe .

## Release bundle: all main OS/ARCH builds
release: clean linux darwin silicon windows
	@echo "🎁  Release binaries created in $(BUILD_DIR):"
	@ls -lh $(BUILD_DIR)
