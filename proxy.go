package playwrightcigo

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/testcontainers/testcontainers-go"
)

func transparentProxy() (string, int, func()) {
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
		srv.Serve(l)
	}()

	split := strings.Split(l.Addr().String(), ":")
	port, _ := strconv.ParseInt(split[1], 10, 64)

	if err := wait4port("http://" + l.Addr().String()); err != nil {
		log.Fatalf("Could not connect to proxy: %s", err)
	}

	return "http://" + testcontainers.HostInternal + ":" + split[1], int(port), func() {
		srv.Shutdown(context.Background())
		l.Close()
	}
}

func wait4port(addr string) error {
	time.Sleep(time.Second)
	for i := 0; i < 15; i++ {
		resp, err := http.Get(addr)
		if err != nil {
			t := 200 * time.Millisecond
			log.Println("could not connect to", addr, "error", err, "retrying in", t)
			time.Sleep(t)
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
