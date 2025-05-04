package playwrightcigo

import (
	"context"
	"fmt"

	"github.com/playwright-community/playwright-go"
)

var chromium string
var chromiumCancel context.CancelFunc

func Chromium() (playwright.Browser, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if chromium != "" {
		return pw.Chromium.Connect(chromium)
	}

	if browsers == nil {
		return nil, fmt.Errorf("container is not running")
	}

	uri, cancel, err := browsers.Exec("chromium", 1024+3)
	if err != nil {
		return nil, fmt.Errorf("could not exec chromium: %w", err)
	}
	chromium = uri
	chromiumCancel = cancel

	return pw.Chromium.Connect(chromium)
}

var firefox string
var firefoxCancel context.CancelFunc

func Firefox() (playwright.Browser, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if firefox != "" {
		return pw.Firefox.Connect(firefox)
	}

	if browsers == nil {
		return nil, fmt.Errorf("container is not running")
	}

	uri, cancel, err := browsers.Exec("firefox", 1024+1)
	if err != nil {
		return nil, fmt.Errorf("could not exec firefox: %w", err)
	}
	firefox = uri
	firefoxCancel = cancel

	return pw.Firefox.Connect(firefox)
}

var webkit string
var webkitCancel context.CancelFunc

func Webkit() (playwright.Browser, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if webkit != "" {
		return pw.WebKit.Connect(webkit)
	}

	if browsers == nil {
		return nil, fmt.Errorf("container is not running")
	}

	uri, cancel, err := browsers.Exec("webkit", 1024+2)
	if err != nil {
		return nil, fmt.Errorf("could not exec webkit: %w", err)
	}
	webkit = uri
	webkitCancel = cancel

	return pw.WebKit.Connect(webkit)
}
