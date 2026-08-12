// Command playwrightversion prints the Playwright CLI version that the
// playwright-go release pinned in go.mod drives.
//
// CI used to obtain this by running the upstream `playwright --version`
// command, but that downloads the whole driver archive first, which makes a
// pure metadata lookup depend on a CDN being reachable. NewDriver only reads
// the constant compiled into playwright-go, so this stays offline.
package main

import (
	"fmt"
	"os"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	driver, err := playwright.NewDriver(&playwright.RunOptions{SkipInstallBrowsers: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not create playwright driver:", err)
		os.Exit(1)
	}

	if driver.Version == "" {
		fmt.Fprintln(os.Stderr, "playwright driver reported an empty version")
		os.Exit(1)
	}

	fmt.Println(driver.Version)
}
