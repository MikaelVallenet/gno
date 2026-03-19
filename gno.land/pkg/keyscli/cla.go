package keyscli

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

const (
	claErrorSubstring  = "has not signed the required CLA"
	sysCLARealmDefault = "gno.land/r/sys/cla"
)

// isCLAError checks the verbose error output for a CLA signing failure.
func isCLAError(err error) bool {
	return strings.Contains(fmt.Sprintf("%#v", err), claErrorSubstring)
}

// queryCLARealmPath returns the CLA realm path from chain params, or the default on failure.
func queryCLARealmPath(remote string) string {
	cfg := &client.QueryCfg{
		RootCfg: &client.BaseCfg{BaseOptions: client.BaseOptions{Remote: remote}},
		Path:    "params/params/vm:p:syscla_pkgpath",
	}
	res, err := client.QueryHandler(cfg)
	if err != nil || res.Response.Error != nil || len(res.Response.Data) == 0 {
		return sysCLARealmDefault
	}
	path := string(res.Response.Data)
	if path == "" {
		return sysCLARealmDefault
	}
	return path
}

// queryCLAInfo returns the required hash and URL from the CLA realm.
func queryCLAInfo(remote, claRealmPath string) (hash, url string) {
	hash = queryEvalString(remote, claRealmPath, "requiredHash")
	url = queryEvalString(remote, claRealmPath, "claURL")
	return
}

// queryEvalString evaluates an expression via vm/qeval and extracts the string result.
func queryEvalString(remote, pkgPath, expr string) string {
	cfg := &client.QueryCfg{
		RootCfg: &client.BaseCfg{BaseOptions: client.BaseOptions{Remote: remote}},
		Path:    "vm/qeval",
		Data:    pkgPath + "." + expr,
	}
	res, err := client.QueryHandler(cfg)
	if err != nil || res.Response.Error != nil {
		return ""
	}
	return parseQEvalString(string(res.Response.Data))
}

// parseQEvalString extracts the string value from a '("value" string)' qeval response.
func parseQEvalString(data string) string {
	const suffix = " string)"
	if !strings.HasSuffix(data, suffix) {
		return ""
	}
	// Strip the suffix and leading "("
	inner := data[:len(data)-len(suffix)]
	if !strings.HasPrefix(inner, "(") {
		return ""
	}
	inner = inner[1:]
	// Strip surrounding quotes if present
	if len(inner) >= 2 && inner[0] == '"' && inner[len(inner)-1] == '"' {
		inner = inner[1 : len(inner)-1]
	}
	return inner
}

// formatCLAHelper builds a user-friendly CLA signing hint, or "" if hash is empty.
func formatCLAHelper(hash, url, claRealmPath, chainID, remote, nameOrBech32 string) string {
	if hash == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("A Contributor License Agreement (CLA) must be signed before deploying packages.\n")
	b.WriteString("It grants the necessary rights for your code to be used on-chain.\n")
	b.WriteString("The CLA document is defined through a GovDAO governance proposal.\n")
	if url != "" {
		fmt.Fprintf(&b, "\nCLA document: %s\n", url)
	}
	fmt.Fprintf(&b, "\nTo sign the CLA, run:\n\n")
	fmt.Fprintf(&b, "  gnokey maketx call -pkgpath %s -func Sign -args %s", claRealmPath, hash)
	fmt.Fprintf(&b, " -gas-fee 100000ugnot -gas-wanted 2000000 -broadcast")
	if remote != "" {
		fmt.Fprintf(&b, " -remote %s", remote)
	}
	if chainID != "" {
		fmt.Fprintf(&b, " -chainid %s", chainID)
	}
	fmt.Fprintf(&b, " %s\n", nameOrBech32)
	return b.String()
}
