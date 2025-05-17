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

	"github.com/docker/go-connections/nat"
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

func new(version string, opts ...Option) (*container, error) {
	splitted := strings.Split(version, ".")
	if len(splitted) != 3 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	imageVersion := fmt.Sprintf("v0.%s%02s.0", splitted[1], splitted[2])
	if imageVersion == "v0.5101.0" {
		// Workaround our CI having failed publishing this version
		imageVersion = "v0.5101.1"
	}

	found := false

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, deps := range info.Deps {
			fmt.Println("Dependency:", deps.Path, "Version:", deps.Version)
			if strings.Contains(deps.Path, "github.com/mountain-reverie/playwright-ci-go") {
				if len(deps.Version) > 0 && deps.Version[0] == 'v' {
					imageVersion = deps.Version
					found = true
					log.Println("Using version from build info:", imageVersion)
					break
				}
			}
		}
	}

	if !found {
		cmd := exec.Command("go", "list", "-json", "-m", "all")
		output, err := cmd.StdoutPipe()
		if err != nil {
			return nil, fmt.Errorf("could not get stdout pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("could not start command: %w", err)
		}
		defer func() {
			_ = cmd.Wait()
		}()

		decoder := json.NewDecoder(output)

		for {
			var mod module
			if err := decoder.Decode(&mod); err != nil {
				if err == io.EOF {
					break
				}
				return nil, fmt.Errorf("could not decode module: %w", err)
			}
			if strings.Contains(mod.Path, "github.com/mountain-reverie/playwright-ci-go") {
				if mod.Main {
					found, imageVersion = getPlaywrightCIGoGitVersion(imageVersion)
					break
				} else if len(mod.Version) > 0 && mod.Version[0] == 'v' {
					found = true
					imageVersion = mod.Version
					log.Println("Using version from go list:", imageVersion)
					break
				}
			}
		}

		if !found {
			log.Println("No build, module or git info found. Keeping version as:", imageVersion)
		}
	}

	c := &config{
		timeout:    5 * time.Minute,
		sleeping:   200 * time.Millisecond,
		retry:      15,
		ctx:        context.Background(),
		repository: "ghcr.io/mountain-reverie/playwright-ci-go",
		tag:        imageVersion,
	}
	for _, opt := range opts {
		opt.apply(c)
	}

	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)

	timeoutSecond := int(c.timeout.Seconds())

	proxy, proxyPort, close := transparentProxy(c.retry, c.sleeping)

	log.Println("Starting browser container", fmt.Sprintf("%s:%s", c.repository, c.tag))
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
	p, err := container.MappedPort(ctx, nat.Port(fmt.Sprintf("%d/tcp", port)))
	if err != nil {
		return 0, fmt.Errorf("could not get browser port: %w", err)
	}
	if err := Wait4Port(fmt.Sprintf("http://%s:%d", host, p.Int())); err != nil {
		return 0, fmt.Errorf("timeout, could not connect to browser container: %w", err)
	}
	return p.Int(), nil
}

func getPlaywrightCIGoGitVersion(imageVersion string) (bool, string) {
	cmd := exec.Command("git", "describe", "--tags")
	output, err := cmd.Output()
	if err != nil {
		return false, imageVersion
	}
	imageVersion = strings.TrimSpace(string(output))
	log.Println("Using version from git:", imageVersion)
	return true, imageVersion
}
