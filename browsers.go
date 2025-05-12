package playwrightcigo

import (
	"context"
	"fmt"
	"sync"

	"github.com/playwright-community/playwright-go"
)

var mutexBrowser sync.Mutex

var chromium string
var chromiumCancel context.CancelFunc
var chromiumCount int

// Chromium launches a Chromium browser instance in the container
// and returns a browser object that can be used to create pages,
// navigate to websites, and perform browser automation.
//
// The connection to the browser is established via WebSockets.
// You should call browser.Close() when you're done with the browser.
// The API of the returned browser object is the Playwright API.
func Chromium() (playwright.Browser, error) {
	mutexBrowser.Lock()
	defer mutexBrowser.Unlock()

	if chromiumCount > 0 {
		chromiumCount++
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
	chromiumCancel = func() {
		mutexBrowser.Lock()
		defer mutexBrowser.Unlock()

		chromiumCount--
		if chromiumCount == 0 {
			chromium = ""
			cancel()
		}
	}
	chromiumCount++

	return pw.Chromium.Connect(chromium)
}

var firefox string
var firefoxCancel context.CancelFunc
var firefoxCount int

// Firefox launches a Firefox browser instance in the container
// and returns a browser object that can be used to create pages,
// navigate to websites, and perform browser automation.
//
// The connection to the browser is established via WebSockets.
// You should call browser.Close() when you're done with the browser.
// The API of the returned browser object is the Playwright API.
func Firefox() (playwright.Browser, error) {
	mutexBrowser.Lock()
	defer mutexBrowser.Unlock()

	if firefoxCount > 0 {
		firefoxCount++
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
	firefoxCancel = func() {
		mutexBrowser.Lock()
		defer mutexBrowser.Unlock()

		firefoxCount--
		if firefoxCount == 0 {
			firefox = ""
			cancel()
		}
	}
	firefoxCount++

	return pw.Firefox.Connect(firefox)
}

var webkit string
var webkitCancel context.CancelFunc
var webkitCount int

// Webkit launches a WebKit browser instance in the container
// and returns a browser object that can be used to create pages,
// navigate to websites, and perform browser automation.
//
// The connection to the browser is established via WebSockets.
// You should call browser.Close() when you're done with the browser.
// The API of the returned browser object is the Playwright API.
func Webkit() (playwright.Browser, error) {
	mutexBrowser.Lock()
	defer mutexBrowser.Unlock()

	if webkitCount > 0 {
		webkitCount++
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
	webkitCancel = func() {
		mutexBrowser.Lock()
		defer mutexBrowser.Unlock()

		webkitCount--
		if webkitCount == 0 {
			webkit = ""
			cancel()
		}
	}
	webkitCount++

	return pw.WebKit.Connect(webkit)
}
