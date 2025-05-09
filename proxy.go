package playwrightcigo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/testcontainers/testcontainers-go"
)

func transparentProxy(retry int, sleeping time.Duration) (string, int, func()) {
	// Listen for incoming connections
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal("Error listening:", err)
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	srv := &http.Server{
		Handler:           proxy,
		ReadHeaderTimeout: time.Second * 5, // Set a reasonable ReadHeaderTimeout value
	}

	go func() {
		err := srv.Serve(l)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error serving proxy: %v", err)
		}
	}()

	_, portStr, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		log.Fatalf("Failed to parse address %s: %v", l.Addr().String(), err)
	}
	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse port number from address %s: %v", l.Addr().String(), err)
	}
	// Ensure the port number is within the valid range for a 16-bit unsigned integer
	if port < 0 || port > 65535 {
		log.Fatalf("Parsed port number %d is out of valid range (0-65535)", port)
	}
	if err := Wait4Port("http://"+l.Addr().String(), WithRetry(retry), WithSleeping(sleeping)); err != nil {
		log.Fatalf("Could not connect to proxy: %s", err)
	}

	return "http://" + testcontainers.HostInternal + ":" + portStr, int(port), func() {
		_ = srv.Shutdown(context.Background())
		_ = l.Close()
	}
}

// Wait4Port checks if a network service is available at the given address.
// It retries according to the provided options.
// This is useful for ensuring that servers are ready before connecting to them.
//
// Parameters:
//   - addr: The URL to check (e.g. "http://localhost:8080")
//   - opts: Configuration options for retries and timeout
func Wait4Port(addr string, opts ...Option) error {
	c := &config{
		sleeping: 200 * time.Millisecond,
		retry:    15,
		ctx:      context.Background(),
	}
	for _, opt := range opts {
		opt.apply(c)
	}

	if err := SleepWithContext(c.ctx, c.sleeping); err != nil {
		return err
	}
	for i := 0; i < c.retry; i++ {
		req, err := http.NewRequestWithContext(c.ctx, "GET", addr, nil)
		if err != nil {
			return fmt.Errorf("could not create a request to %s: %w", addr, err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("could not connect to", addr, "yet, error", err, "retrying in", c.sleeping)
			if err := SleepWithContext(c.ctx, c.sleeping); err != nil {
				return err
			}
			continue
		}
		if err := resp.Body.Close(); err != nil {
			log.Println("could not close response body", err)
			return err
		}
		return nil
	}
	return fmt.Errorf("could not connect to %s after retry and timeout", addr)
}

// SleepWithContext sleeps for the specified duration or until the context is canceled.
// It returns nil if the sleep completes or the context's error if canceled early.
func SleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
