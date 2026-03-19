package vm

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLAUnsignedError_Error(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		err     CLAUnsignedError
		wantMsg string
	}{
		{
			name:    "normal address",
			err:     CLAUnsignedError{Address: "g1abc123"},
			wantMsg: "address g1abc123 has not signed the required CLA",
		},
		{
			name:    "empty address",
			err:     CLAUnsignedError{Address: ""},
			wantMsg: "address  has not signed the required CLA",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMsg, tt.err.Error())
		})
	}
}

func TestCLAUnsignedError_Is(t *testing.T) {
	t.Parallel()
	err := CLAUnsignedError{Address: "g1abc", RealmPath: "gno.land/r/sys/cla", Hash: "h1"}

	// Should match zero-value CLAUnsignedError
	assert.True(t, err.Is(CLAUnsignedError{}))
	// Should match CLAUnsignedError with different fields
	assert.True(t, err.Is(CLAUnsignedError{Address: "different"}))
	// Should NOT match other error types
	assert.False(t, err.Is(UnauthorizedUserError{}))
	assert.False(t, err.Is(InvalidPkgPathError{}))
}

func TestCLAUnsignedError_WrappedCause(t *testing.T) {
	t.Parallel()
	claErr := CLAUnsignedError{Address: "g1abc", RealmPath: "gno.land/r/sys/cla", Hash: "h1"}
	wrapped := errors.Wrap(claErr, claErr.Error())

	// errors.Cause should return the original CLAUnsignedError
	cause := errors.Cause(wrapped)
	extracted, ok := cause.(CLAUnsignedError)
	require.True(t, ok, "cause should be CLAUnsignedError")
	assert.Equal(t, "g1abc", extracted.Address)
	assert.Equal(t, "h1", extracted.Hash)
}

func TestCLAUnsignedError_InfoKV(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      CLAUnsignedError
		contains []string
		excludes []string
	}{
		{
			name: "all fields populated",
			err: CLAUnsignedError{
				RealmPath: "gno.land/r/sys/cla",
				Hash:      "abc123",
				URL:       "https://example.com/cla",
			},
			contains: []string{
				"cla.realm=gno.land/r/sys/cla",
				"cla.hash=abc123",
				"cla.url=https://example.com/cla",
			},
		},
		{
			name: "no hash - should omit cla.hash",
			err: CLAUnsignedError{
				RealmPath: "gno.land/r/sys/cla",
				URL:       "https://example.com/cla",
			},
			contains: []string{"cla.realm=gno.land/r/sys/cla", "cla.url=https://example.com/cla"},
			excludes: []string{"cla.hash="},
		},
		{
			name: "no URL - should omit cla.url",
			err: CLAUnsignedError{
				RealmPath: "gno.land/r/sys/cla",
				Hash:      "abc123",
			},
			contains: []string{"cla.realm=gno.land/r/sys/cla", "cla.hash=abc123"},
			excludes: []string{"cla.url="},
		},
		{
			name: "only realm",
			err: CLAUnsignedError{
				RealmPath: "gno.land/r/sys/cla",
			},
			contains: []string{"cla.realm=gno.land/r/sys/cla"},
			excludes: []string{"cla.hash=", "cla.url="},
		},
		{
			name: "URL with query params",
			err: CLAUnsignedError{
				RealmPath: "gno.land/r/sys/cla",
				Hash:      "h1",
				URL:       "https://example.com/cla?a=1&b=2",
			},
			contains: []string{"cla.url=https://example.com/cla?a=1&b=2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.InfoKV()
			for _, c := range tt.contains {
				assert.Contains(t, result, c)
			}
			for _, e := range tt.excludes {
				assert.NotContains(t, result, e)
			}
			// Should never end with a newline (TrimRight is applied)
			if len(result) > 0 {
				assert.NotEqual(t, byte('\n'), result[len(result)-1], "InfoKV should not end with newline")
			}
		})
	}
}

func TestCLAUnsignedError_InfoKV_NewlineInjection(t *testing.T) {
	t.Parallel()
	// A malicious hash containing newlines should be sanitized so that
	// the injected content does NOT appear as a separate line.
	err := CLAUnsignedError{
		RealmPath: "gno.land/r/sys/cla",
		Hash:      "abc\ncla.url=https://evil.com",
		URL:       "https://example.com\ncla.hash=overridden",
	}
	result := err.InfoKV()

	// Newlines should be stripped — no injected lines
	for _, line := range strings.Split(result, "\n") {
		// Each line should start with "cla." — no raw injected lines
		assert.True(t, strings.HasPrefix(line, "cla."), "unexpected line: %q", line)
	}
	// The hash value should be concatenated (newline removed), not split
	assert.Contains(t, result, "cla.hash=abccla.url=https://evil.com")
}

func TestCLAUnsignedError_InfoKV_CarriageReturn(t *testing.T) {
	t.Parallel()
	// Carriage returns should also be stripped
	err := CLAUnsignedError{
		RealmPath: "gno.land/r/sys/cla",
		Hash:      "abc\r\nmore",
	}
	result := err.InfoKV()
	assert.NotContains(t, result, "\r")
	// The hash should be "abcmore" (both \r and \n stripped)
	assert.Contains(t, result, "cla.hash=abcmore")
}

func TestAbciResult_WithCLAError(t *testing.T) {
	t.Parallel()
	claErr := CLAUnsignedError{
		Address:   "g1abc",
		RealmPath: "gno.land/r/sys/cla",
		Hash:      "hash123",
		URL:       "https://example.com/cla",
	}
	wrapped := errors.Wrap(claErr, claErr.Error())
	result := abciResult(wrapped)

	// Info should contain CLA metadata
	assert.Contains(t, result.Info, "cla.realm=gno.land/r/sys/cla")
	assert.Contains(t, result.Info, "cla.hash=hash123")
	assert.Contains(t, result.Info, "cla.url=https://example.com/cla")
	// Info should also contain vm.version
	assert.Contains(t, result.Info, "vm.version=")
}

func TestAbciResult_WithNonCLAError(t *testing.T) {
	t.Parallel()
	err := ErrUnauthorizedUser("some error")
	result := abciResult(err)

	// Should NOT contain CLA metadata
	assert.NotContains(t, result.Info, "cla.")
	// Should contain vm.version
	assert.Contains(t, result.Info, "vm.version=")
}
