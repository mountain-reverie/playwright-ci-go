package playwrightcigo

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaywrightGoTagLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		want    string
		ok      bool
	}{
		{"first of a line", "v0.6100.0", "v0.6100.0", true},
		{"later patch maps back to the line start", "v0.6100.7", "v0.6100.0", true},
		{"previous line", "v0.5700.1", "v0.5700.0", true},
		{"no v prefix", "0.6100.0", "", false},
		{"too few parts", "v0.6100", "", false},
		{"empty", "", "", false},
		{"devel", "(devel)", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, ok := playwrightGoTagLine(test.version)
			assert.Equal(t, test.ok, ok)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestVersionFromBuildInfo(t *testing.T) {
	t.Parallel()

	dep := func(path, version string) *debug.Module {
		return &debug.Module{Path: path, Version: version}
	}

	tests := []struct {
		name string
		info *debug.BuildInfo
		want string
		ok   bool
	}{
		{
			// A consumer of this library: its own pinned version is the
			// exact image tag, so prefer it over anything derived.
			name: "playwright-ci-go dependency wins",
			info: &debug.BuildInfo{Deps: []*debug.Module{
				dep("github.com/mxschmitt/playwright-go", "v0.6100.0"),
				dep("github.com/mountain-reverie/playwright-ci-go", "v0.6100.7"),
			}},
			want: "v0.6100.7",
			ok:   true,
		},
		{
			// No playwright-ci-go entry, so fall back to the release line
			// implied by playwright-go's module version.
			name: "playwright-go alone gives the line start",
			info: &debug.BuildInfo{Deps: []*debug.Module{
				dep("github.com/mxschmitt/playwright-go", "v0.6100.0"),
			}},
			want: "v0.6100.0",
			ok:   true,
		},
		{
			name: "archived module path still resolves",
			info: &debug.BuildInfo{Deps: []*debug.Module{
				dep("github.com/playwright-community/playwright-go", "v0.5700.1"),
			}},
			want: "v0.5700.0",
			ok:   true,
		},
		{
			name: "replaced playwright-ci-go is ignored",
			info: &debug.BuildInfo{Deps: []*debug.Module{
				dep("github.com/mountain-reverie/playwright-ci-go", "(devel)"),
				dep("github.com/mxschmitt/playwright-go", "v0.6100.0"),
			}},
			want: "v0.6100.0",
			ok:   true,
		},
		{
			name: "no build info",
			info: nil,
			want: "",
			ok:   false,
		},
		{
			name: "nothing relevant",
			info: &debug.BuildInfo{Deps: []*debug.Module{dep("github.com/other/thing", "v1.0.0")}},
			want: "",
			ok:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, ok := versionFromBuildInfo(test.info, true)
			assert.Equal(t, test.ok, ok)
			assert.Equal(t, test.want, got)
		})
	}
}
