package playwrightcigo

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type config struct {
	ctx        context.Context
	timeout    time.Duration
	repository string
	tag        string
}

type container struct {
	context    context.Context
	proxy      string
	proxyPort  int
	proxyClose func()
	browsers   testcontainers.Container
	terminate  func()
}

func new(version string, opts ...Option) (*container, error) {
	c := &config{
		timeout:    5 * time.Minute,
		ctx:        context.Background(),
		repository: "mountain-reverie/playwright-ci-go",
		tag:        version,
	}
	for _, opt := range opts {
		opt.apply(c)
	}

	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)

	timeoutSecond := int(c.timeout.Seconds())

	proxy, proxyPort, close := transparentProxy()

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

func (c *container) Close() error {
	if err := c.browsers.Terminate(context.Background()); err != nil {
		return fmt.Errorf("could not terminate browser container: %w", err)
	}
	c.proxyClose()
	c.terminate()
	return nil
}

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
	if err := wait4port(fmt.Sprintf("http://%s:%d", host, p.Int())); err != nil {
		return 0, fmt.Errorf("timeout, could not connect to browser container: %w", err)
	}
	return p.Int(), nil
}
