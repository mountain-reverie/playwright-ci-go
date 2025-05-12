package playwrightcigo

import (
	"errors"
	"fmt"
	"image"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "image/png"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_HelloWorld(t *testing.T) {
	t.Parallel()

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

	err = Wait4Port(base)
	require.NoError(t, err)

	tests := []struct {
		browser     string
		instantiate func() (playwright.Browser, error)
	}{
		{"chromium", Chromium},
		{"firefox", Firefox},
		{"webkit", Webkit},
	}
	for _, test := range tests {
		t.Run(test.browser, func(t *testing.T) {
			t.Parallel()

			browser, err := test.instantiate()
			require.NoError(t, err)

			page, err := browser.NewPage()
			require.NoError(t, err)

			_, err = page.Goto(base)
			require.NoError(t, err)

			content, err := page.Content()
			require.NoError(t, err)

			require.Contains(t, content, "Hello World!")

			testresult := filepath.Join("testdata", "failed", fmt.Sprintf("screenshot-%s.png", test.browser))

			err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateLoad})
			assert.NoError(t, err)

			_, err = page.Screenshot(playwright.PageScreenshotOptions{
				Path: playwright.String(testresult),
			})
			assert.NoError(t, err)

			expectedSize, expectedPixels := pixels(t, filepath.Join("testdata", fmt.Sprintf("screenshot-%s.png", test.browser)))
			actualSize, actualPixels := pixels(t, testresult)

			require.Equal(t, expectedSize, actualSize)
			assert.Equal(t, expectedPixels, actualPixels)
			if !t.Failed() {
				err = os.Remove(testresult)
				assert.NoError(t, err)
			}

			err = page.Close()
			require.NoError(t, err)

			err = browser.Close()
			require.NoError(t, err)
		})
	}
}

func Test_OverlapLifecycle(t *testing.T) {
	t.Parallel()

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

	err = Wait4Port(base)
	require.NoError(t, err)

	tests := []struct {
		browser     string
		instantiate func() (playwright.Browser, error)
	}{
		{"chromium", Chromium},
		{"firefox", Firefox},
		{"webkit", Webkit},
	}
	for _, test := range tests {
		t.Run(test.browser, func(t *testing.T) {
			t.Parallel()

			browser1, err := test.instantiate()
			require.NoError(t, err)

			browser2, err := test.instantiate()
			require.NoError(t, err)

			page, err := browser1.NewPage()
			require.NoError(t, err)

			page2, err := browser2.NewPage()
			require.NoError(t, err)

			_, err = page.Goto(base)
			require.NoError(t, err)

			_, err = page2.Goto(base)
			require.NoError(t, err)

			content, err := page.Content()
			require.NoError(t, err)
			require.Contains(t, content, "Hello World!")

			content2, err := page2.Content()
			require.NoError(t, err)
			require.Contains(t, content2, "Hello World!")

			testresult := filepath.Join("testdata", "failed", fmt.Sprintf("screenshot-%s.png", test.browser))
			err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateLoad})
			assert.NoError(t, err)

			_, err = page.Screenshot(playwright.PageScreenshotOptions{
				Path: playwright.String(testresult),
			})
			assert.NoError(t, err)

			expectedSize, expectedPixels := pixels(t, filepath.Join("testdata", fmt.Sprintf("screenshot-%s.png", test.browser)))
			actualSize, actualPixels := pixels(t, testresult)

			require.Equal(t, expectedSize, actualSize)
			assert.Equal(t, expectedPixels, actualPixels)
			if !t.Failed() {
				err = os.Remove(testresult)
				assert.NoError(t, err)
			}

			err = page.Close()
			require.NoError(t, err)

			err = browser1.Close()
			require.NoError(t, err)

			content2, err = page2.Content()
			require.NoError(t, err)
			require.Contains(t, content2, "Hello World!")

			err = page2.Close()
			require.NoError(t, err)

			page2, err = browser2.NewPage()
			require.NoError(t, err)

			_, err = page2.Goto(base)
			require.NoError(t, err)

			content2, err = page2.Content()
			require.NoError(t, err)
			require.Contains(t, content2, "Hello World!")

			err = page2.Close()
			require.NoError(t, err)

			err = browser2.Close()
			require.NoError(t, err)
		})
	}
}

func TestMain(m *testing.M) {
	if err := os.MkdirAll("testdata/failed", 0755); err != nil {
		log.Fatalf("could not create directory: %v", err)
	}

	if err := Install(WithRepository(os.Getenv("PLAYWRIGHTCI_REPOSITORY"), os.Getenv("PLAYWRIGHTCI_TAG")), WithTimeout(time.Minute)); err != nil {
		log.Fatalf("could not install playwright ci: %v", err)
	}

	code := m.Run()

	if err := Uninstall(); err != nil {
		log.Fatalf("could not uninstall playwright ci: %v", err)
	}
	os.Exit(code)
}

func pixels(t *testing.T, path string) (image.Rectangle, []uint8) {
	f, err := os.Open(path)
	assert.NoError(t, err)
	defer func() { _ = f.Close() }()

	raw, _, err := image.Decode(f)
	assert.NoError(t, err)

	var pixels []uint8
	switch raw := raw.(type) {
	case *image.RGBA:
		pixels = raw.Pix
	case *image.NRGBA:
		pixels = raw.Pix
	default:
		t.Fatalf("unsupported image type: %T", raw)
	}

	return raw.Bounds(), pixels
}
