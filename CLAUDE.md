# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Testing
- `go test -v .` - Run all tests with verbose output
- `go test -run=TestName .` - Run specific test
- `go test -covermode=atomic -coverprofile=coverage.out .` - Run tests with coverage
- `go tool cover -html=coverage.out -o coverage.html` - Generate HTML coverage report
- `gotestsum --jsonfile tests.json --format standard-verbose -- -covermode=atomic -coverprofile=coverage.out .` - Run tests with gotestsum (requires installation)

### Benchmarking
- `go test -run='^$' -bench=. -count=7 -benchmem .` - Run benchmarks
- `benchstat benchmark.txt` - Analyze benchmark results (requires benchstat installation)

### Linting
- `golangci-lint run` - Run Go linter (requires golangci-lint installation)

### Building
- `go build .` - Build the package
- `go mod tidy` - Clean up module dependencies

### Examples
- `cd examples && go test .` - Run example tests

## Architecture Overview

This is a Go library that provides reproducible browser testing using Playwright and containerization. The architecture consists of several key components:

### Core Components

- **Container Management** (`container.go`): Manages Docker containers for browser instances using testcontainers-go. Handles container lifecycle, version detection, and WebSocket connections.

- **Browser Interfaces** (`browsers.go`): Provides high-level APIs for Chromium, Firefox, and WebKit browsers. Uses mutex-based connection pooling to manage browser instances efficiently.

- **Installation** (`install.go`): Handles installation and setup of the containerized environment with configurable options for timeouts, retries, and custom repositories.

- **Proxy System** (`proxy.go`): Implements a transparent HTTP proxy using goproxy for traffic inspection and manipulation during browser testing.

- **Configuration** (`options.go`): Provides functional options pattern for customizing behavior (timeouts, repositories, contexts, etc.).

### Key Design Patterns

- **Functional Options**: All configuration uses the options pattern (`WithTimeout()`, `WithRepository()`, etc.)
- **Resource Management**: Containers and browsers are properly closed with defer patterns
- **Version Detection**: Automatically detects Playwright versions from go.mod, build info, or git tags
- **Connection Pooling**: Browser connections are reused when possible to improve performance

### Container Integration

The library uses Docker containers with:
- Pre-built browser images hosted on GitHub Container Registry
- WebSocket connections on ports 1025-1027 for browser communication  
- Transparent proxy integration for traffic monitoring
- Multi-architecture support (amd64/arm64)

### Dependencies

- `github.com/playwright-community/playwright-go` - Core Playwright Go bindings
- `github.com/testcontainers/testcontainers-go` - Container management
- `github.com/docker/go-connections` - Docker networking utilities
- `github.com/elazarl/goproxy` - HTTP proxy implementation

## Environment Variables

- `PLAYWRIGHTCI_REPOSITORY` - Override container repository (used in CI)
- `PLAYWRIGHTCI_TAG` - Override container tag (used in CI)
- Various testcontainers environment variables for Docker configuration

## CI/CD

The project uses GitHub Actions with:
- Multi-platform testing (Ubuntu amd64/arm64)
- Automated Docker image building and publishing
- Security scanning with Anchore
- CodeQL analysis
- Benchmark tracking and coverage reporting
- Automatic versioning based on Playwright dependency versions