package keyscli

import (
	"fmt"
	"strings"
)

// parseCLAInfo extracts CLA-related key-value pairs from the ABCI Info field.
// Returns a map of cla.* keys (without the "cla." prefix) to their values.
func parseCLAInfo(info string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(info, "\n") {
		if !strings.HasPrefix(line, "cla.") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		// Strip "cla." prefix for convenience.
		result[strings.TrimPrefix(k, "cla.")] = v
	}
	return result
}

// formatCLAHint builds a user-friendly CLA signing hint from ABCI Info metadata.
// Returns an empty string if the Info doesn't contain CLA data.
func formatCLAHint(info, chainID, nameOrBech32 string) string {
	claInfo := parseCLAInfo(info)
	realm, ok := claInfo["realm"]
	if !ok {
		return ""
	}
	hash := claInfo["hash"]
	if hash == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("A Contributor License Agreement (CLA) must be signed before deploying packages.\n")
	b.WriteString("It grants the necessary rights for your code to be used on-chain.\n")
	b.WriteString("The CLA document is defined through a GovDAO governance proposal.\n")
	if url, ok := claInfo["url"]; ok && url != "" {
		fmt.Fprintf(&b, "\nCLA document: %s\n", url)
	}
	fmt.Fprintf(&b, "\nTo sign the CLA, run:\n\n")
	fmt.Fprintf(&b, "  gnokey maketx call -pkgpath %s -func Sign -args %s", realm, hash)
	fmt.Fprintf(&b, " -gas-fee 100000ugnot -gas-wanted 2000000 -broadcast")
	if chainID != "" {
		fmt.Fprintf(&b, " -chainid %s", chainID)
	}
	fmt.Fprintf(&b, " %s\n", nameOrBech32)
	return b.String()
}
