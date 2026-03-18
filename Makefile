# YTed Makefile

.PHONY: all build build-versioned dev test lint fmt clean help

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
	wails build -tags webkit2_41 -ldflags "$(LDFLAGS)"

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
