package playwrightcigo

import (
	"fmt"
	"log"
	"sync"

	"github.com/playwright-community/playwright-go"
)

var pw *playwright.Playwright
var browsers *container
var count = 0

var mutex sync.Mutex

func Install(opts ...Option) error {
	mutex.Lock()
	defer mutex.Unlock()

	if pw != nil {
		count++
		return nil
	}

	if err := playwright.Install(); err != nil {
		log.Fatalf("could not install Playwright: %v", err)
	}

	driver, err := playwright.NewDriver(&playwright.RunOptions{SkipInstallBrowsers: true})
	if err != nil {
		log.Fatalf("Could not create playwright driver: %s", err)
	}

	if err := driver.Install(); err != nil {
		log.Fatalf("Could not install playwright: %s", err)
	}

	pw, err = playwright.Run()
	if err != nil {
		log.Fatalf("Could not run playwright: %s", err)
	}

	browsers, err = new(driver.Version, opts...)
	if err != nil {
		log.Fatalf("Could not create container: %s", err)
	}

	count++
	return nil
}

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
