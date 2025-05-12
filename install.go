// Package playwrightcigo provides a containerized solution for running
// Playwright browser tests in Go with consistent behavior across environments.
package playwrightcigo

import (
	"fmt"
	"sync"

	"github.com/playwright-community/playwright-go"
)

var pw *playwright.Playwright
var browsers *container
var count = 0

var mutex sync.Mutex

// Install sets up the containerized Playwright environment.
// It installs the Playwright driver and creates a container to run the browsers.
// This function should be called once before using any browser.
// Options can be provided to customize the installation behavior.
// Multiple calls to Install are supported, but only the first one will
// perform the actual installation.
// Generally you want to call this in your TestMain function.
func Install(opts ...Option) error {
	mutex.Lock()
	defer mutex.Unlock()

	if pw != nil {
		count++
		return nil
	}

	if err := playwright.Install(); err != nil {
		return fmt.Errorf("error while installing playwright: %w", err)
	}

	driver, err := playwright.NewDriver(&playwright.RunOptions{SkipInstallBrowsers: true})
	if err != nil {
		return fmt.Errorf("error while setting up driver: %w", err)
	}

	if err := driver.Install(); err != nil {
		return fmt.Errorf("error while installing driver: %w", err)
	}

	pw, err = playwright.Run()
	if err != nil {
		return fmt.Errorf("error while starting to run playwright: %w", err)
	}

	browsers, err = new(driver.Version, opts...)
	if err != nil {
		return err
	}

	count++
	return nil
}

// Uninstall cleans up resources created by Install.
// This function should be called when you're finished with all browser testing.
// There should be as many calls to Uninstall as there were calls to Install.
// Generally you want to call this in your TestMain function.
func Uninstall() error {
	if pw == nil {
		return nil
	}

	mutex.Lock()
	defer mutex.Unlock()
	count--
	if count > 0 {
		return nil
	}

	if chromium != "" {
		chromiumCancel()
		chromium = ""
	}
	if firefox != "" {
		firefoxCancel()
		firefox = ""
	}
	if webkit != "" {
		webkitCancel()
		webkit = ""
	}

	if err := browsers.Close(); err != nil {
		return fmt.Errorf("could not close container: %w", err)
	}

	if err := pw.Stop(); err != nil {
		return fmt.Errorf("could not stop playwright: %w", err)
	}
	return nil
}
