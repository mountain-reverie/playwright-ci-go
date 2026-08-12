package playwrightcigo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type config struct {
	ctx        context.Context
	timeout    time.Duration
	sleeping   time.Duration
	repository string
	tag        string
	retry      int
	verbose    bool
}

type container struct {
	context    context.Context
	proxy      string
	proxyPort  int
	proxyClose func()
	browsers   testcontainers.Container
	terminate  func()
}

type module struct {
	Path    string
	Version string
	Main    bool
}

func new(opts ...Option) (*container, error) {
	c := &config{
		timeout:    5 * time.Minute,
		sleeping:   200 * time.Millisecond,
		retry:      15,
		ctx:        context.Background(),
		repository: "ghcr.io/mountain-reverie/playwright-ci-go",
		tag:        "",
		verbose:    false,
	}
	for _, opt := range opts {
		opt.apply(c)
	}

	if c.tag == "" {
		tag, err := noTagVersion(c.verbose)
		if err != nil {
			return nil, err
		}
		c.tag = tag
	}

	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)

	timeoutSecond := int(c.timeout.Seconds())

	proxy, proxyPort, close := transparentProxy(c.retry, c.sleeping, c.verbose)

	if c.verbose {
		log.Println("Starting browser container", fmt.Sprintf("%s:%s", c.repository, c.tag))
	}
	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:           fmt.Sprintf("%s:%s", c.repository, c.tag),
			HostAccessPorts: []int{int(proxyPort)},
			WorkingDir:      "/src",
			ExposedPorts:    []string{"1025/tcp", "1026/tcp", "1027/tcp"},
			Cmd:             []string{fmt.Sprintf("sleep %v", timeoutSecond+10)},
			WaitingFor:      wait.ForExec([]string{"echo", "ready"}),
		},
		Started: true,
	}

	browsers, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not start browser container: %w", err)
	}

	return &container{
		context:    ctx,
		proxy:      proxy,
		proxyPort:  proxyPort,
		proxyClose: close,
		browsers:   browsers,
		terminate:  cancel,
	}, nil
}

// Close terminates the container and cleans up associated resources.
func (c *container) Close() error {
	if err := c.browsers.Terminate(context.Background()); err != nil {
		return fmt.Errorf("could not terminate browser container: %w", err)
	}
	c.proxyClose()
	c.terminate()
	return nil
}

// Exec executes a browser command in the container and returns a WebSocket connection URL.
// The browser parameter should be one of: "chromium", "firefox", or "webkit".
// It also returns a cancel function to terminate the browser session.
func (c *container) Exec(browser string, containerPort int) (string, context.CancelFunc, error) {
	execCtx, execCancel := context.WithCancel(c.context)
	go func() {
		code, stuff, err := c.browsers.Exec(execCtx, []string{"node", browser + ".js", c.proxy, strconv.Itoa(c.proxyPort)})

		// Check that the context is not expired
		select {
		case <-execCtx.Done():
			return
		default:
		}

		if err != nil {
			log.Fatalf("Could not exec in browser container: %s", err)
		}
		if code != 0 {
			s, err := io.ReadAll(stuff)
			if err != nil {
				fmt.Println("Could not read stdout/stderr from browser container:", err)
			} else {
				fmt.Println("Browser container output:", string(s))
			}
			log.Fatalf("Exec failed in browser container: %d", code)
		}
	}()

	host, err := c.browsers.Host(c.context)
	if err != nil {
		execCancel()
		return "", nil, fmt.Errorf("could not get browser host: %w", err)
	}

	p, err := port(c.context, c.browsers, host, containerPort)
	if err != nil {
		execCancel()
		return "", nil, fmt.Errorf("could not get %s port: %w", browser, err)
	}

	return fmt.Sprintf("ws://%s:%d/"+browser, host, p), execCancel, nil
}

func port(ctx context.Context, container testcontainers.Container, host string, port int) (int, error) {
	p, err := container.MappedPort(ctx, fmt.Sprintf("%d/tcp", port))
	if err != nil {
		return 0, fmt.Errorf("could not get browser port: %w", err)
	}
	if err := Wait4Port(fmt.Sprintf("http://%s:%d", host, p.Num())); err != nil {
		return 0, fmt.Errorf("timeout, could not connect to browser container: %w", err)
	}
	return int(p.Num()), nil
}

// noTagVersion resolves which playwright-ci-go image to pull when the caller
// did not pin one with WithRepository.
//
// Both strategies read versions the Go toolchain already knows, so neither
// guesses. Build info covers consumers of this library; go list covers
// development inside this repository, where the test binary carries no
// dependency info.
func noTagVersion(verbose bool) (string, error) {
	if imageVersion, found := getPlaywrightCIGoFromBuildInfo(verbose); found {
		return imageVersion, nil
	}

	if imageVersion, found := getPlaywrightCIGoFromGoList(verbose); found {
		return imageVersion, nil
	}

	return "", fmt.Errorf("could not determine which %s image to use: no version found in build info or go list; pass one explicitly with WithRepository", playwrightCIGoModule)
}

func filterVersion(version string) string {
	parts := strings.Split(version, "-")
	if len(parts) > 0 {
		return parts[0]
	}

	return version
}

func getPlaywrightCIGoGitVersion(verbose bool) (string, bool) {
	cmd := exec.Command("git", "describe", "--tags")
	output, err := cmd.Output()
	if err != nil {
		if verbose {
			log.Printf("could not get git version: %v\n", err)
		}
		return "", false
	}
	imageVersion := filterVersion(strings.TrimSpace(string(output)))
	if verbose {
		log.Println("Using version from git:", imageVersion)
	}
	return imageVersion, true
}

func getPlaywrightCIGoFromBuildInfo(verbose bool) (string, bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		if verbose {
			log.Println("no build info available")
		}
		return "", false
	}

	return versionFromBuildInfo(info, verbose)
}

func getPlaywrightCIGoFromGoList(verbose bool) (string, bool) {
	cmd := exec.Command("go", "list", "-json", "-m", "all")
	output, err := cmd.StdoutPipe()
	if err != nil {
		if verbose {
			log.Printf("could not get stdout pipe: %v\n", err)
		}
		return "", false
	}

	if err := cmd.Start(); err != nil {
		if verbose {
			log.Printf("could not start command: %v\n", err)
		}
		return "", false
	}
	defer func() {
		_ = cmd.Wait()
	}()

	return parseGoListJSONStream(output, verbose)
}

func parseGoListJSONStream(output io.Reader, verbose bool) (string, bool) {
	decoder := json.NewDecoder(output)

	defer func() {
		// Consume the rest of the stream
		if _, err := io.Copy(io.Discard, output); err != nil && verbose {
			log.Printf("could not discard remaining output: %v\n", err)
		}
	}()

	modules := map[string]module{}

	for {
		var mod module
		if err := decoder.Decode(&mod); err != nil {
			if err == io.EOF {
				break
			}
			if verbose {
				log.Printf("could not decode module: %v\n", err)
			}
			return "", false
		}

		if !strings.Contains(mod.Path, "playwright") {
			continue
		}

		if verbose {
			if mod.Main {
				log.Printf("Found main module: %s\n", mod.Path)
			} else {
				log.Printf("Found module: %s Version: %s\n", mod.Path, mod.Version)
			}
		}

		modules[mod.Path] = mod
	}

	// The release line implied by playwright-go, used whenever we cannot read
	// a published playwright-ci-go version directly.
	line := ""
	mod, exists := modules[playwrightGoModule]
	if !exists {
		// playwright-go moved back to the mxschmitt org in v0.6100.0; a
		// consumer still pinning the archived playwright-community path
		// should keep resolving.
		mod, exists = modules[playwrightGoArchivedModule]
	}
	if exists {
		if verbose {
			log.Println("Found playwright-go module:", mod.Path, "Version:", mod.Version)
		}
		if tag, ok := playwrightGoTagLine(mod.Version); ok {
			line = tag
		}
	}

	if mod, exists := modules[playwrightCIGoModule]; exists {
		if verbose {
			log.Println("Found playwright-ci-go module:", mod.Path, "Version:", mod.Version)
		}

		// Developing in this repository: there is no published version of
		// ourselves to use, so fall back to the line, then to the git tag.
		if mod.Main {
			if line != "" {
				return line, true
			}
			return getPlaywrightCIGoGitVersion(verbose)
		}

		if len(mod.Version) > 0 && mod.Version[0] == 'v' {
			if verbose {
				log.Println("Using version from go list:", mod.Version)
			}
			return mod.Version, true
		}

		if verbose {
			log.Println("No version found in go list for playwright-ci-go module")
		}
	}

	if line != "" {
		if verbose {
			log.Println("Using release line from playwright-go module:", line)
		}
		return line, true
	}

	if verbose {
		log.Println("No build, module or git info found")
	}
	return "", false
}
