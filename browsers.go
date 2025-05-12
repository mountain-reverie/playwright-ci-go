package playwrightcigo

import (
	"context"
	"fmt"
	"sync"

	"github.com/playwright-community/playwright-go"
)

var mutexBrowser sync.Mutex

type browser struct {
	instanceOf   string
	instancePort int
	uri          string
	cancel       context.CancelFunc
	count        int
}

func (b *browser) connect() (playwright.Browser, error) {
	mutexBrowser.Lock()
	defer mutexBrowser.Unlock()

	if b.count > 0 {
		b.count++
		return connect(b.instanceOf, b.uri)
	}

	if browsers == nil {
		return nil, fmt.Errorf("container is not running")
	}

	uri, cancel, err := browsers.Exec(b.instanceOf, b.instancePort)
	if err != nil {
		return nil, fmt.Errorf("could not exec chromium: %w", err)
	}
	b.uri = uri
	b.cancel = func() {
		mutexBrowser.Lock()
		defer mutexBrowser.Unlock()

		b.count--
		if b.count == 0 {
			cancel()
		}
	}
	b.count++

	return connect(b.instanceOf, b.uri)
}

func connect(instanceOf, uri string) (playwright.Browser, error) {
	switch instanceOf {
	case "chromium":
		return pw.Chromium.Connect(uri)
	case "firefox":
		return pw.Firefox.Connect(uri)
	case "webkit":
		return pw.WebKit.Connect(uri)
	default:
		return nil, fmt.Errorf("unknown browser instance: %s", instanceOf)
	}
}

var chromium = browser{
	instanceOf:   "chromium",
	instancePort: 1024 + 3,
}

// Chromium launches a Chromium browser instance in the container
// and returns a browser object that can be used to create pages,
// navigate to websites, and perform browser automation.
//
// The connection to the browser is established via WebSockets.
// You should call browser.Close() when you're done with the browser.
// The API of the returned browser object is the Playwright API.
func Chromium() (playwright.Browser, error) {
	return chromium.connect()
}

var firefox = browser{
	instanceOf:   "firefox",
	instancePort: 1024 + 1,
}

// Firefox launches a Firefox browser instance in the container
// and returns a browser object that can be used to create pages,
// navigate to websites, and perform browser automation.
//
// The connection to the browser is established via WebSockets.
// You should call browser.Close() when you're done with the browser.
// The API of the returned browser object is the Playwright API.
func Firefox() (playwright.Browser, error) {
	return firefox.connect()
}

var webkit = browser{
	instanceOf:   "webkit",
	instancePort: 1024 + 2,
}

// Webkit launches a WebKit browser instance in the container
// and returns a browser object that can be used to create pages,
// navigate to websites, and perform browser automation.
//
// The connection to the browser is established via WebSockets.
// You should call browser.Close() when you're done with the browser.
// The API of the returned browser object is the Playwright API.
func Webkit() (playwright.Browser, error) {
	return webkit.connect()
}
