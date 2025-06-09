# playwright-ci-go

[![Go Reference](https://pkg.go.dev/badge/github.com/mountain-reverie/playwright-ci-go.svg)](https://pkg.go.dev/github.com/mountain-reverie/playwright-ci-go)
[![Coverage](https://mountain-reverie.github.io/playwright-ci-go/coverage-badge.svg)](https://mountain-reverie.github.io/playwright-ci-go/coverage.html#file0)
[![Benchmark](https://mountain-reverie.github.io/playwright-ci-go/benchmark/badge.svg)](https://mountain-reverie.github.io/playwright-ci-go/benchmark/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://github.com/mountain-reverie/playwright-ci-go/actions/workflows/main.yml/badge.svg)](https://github.com/mountain-reverie/playwright-ci-go/actions)

Reproducible browser testing with [Playwright for Go](https://github.com/playwright-community/playwright-go) using containerization for consistent CI environments.

This package enables you to run Playwright browser tests in Go with predictable behavior across different CI/CD platforms by leveraging containers.

## Features

- 🔄 **Consistent testing environment** across local and CI setups
- 🌐 **Multi-browser support** for Chromium, Firefox, and WebKit
- 🔌 **Transparent proxy** for browser traffic inspection and manipulation
- 🛠 **Simple API** with minimal configuration needed
- 🧪 **CI-friendly** design with GitHub Actions workflows in mind

## Installation

```bash
go get github.com/mountain-reverie/playwright-ci-go
```

## Quick Start

```go
package main

import (
	"log"
	"time"

	playwrightcigo "github.com/mountain-reverie/playwright-ci-go"
)

func main() {
	// Install the necessary components (do this once)
	if err := playwrightcigo.Install(playwrightcigo.WithTimeout(time.Minute)); err != nil {
		log.Fatalf("Could not install playwright-ci-go: %v", err)
	}
	defer playwrightcigo.Uninstall()

	// Launch a Firefox browser
	browser, err := playwrightcigo.Firefox()
	if err != nil {
		log.Fatalf("Could not launch Firefox: %v", err)
	}
	defer browser.Close()

	// Create a page and navigate to a website
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("Could not create page: %v", err)
	}

	if _, err = page.Goto("https://example.com"); err != nil {
		log.Fatalf("Could not navigate: %v", err)
	}

	// Take a screenshot
	if _, err = page.Screenshot(); err != nil {
		log.Fatalf("Could not take screenshot: %v", err)
	}
}
```

## Configuration Options

```go
// Available options for customization
playwrightcigo.Install(
    // Set a custom timeout duration
    playwrightcigo.WithTimeout(time.Minute * 2),
    
    // Use a custom container repository and tag
    playwrightcigo.WithRepository("custom-registry/custom-image", "latest"),
    
    // Set context for cancellation
    playwrightcigo.WithContext(ctx),
    
    // Configure retry attempts
    playwrightcigo.WithRetry(10),
    
    // Adjust sleeping duration between retries
    playwrightcigo.WithSleeping(300 * time.Millisecond),
)
```

## API Reference

### Installation and Cleanup

#### Install

```go
func Install(opts ...Option) error
```

Installs the necessary components for Playwright testing in a containerized environment.

**Options:**
- `WithTimeout(timeout time.Duration)` - Sets a custom timeout for installation (default: 5 minutes)
- `WithContext(ctx context.Context)` - Provides a context for cancellation (default: background context)
- `WithRetry(count int)` - Sets the number of retry attempts (default: 15)
- `WithSleeping(duration time.Duration)` - Sets sleep duration between retries (default: 200ms)
- `WithRepository(repository, tag string)` - Uses a custom container repository and tag

**Example:**
```go
err := playwrightcigo.Install(
    playwrightcigo.WithTimeout(3 * time.Minute),
    playwrightcigo.WithRetry(20),
)
```

#### Uninstall

```go
func Uninstall() error
```

Cleans up and removes the Playwright resources.

**Example:**
```go
defer playwrightcigo.Uninstall()
```

### Browsers

#### Chromium

```go
func Chromium() (playwright.Browser, error)
```

Launches and returns a Chromium browser instance.

**Example:**
```go
browser, err := playwrightcigo.Chromium()
if err != nil {
    log.Fatalf("Could not launch Chromium: %v", err)
}
defer browser.Close()
```

#### Firefox

```go
func Firefox() (playwright.Browser, error)
```

Launches and returns a Firefox browser instance.

**Example:**
```go
browser, err := playwrightcigo.Firefox()
if err != nil {
    log.Fatalf("Could not launch Firefox: %v", err)
}
defer browser.Close()
```

#### Webkit

```go
func Webkit() (playwright.Browser, error)
```

Launches and returns a WebKit browser instance.

**Example:**
```go
browser, err := playwrightcigo.Webkit()
if err != nil {
    log.Fatalf("Could not launch WebKit: %v", err)
}
defer browser.Close()
```

### Utilities

#### Wait4Port

```go
func Wait4Port(addr string, opts ...Option) error
```

Waits for a specific network port to be available with configurable retry logic.

**Parameters:**
- `addr` - The address to check (e.g., "http://localhost:8080")
- `opts` - Options for customizing the wait behavior

**Example:**
```go
err := playwrightcigo.Wait4Port(
    "http://localhost:8080", 
    playwrightcigo.WithRetry(10),
    playwrightcigo.WithSleeping(500 * time.Millisecond),
)
```

#### Option Customization

```go
func WithContext(ctx context.Context) Option
func WithTimeout(timeout time.Duration) Option
func WithRetry(count int) Option
func WithSleeping(sleeping time.Duration) Option
func WithRepository(repository, tag string) Option
```

Functions for customizing behavior of the library operations.

**Example:**
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

playwrightcigo.Install(
    playwrightcigo.WithContext(ctx),
    playwrightcigo.WithTimeout(time.Minute * 2),
    playwrightcigo.WithRetry(10),
)
```

## CI Integration

This package includes built-in GitHub Actions workflows for continuous integration. See the `.github/workflows` directory for examples.

## Requirements

- Go 1.24 or later
- Docker or compatible container runtime

## License

[MIT](LICENSE)
