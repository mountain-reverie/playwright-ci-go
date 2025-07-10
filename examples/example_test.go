package examples

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	playwrightcigo "github.com/mountain-reverie/playwright-ci-go"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

var browser playwright.Browser

func Test_Firefox(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not listen: %v", err)
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
	defer func() { _ = srv.Shutdown(context.Background()) }()

	err = playwrightcigo.Wait4Port(base)
	require.NoError(t, err)

	page, err := browser.NewPage()
	require.NoError(t, err)

	_, err = page.Goto(base)
	require.NoError(t, err)

	content, err := page.Content()
	require.NoError(t, err)

	require.Contains(t, content, "Hello World!")

	err = page.Close()
	require.NoError(t, err)
}

func TestMain(m *testing.M) {
	// Install once Playwright before running tests
	err := playwrightcigo.Install(playwrightcigo.WithTimeout(time.Minute), playwrightcigo.WithVerbose())
	if err != nil {
		log.Fatalf("could not install playwright: %v", err)
	}

	browser, err = playwrightcigo.Firefox()
	if err != nil {
		log.Fatalf("could not instantiate Firefox browser: %v", err)
	}

	exitCode := m.Run()

	err = browser.Close()
	if err != nil {
		log.Fatalf("could not close browser: %v", err)
	}

	err = playwrightcigo.Uninstall()
	if err != nil {
		log.Fatalf("could not uninstall playwright: %v", err)
	}

	os.Exit(exitCode)
}
