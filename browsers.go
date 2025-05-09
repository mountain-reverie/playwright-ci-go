package playwrightcigo

import (
	"context"
	"fmt"

	"github.com/playwright-community/playwright-go"
)

var chromium string
var chromiumCancel context.CancelFunc

// Chromium launches a Chromium browser instance in the container
// and returns a browser object that can be used to create pages,
// navigate to websites, and perform browser automation.
//
// The connection to the browser is established via WebSockets.
// You should call browser.Close() when you're done with the browser.
// The API of the returned browser object is the Playwright API.
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

// Firefox launches a Firefox browser instance in the container
// and returns a browser object that can be used to create pages,
// navigate to websites, and perform browser automation.
//
// The connection to the browser is established via WebSockets.
// You should call browser.Close() when you're done with the browser.
// The API of the returned browser object is the Playwright API.
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

// Webkit launches a WebKit browser instance in the container
// and returns a browser object that can be used to create pages,
// navigate to websites, and perform browser automation.
//
// The connection to the browser is established via WebSockets.
// You should call browser.Close() when you're done with the browser.
// The API of the returned browser object is the Playwright API.
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
