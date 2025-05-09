# playwright-ci-go

[![Go Reference](https://pkg.go.dev/badge/github.com/mountain-reverie/playwright-ci-go.svg)](https://pkg.go.dev/github.com/mountain-reverie/playwright-ci-go)
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

## Supported Browsers

Access any of these browsers with a simple API:

```go
// Get a Chromium browser instance
chromium, err := playwrightcigo.Chromium()

// Get a Firefox browser instance
firefox, err := playwrightcigo.Firefox()

// Get a WebKit browser instance
webkit, err := playwrightcigo.Webkit()
```

## CI Integration

This package includes built-in GitHub Actions workflows for continuous integration. See the `.github/workflows` directory for examples.

## Requirements

- Go 1.24 or later
- Docker or compatible container runtime

## License

[MIT](LICENSE)
