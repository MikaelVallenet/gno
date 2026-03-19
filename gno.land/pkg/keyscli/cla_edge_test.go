package keyscli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCLAInfo_NewlineInjection(t *testing.T) {
	// If hash or URL values contain newlines, the parser could interpret
	// them as additional key-value pairs.
	info := "cla.realm=gno.land/r/sys/cla\ncla.hash=abc\ninjected=bad\ncla.url=https://example.com"
	result := parseCLAInfo(info)

	// "injected=bad" is NOT prefixed with "cla." so should be ignored
	assert.Empty(t, result["injected"])
	// But "cla.hash" gets truncated to just "abc" (the newline splits it)
	assert.Equal(t, "abc", result["hash"])
}

func TestParseCLAInfo_DuplicateKeys(t *testing.T) {
	// Last value should win
	info := "cla.hash=first\ncla.hash=second"
	result := parseCLAInfo(info)
	assert.Equal(t, "second", result["hash"])
}

func TestParseCLAInfo_EmptyValue(t *testing.T) {
	info := "cla.realm=\ncla.hash="
	result := parseCLAInfo(info)
	assert.Equal(t, "", result["realm"])
	assert.Equal(t, "", result["hash"])
}

func TestParseCLAInfo_NoEqualsSign(t *testing.T) {
	info := "cla.realm\ncla.hash=abc"
	result := parseCLAInfo(info)
	assert.Empty(t, result["realm"]) // skipped because no "="
	assert.Equal(t, "abc", result["hash"])
}

func TestParseCLAInfo_ValueWithEquals(t *testing.T) {
	info := "cla.url=https://example.com/cla?v=1&t=2"
	result := parseCLAInfo(info)
	assert.Equal(t, "https://example.com/cla?v=1&t=2", result["url"])
}

func TestParseCLAInfo_EmptyString(t *testing.T) {
	result := parseCLAInfo("")
	assert.Empty(t, result)
}

func TestFormatCLAHint_EmptyChainID(t *testing.T) {
	info := "cla.realm=gno.land/r/sys/cla\ncla.hash=abc123"
	hint := formatCLAHint(info, "", "g1abc")

	assert.Contains(t, hint, "To sign the CLA, run:")
	assert.Contains(t, hint, "-args abc123")
	// Should NOT contain -chainid when empty
	assert.NotContains(t, hint, "-chainid")
}

func TestFormatCLAHint_EmptyNameOrBech32(t *testing.T) {
	info := "cla.realm=gno.land/r/sys/cla\ncla.hash=abc123"
	hint := formatCLAHint(info, "testchain", "")

	// Should still produce a hint, but with empty user
	assert.Contains(t, hint, "To sign the CLA, run:")
}

func TestFormatCLAHint_RealmWithoutHash(t *testing.T) {
	// Realm exists but hash is empty - should return empty
	info := "cla.realm=gno.land/r/sys/cla\ncla.hash="
	hint := formatCLAHint(info, "testchain", "g1abc")
	assert.Empty(t, hint)
}

func TestFormatCLAHint_CommandOnOneLine(t *testing.T) {
	info := "cla.realm=gno.land/r/sys/cla\ncla.hash=abc123hash\ncla.url=https://example.com/cla"
	hint := formatCLAHint(info, "testchain", "g1abc")

	// Verify the gnokey command is on a single line
	lines := strings.Split(hint, "\n")
	var cmdLine string
	for _, line := range lines {
		if strings.Contains(line, "gnokey maketx call") {
			cmdLine = line
			break
		}
	}
	assert.NotEmpty(t, cmdLine, "should contain a gnokey command line")
	assert.Contains(t, cmdLine, "-broadcast")
	assert.Contains(t, cmdLine, "g1abc")
}
