APP       := envsgen
PKG       := main
BUILD_DIR := builds

# FIX: $(PKG).InstallDir=$(INSTALL_DIR) is not working as expected
INSTALL_DIR ?= /usr/local/bin

# Injected metadata
VERSION    := $(shell git describe --tags --always --dirty)
COMMIT     := $(shell git rev-parse HEAD)
BUILDDATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GOVERSION  := $(shell go version | sed 's/go version //')

INSTALLED_PATH := $(shell envsgen 2>/dev/null | grep InstallDir | cut -d':' -f2 | tr -d ' ')


LDFLAGS := -ldflags "\
	-X $(PKG).Version=$(VERSION) \
	-X $(PKG).Commit=$(COMMIT) \
	-X $(PKG).BuildDate=$(BUILDDATE) \
	-X '$(PKG).GoVersion=$(GOVERSION)' \
	-X '$(PKG).InstallDir=$(INSTALL_DIR)' \
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

## Install native binary
install: build
	@echo "📦 Installing $(APP) into $(INSTALL_DIR)…"
	@install -m 755 $(BUILD_DIR)/$(APP) $(INSTALL_DIR)/$(APP)
	@echo "✔️ Installed $(APP) to $(INSTALL_DIR)"

## Uninstall
uninstall:
	@if ! command -v envsgen >/dev/null 2>&1; then \
		echo "⚠️  'envsgen' not found in PATH, nothing to uninstall."; \
		exit 0; \
	fi \

	@echo "ℹ️ Checking $(INSTALLED_PATH)"
	
	@if [ -z "$(INSTALLED_PATH)" ]; then \
		echo "❌ Could not detect install path"; \
		exit 1; \
	fi \
	
	@echo "🗑️  Removing $(INSTALLED_PATH)/$(APP)"
	@rm -f $(INSTALLED_PATH)/$(APP)
	@echo "✔️ Removed $(APP)"

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
