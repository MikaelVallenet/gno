package keyscli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCLAError(t *testing.T) {
	assert.True(t, isCLAError(fmt.Errorf("address g1abc has not signed the required CLA")))
	assert.True(t, isCLAError(fmt.Errorf("deliver transaction failed: log:address g1abc has not signed the required CLA")))
	assert.False(t, isCLAError(fmt.Errorf("unauthorized user")))
	assert.False(t, isCLAError(fmt.Errorf("")))
}

func TestParseQEvalString(t *testing.T) {
	// Quoted format (actual vm/qeval output)
	assert.Equal(t, "abc123hash", parseQEvalString(`("abc123hash" string)`))
	assert.Equal(t, "https://example.com/cla", parseQEvalString(`("https://example.com/cla" string)`))
	// Empty quoted string
	assert.Equal(t, "", parseQEvalString(`("" string)`))
	// Non-string type
	assert.Equal(t, "", parseQEvalString("(true bool)"))
	// Garbage
	assert.Equal(t, "", parseQEvalString(""))
	assert.Equal(t, "", parseQEvalString("garbage"))
	// Unquoted fallback (shouldn't happen in practice but handle gracefully)
	assert.Equal(t, "abc123hash", parseQEvalString("(abc123hash string)"))
}

func TestFormatCLAHelper(t *testing.T) {
	hint := formatCLAHelper("abc123hash", "https://example.com/cla", "gno.land/r/sys/cla", "testchain", "localhost:26657", "g1abc")

	assert.Contains(t, hint, "CLA document: https://example.com/cla")
	assert.Contains(t, hint, "To sign the CLA, run:")
	assert.Contains(t, hint, "-pkgpath gno.land/r/sys/cla")
	assert.Contains(t, hint, "-func Sign")
	assert.Contains(t, hint, "-args abc123hash")
	assert.Contains(t, hint, "-chainid testchain")
	assert.Contains(t, hint, "g1abc")
}

func TestFormatCLAHelper_NoHash(t *testing.T) {
	assert.Empty(t, formatCLAHelper("", "https://example.com", "gno.land/r/sys/cla", "chain", "localhost:26657", "g1abc"))
}

func TestFormatCLAHelper_NoURL(t *testing.T) {
	hint := formatCLAHelper("abc123", "", "gno.land/r/sys/cla", "chain", "localhost:26657", "g1abc")
	assert.NotContains(t, hint, "CLA document:")
	assert.Contains(t, hint, "To sign the CLA, run:")
}

func TestFormatCLAHelper_NoChainID(t *testing.T) {
	hint := formatCLAHelper("abc123", "", "gno.land/r/sys/cla", "", "", "g1abc")
	assert.NotContains(t, hint, "-chainid")
}
