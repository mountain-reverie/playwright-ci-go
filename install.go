// Package playwrightcigo provides a containerized solution for running
// Playwright browser tests in Go with consistent behavior across environments.
package playwrightcigo

import (
	"fmt"
	"log"
	"sync"

	"github.com/docker/docker/errdefs"
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

	if count > 0 {
		count++
		return nil
	}

	c := config{}
	for _, opt := range opts {
		opt.apply(&c)
	}

	driver, err := playwright.NewDriver(&playwright.RunOptions{SkipInstallBrowsers: true, Verbose: c.verbose})
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
	mutex.Lock()
	defer mutex.Unlock()
	count--
	if count > 0 {
		return nil
	}

	for chromium.count > 0 {
		chromium.cancel()
	}
	for firefox.count > 0 {
		firefox.cancel()
	}
	for webkit.count > 0 {
		webkit.cancel()
	}

	if err := browsers.Close(); err != nil {
		// Ignore "not found" errors since the container may have already been terminated due to timeout or other cleanup.
		if !errdefs.IsNotFound(err) {
			return fmt.Errorf("could not close container: %w", err)
		}
		log.Println("container already closed or not found, ignoring error:", err)
	}

	if err := pw.Stop(); err != nil {
		return fmt.Errorf("could not stop playwright: %w", err)
	}
	return nil
}
