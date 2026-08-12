package playwrightcigo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseGoListJSONStreamEmpty(t *testing.T) {
	t.Parallel()

	jsonStream := ``

	result, found := parseGoListJSONStream(strings.NewReader(jsonStream), true)
	assert.False(t, found)
	assert.Empty(t, result)
}

func Test_parseGoListJSONStreamNoPlaywrightCIGo(t *testing.T) {
	t.Parallel()

	jsonStream := `
	{"Path":"github.com/some/other/package","Version":"v1.0.0","Main":false}
	{"Path":"github.com/another/package","Version":"v2.0.0","Main":true}
	`

	result, found := parseGoListJSONStream(strings.NewReader(jsonStream), true)
	assert.False(t, found)
	assert.Empty(t, result)
}

func Test_parseGoListJSONStreamPlaywrightCIGoMain(t *testing.T) {
	t.Parallel()

	jsonStream := `
	{"Path":"github.com/mountain-reverie/playwright-ci-go","Version":"v1.0.0","Main":true}
	`

	// As this call `git` if the command fail, it is possible that the result is not found
	result, _ := parseGoListJSONStream(strings.NewReader(jsonStream), true)
	assert.NotEqual(t, "v1.0.0", result)
}

func Test_parseGoListJSONStreamPlaywrightCIGoNotMain(t *testing.T) {
	t.Parallel()

	jsonStream := `
	{"Path":"github.com/mountain-reverie/playwright-ci-go","Version":"v1.0.0","Main":false}
	`

	result, found := parseGoListJSONStream(strings.NewReader(jsonStream), true)
	assert.True(t, found)
	assert.Equal(t, "v1.0.0", result)
}

func Test_parseGoListJSONStreamPlaywrightCIGoNoVersion(t *testing.T) {
	t.Parallel()

	jsonStream := `
	{"Path":"github.com/mountain-reverie/playwright-ci-go","Version":"","Main":false}
	`

	result, found := parseGoListJSONStream(strings.NewReader(jsonStream), true)
	assert.False(t, found)
	assert.Empty(t, result)
}

func Test_parseGoListJSONStreamInvalidStream(t *testing.T) {
	t.Parallel()

	jsonStream := `
	{"Invalid JSON"}
	{"Path":"github.com/mountain-reverie/playwright-ci-go","Version":"v1.0.0","Main":false}
	`

	result, found := parseGoListJSONStream(strings.NewReader(jsonStream), true)
	assert.False(t, found)
	assert.Empty(t, result)
}

func Test_BuildInfoPath(t *testing.T) {
	t.Parallel()

	// A test binary of the main module carries no dependency info, so this
	// resolves nothing here. Consumers of the library do have it, which
	// TestVersionFromBuildInfo covers against a synthetic BuildInfo.
	_, ok := getPlaywrightCIGoFromBuildInfo(true)
	assert.False(t, ok)
}

func Test_GoListInfo(t *testing.T) {
	t.Parallel()

	version, ok := getPlaywrightCIGoFromGoList(true)
	assert.True(t, ok)
	assert.NotEmpty(t, version)
	assert.Greater(t, len(version), 3)
	assert.Equal(t, "v0.", version[:3])
}

func Test_NoTag(t *testing.T) {
	t.Parallel()

	tag, err := noTagVersion(true)
	assert.NoError(t, err)
	assert.Greater(t, len(tag), 3)
	assert.Equal(t, "v0.", tag[:3])
}
