package playwrightcigo

import (
	"log"
	"runtime/debug"
	"strings"
)

const (
	playwrightCIGoModule = "github.com/mountain-reverie/playwright-ci-go"
	playwrightGoModule   = "github.com/mxschmitt/playwright-go"

	// playwrightGoArchivedModule is the path playwright-go used while it
	// lived in the archived playwright-community org, up to v0.6000.0.
	playwrightGoArchivedModule = "github.com/playwright-community/playwright-go"
)

// playwrightGoTagLine turns a playwright-go module version into the first
// image tag of the matching playwright-ci-go release line.
//
// Our images are tagged v0.<line>.<n>, where <line> is taken verbatim from the
// playwright-go version we build against (see the release workflow) and <n>
// counts builds. Deriving the line from playwright-go's own version is exact.
// Deriving it from the Playwright CLI version is not: playwright-go encodes
// the CLI version in its tag by convention only, and v0.6100.0 ships CLI
// 1.61.1, which that convention would have written as v0.6101.0.
func playwrightGoTagLine(version string) (string, bool) {
	if !strings.HasPrefix(version, "v") {
		return "", false
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "", false
	}

	return parts[0] + "." + parts[1] + ".0", true
}

// versionFromBuildInfo resolves the image tag from the versions the Go
// toolchain recorded in the binary, needing neither the network nor the go
// command at run time.
func versionFromBuildInfo(info *debug.BuildInfo, verbose bool) (string, bool) {
	if info == nil {
		return "", false
	}

	playwrightGo := ""

	for _, dep := range info.Deps {
		switch dep.Path {
		case playwrightCIGoModule:
			// Our own pinned version *is* the image tag, so it beats
			// anything we could derive.
			if len(dep.Version) > 0 && dep.Version[0] == 'v' {
				if verbose {
					log.Println("Using playwright-ci-go version from build info:", dep.Version)
				}
				return dep.Version, true
			}
		case playwrightGoModule, playwrightGoArchivedModule:
			playwrightGo = dep.Version
		}
	}

	if tag, ok := playwrightGoTagLine(playwrightGo); ok {
		if verbose {
			log.Println("Using release line from playwright-go version", playwrightGo, "->", tag)
		}
		return tag, true
	}

	return "", false
}
