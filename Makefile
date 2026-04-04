# YTed Makefile

.PHONY: all build build-versioned build-installer-linux build-installer-windows dev test lint fmt clean install install-system uninstall uninstall-system deps security generate run help

# Version (override with VERSION=x.y.z)
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# LDFLAGS for version injection (just the flags, not the -ldflags prefix)
LDFLAGS := -X yted/internal/version.Version=$(VERSION) -X yted/internal/version.Commit=$(COMMIT) -X yted/internal/version.BuildDate=$(BUILD_DATE)

# Default target
all: fmt lint build

## Build the application for production
build:
	@echo "Building YTed..."
	wails build -tags webkit2_41

## Build with version info injected
build-versioned:
	@echo "Building YTed version $(VERSION) (commit: $(COMMIT))..."
	@echo "Updating version in frontend/package.json..."
	@cd frontend && npm version $(VERSION) --no-git-tag-version --allow-same-version 2>/dev/null || true
	wails build -tags webkit2_41 -ldflags "$(LDFLAGS)"

## Build Linux .deb installer
build-installer-linux:
	@echo "Building Linux .deb package..."
	@build/scripts/build-deb.sh $(VERSION)

## Build Windows NSIS installer (requires NSIS)
build-installer-windows:
	@echo "Building Windows installer..."
	wails build --platform windows/amd64 -ldflags "-s -w" -nsis -o YTed.exe

## Build with dev mode
dev:
	@echo "Running in development mode..."
	wails dev -tags webkit2_41

## Run tests
test:
	@echo "Running Go tests..."
	go test -v ./...
	@echo "Running frontend tests..."
	cd frontend && npm test

## Run linter
lint:
	@echo "Running Go linter..."
	golangci-lint run
	@echo "Running frontend linter..."
	cd frontend && npm run lint

## Format code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	goimports -w -local yted .
	gci write --skip-generated -s standard -s default -s "prefix(yted)" .
	@echo "Formatting frontend code..."
	cd frontend && npm run format

## Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf build/bin
	rm -rf build/linux/deb-pkg
	rm -rf frontend/dist
	rm -rf frontend/node_modules

## Install dependencies
deps:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Installing dev tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/daixiang0/gci@latest

## Check for security vulnerabilities
security:
	@echo "Running security checks..."
	gosec ./...
	cd frontend && npm audit

## Generate Wails bindings
generate:
	wails generate module

## Run the application
run:
	./build/bin/yted

## Install for current user (~/.local/)
install:
	@echo "Installing YTed for current user..."
	@build/linux/install.sh

## Install system-wide (/usr/local/)
install-system:
	@echo "Installing YTed system-wide..."
	@sudo build/linux/install.sh --system

## Uninstall for current user
uninstall:
	@echo "Uninstalling YTed..."
	@rm -f ~/.local/bin/yted
	@rm -f ~/.local/share/icons/yted.png
	@rm -f ~/.local/share/applications/yted.desktop
	@update-desktop-database ~/.local/share/applications 2>/dev/null || true
	@echo "YTed uninstalled from user directories"

## Uninstall system-wide
uninstall-system:
	@echo "Uninstalling YTed system-wide..."
	@sudo rm -f /usr/local/bin/yted
	@sudo rm -f /usr/share/pixmaps/yted.png
	@sudo rm -f /usr/share/applications/yted.desktop
	@sudo update-desktop-database /usr/share/applications 2>/dev/null || true
	@echo "YTed uninstalled from system directories"

## Show help
help:
	@echo "Available targets:"
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
		info = match(lastLine, /^## (.*)/); \
		if (info) { \
			target = substr($$1, 1, length($$1)-1); \
			printf "  %-15s - %s\n", target, substr(lastLine, 4); \
		} \
	} { lastLine = $$0 }' $(MAKEFILE_LIST)
