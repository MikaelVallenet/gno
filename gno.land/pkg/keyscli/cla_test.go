package keyscli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCLAInfo(t *testing.T) {
	info := "cla.realm=gno.land/r/sys/cla\ncla.hash=abc123\ncla.url=https://example.com/cla\nvm.version=0.1"
	result := parseCLAInfo(info)

	assert.Equal(t, "gno.land/r/sys/cla", result["realm"])
	assert.Equal(t, "abc123", result["hash"])
	assert.Equal(t, "https://example.com/cla", result["url"])
	assert.Empty(t, result["version"]) // vm.version is not a cla.* key
}

func TestParseCLAInfo_Empty(t *testing.T) {
	result := parseCLAInfo("vm.version=0.1")
	assert.Empty(t, result)
}

func TestFormatCLAHint(t *testing.T) {
	info := "cla.realm=gno.land/r/sys/cla\ncla.hash=abc123hash\ncla.url=https://example.com/cla\nvm.version=0.1"
	hint := formatCLAHelper(info, "testchain", "g1abc")

	assert.Contains(t, hint, "CLA document: https://example.com/cla")
	assert.Contains(t, hint, "To sign the CLA, run:")
	assert.Contains(t, hint, "-pkgpath gno.land/r/sys/cla")
	assert.Contains(t, hint, "-func Sign")
	assert.Contains(t, hint, "-args abc123hash")
	assert.Contains(t, hint, "-chainid testchain")
	assert.Contains(t, hint, "g1abc")
}

func TestFormatCLAHint_NoCLAData(t *testing.T) {
	hint := formatCLAHelper("vm.version=0.1", "testchain", "g1abc")
	assert.Empty(t, hint)
}

func TestFormatCLAHint_NoURL(t *testing.T) {
	info := "cla.realm=gno.land/r/sys/cla\ncla.hash=abc123hash\nvm.version=0.1"
	hint := formatCLAHelper(info, "testchain", "g1abc")

	assert.NotContains(t, hint, "CLA document:")
	assert.Contains(t, hint, "To sign the CLA, run:")
	assert.Contains(t, hint, "-args abc123hash")
}

func TestFormatCLAHint_NoHash(t *testing.T) {
	info := "cla.realm=gno.land/r/sys/cla\ncla.url=https://example.com/cla\nvm.version=0.1"
	hint := formatCLAHelper(info, "testchain", "g1abc")
	assert.Empty(t, hint) // Can't sign without a hash
}
