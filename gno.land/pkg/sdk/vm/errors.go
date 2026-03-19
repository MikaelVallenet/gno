package vm

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/errors"
	"go.uber.org/multierr"
)

// for convenience:
type abciError struct{}

func (abciError) AssertABCIError() {}

// declare all script errors.
// NOTE: these are meant to be used in conjunction with pkgs/errors.
type (
	InvalidPkgPathError   struct{ abciError }
	NoRenderDeclError     struct{ abciError }
	PkgExistError         struct{ abciError }
	InvalidStmtError      struct{ abciError }
	InvalidExprError      struct{ abciError }
	UnauthorizedUserError struct{ abciError }
	InvalidPackageError   struct{ abciError }
	InvalidFileError      struct{ abciError }
	TypeCheckError        struct {
		abciError
		Errors []string `json:"errors"`
	}
)

func (e InvalidPkgPathError) Error() string   { return "invalid package path" }
func (e NoRenderDeclError) Error() string     { return "render function not declared" }
func (e PkgExistError) Error() string         { return "package already exists" }
func (e InvalidStmtError) Error() string      { return "invalid statement" }
func (e InvalidFileError) Error() string      { return "file is not available" }
func (e InvalidExprError) Error() string      { return "invalid expression" }
func (e UnauthorizedUserError) Error() string { return "unauthorized user" }
func (e InvalidPackageError) Error() string   { return "invalid package" }
func (e TypeCheckError) Error() string {
	var bld strings.Builder
	bld.WriteString("invalid gno package; type check errors:\n")
	bld.WriteString(strings.Join(e.Errors, "\n"))
	return bld.String()
}

func ErrPkgAlreadyExists(msg string) error {
	return errors.Wrap(PkgExistError{}, msg)
}

func ErrUnauthorizedUser(msg string) error {
	return errors.Wrap(UnauthorizedUserError{}, msg)
}

func ErrInvalidPkgPath(msg string) error {
	return errors.Wrap(InvalidPkgPathError{}, msg)
}

func ErrInvalidFile(msg string) error {
	return errors.Wrap(InvalidFileError{}, msg)
}

func ErrInvalidStmt(msg string) error {
	return errors.Wrap(InvalidStmtError{}, msg)
}

func ErrInvalidExpr(msg string) error {
	return errors.Wrap(InvalidExprError{}, msg)
}

func ErrInvalidPackage(msg string) error {
	return errors.Wrap(InvalidPackageError{}, msg)
}

// CLAUnsignedError is returned when a user tries to deploy a package without
// signing the required CLA. It carries CLA realm state so clients can build
// actionable hints (e.g. a gnokey command to sign).
type CLAUnsignedError struct {
	abciError
	Address   string `json:"address"`    // bech32 address of the deployer
	RealmPath string `json:"realm_path"` // CLA realm package path
	Hash      string `json:"hash"`       // required CLA hash
	URL       string `json:"url"`        // URL to the CLA document
}

func (e CLAUnsignedError) Error() string {
	return fmt.Sprintf("address %s has not signed the required CLA", e.Address)
}

// Is allows errors.Is(err, CLAUnsignedError{}) to match regardless of field values.
func (e CLAUnsignedError) Is(target error) bool {
	_, ok := target.(CLAUnsignedError)
	return ok
}

// InfoKV returns parseable key-value pairs for the ABCI Info field.
func (e CLAUnsignedError) InfoKV() string {
	var b strings.Builder
	fmt.Fprintf(&b, "cla.realm=%s\n", e.RealmPath)
	if e.Hash != "" {
		fmt.Fprintf(&b, "cla.hash=%s\n", e.Hash)
	}
	if e.URL != "" {
		fmt.Fprintf(&b, "cla.url=%s\n", e.URL)
	}
	return strings.TrimRight(b.String(), "\n")
}

func ErrTypeCheck(err error) error {
	var tce TypeCheckError
	errs := multierr.Errors(err)
	for _, err := range errs {
		tce.Errors = append(tce.Errors, err.Error())
	}
	return errors.NewWithData(tce).Stacktrace()
}
