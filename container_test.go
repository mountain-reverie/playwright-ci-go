package playwrightcigo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseGoListJSONStreamEmpty(t *testing.T) {
	t.Parallel()

	jsonStream := ``

	found, result := parseGoListJSONStream(strings.NewReader(jsonStream), "none", true)
	assert.False(t, found)
	assert.Equal(t, "none", result)
}

func Test_parseGoListJSONStreamNoPlaywrightCIGo(t *testing.T) {
	t.Parallel()

	jsonStream := `
	{"Path":"github.com/some/other/package","Version":"v1.0.0","Main":false}
	{"Path":"github.com/another/package","Version":"v2.0.0","Main":true}
	`

	found, result := parseGoListJSONStream(strings.NewReader(jsonStream), "none", true)
	assert.False(t, found)
	assert.Equal(t, "none", result)
}

func Test_parseGoListJSONStreamPlaywrightCIGoMain(t *testing.T) {
	t.Parallel()

	jsonStream := `
	{"Path":"github.com/mountain-reverie/playwright-ci-go","Version":"v1.0.0","Main":true}
	`

	// As this call `git` if the command fail, it is possible that the result is not found
	_, result := parseGoListJSONStream(strings.NewReader(jsonStream), "none", true)
	assert.NotEqual(t, "v1.0.0", result)
}

func Test_parseGoListJSONStreamPlaywrightCIGoNotMain(t *testing.T) {
	t.Parallel()

	jsonStream := `
	{"Path":"github.com/mountain-reverie/playwright-ci-go","Version":"v1.0.0","Main":false}
	`

	found, result := parseGoListJSONStream(strings.NewReader(jsonStream), "none", true)
	assert.True(t, found)
	assert.Equal(t, "v1.0.0", result)
}

func Test_parseGoListJSONStreamPlaywrightCIGoNoVersion(t *testing.T) {
	t.Parallel()

	jsonStream := `
	{"Path":"github.com/mountain-reverie/playwright-ci-go","Version":"","Main":false}
	`

	found, result := parseGoListJSONStream(strings.NewReader(jsonStream), "none", true)
	assert.False(t, found)
	assert.Equal(t, "none", result)
}

func Test_parseGoListJSONStreamInvalidStream(t *testing.T) {
	t.Parallel()

	jsonStream := `
	{"Invalid JSON"}
	{"Path":"github.com/mountain-reverie/playwright-ci-go","Version":"v1.0.0","Main":false}
	`

	found, result := parseGoListJSONStream(strings.NewReader(jsonStream), "none", true)
	assert.False(t, found)
	assert.Equal(t, "none", result)
}

func Test_BuildInfoPath(t *testing.T) {
	t.Parallel()

	ok, version := getPlaywrightCIGoFromBuildInfo("none", true)
	assert.False(t, ok)
	assert.Equal(t, "none", version)
}

func Test_GoListInfo(t *testing.T) {
	t.Parallel()

	ok, version := getPlaywrightCIGoFromGoList("none", true)
	assert.True(t, ok)
	assert.NotEqual(t, "none", version)
	assert.NotEmpty(t, version)
	assert.Greater(t, len(version), 3)
	assert.Equal(t, "v0.", version[:3])
}

func Test_NoTag(t *testing.T) {
	t.Parallel()

	tag, err := noTagVersion("0.51.01", true)
	assert.NoError(t, err)
	assert.NotEqual(t, "v0.5101.0", tag)
	assert.Greater(t, len(tag), 3)
	assert.Equal(t, "v0.", tag[:3])
}
