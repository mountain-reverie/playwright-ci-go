package playwrightcigo_test

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	playwrightcigo "github.com/mountain-reverie/playwright-ci-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Firefox(t *testing.T) {
	// Install really need to be done once and it is better done in TestMain
	err := playwrightcigo.Install(playwrightcigo.WithTimeout(time.Minute))
	assert.NoError(t, err)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not listen: %v", err)
	}

	base := "http://" + l.Addr().String()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World!"))
	})
	srv := &http.Server{}

	go func() {
		err := srv.Serve(l)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not serve: %v", err)
		}
	}()
	defer func() { _ = srv.Shutdown(context.Background()) }()

	err = playwrightcigo.Wait4Port(base)
	require.NoError(t, err)

	browser, err := playwrightcigo.Firefox()
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

	err = browser.Close()
	require.NoError(t, err)

	err = playwrightcigo.Uninstall()
	assert.NoError(t, err)
}
