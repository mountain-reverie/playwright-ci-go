package playwrightcigo_test

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	playwrightcigo "github.com/mountain-reverie/playwright-ci-go"
	"github.com/playwright-community/playwright-go"
)

func BenchmarkPageChromium(b *testing.B) {
	teardown, base := setupBenchmark(b)
	defer teardown()

	log.Println("Benchmarking Chromium...")
	for b.Loop() {
		checkPage(b, playwrightcigo.Chromium, base)
	}
	log.Println("Chromium benchmark completed.")
}

func BenchmarkPageFirefox(b *testing.B) {
	teardown, base := setupBenchmark(b)
	defer teardown()

	log.Println("Benchmarking Firefox...")
	for b.Loop() {
		checkPage(b, playwrightcigo.Firefox, base)
	}
	log.Println("Firefox benchmark completed.")
}

func BenchmarkPageWebkit(b *testing.B) {
	teardown, base := setupBenchmark(b)
	defer teardown()

	log.Println("Benchmarking Webkit...")
	for b.Loop() {
		checkPage(b, playwrightcigo.Webkit, base)
	}
	log.Println("Webkit benchmark completed.")
}

func setupBenchmark(b *testing.B) (func(), string) {
	if err := playwrightcigo.Install(playwrightcigo.WithTimeout(10 * time.Minute)); err != nil {
		b.Fatalf("could not install playwright: %v", err)
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("could not listen: %v", err)
	}

	base := "http://" + l.Addr().String()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World!"))
	})
	srv := &http.Server{
		Handler: mux,
	}

	go func() {
		err := srv.Serve(l)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not serve: %v", err)
		}
	}()

	if err := playwrightcigo.Wait4Port(base); err != nil {
		b.Fatalf("could not wait for port: %v", err)
	}

	return func() {
		if err := srv.Shutdown(context.Background()); err != nil {
			b.Fatalf("could not shutdown server: %v", err)
		}
		if err := playwrightcigo.Uninstall(); err != nil {
			b.Fatalf("could not uninstall playwright: %v", err)
		}
	}, base
}

func checkPage(b *testing.B, createBrowser func() (playwright.Browser, error), base string) {
	b.Helper()

	browser, err := createBrowser()
	if err != nil {
		b.Fatalf("could not create browser: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		b.Fatalf("could not create page: %v", err)
	}

	if _, err := page.Goto(base); err != nil {
		b.Fatalf("could not go to base: %v", err)
	}

	content, err := page.Content()
	if err != nil {
		b.Fatalf("could not get content: %v", err)
	}

	if !strings.Contains(content, "Hello World!") {
		b.Fatalf("content does not contain 'Hello World!': %v", content)
	}

	if err := page.Close(); err != nil {
		b.Fatalf("could not close page: %v", err)
	}

	if err := browser.Close(); err != nil {
		b.Fatalf("could not close browser: %v", err)
	}
}
