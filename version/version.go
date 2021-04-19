package version

import (
	"strings"
)

const defaultBranch   = "develop"

var (
	// ProviderVersion is set to the release version of
	// the binary during the automated release process.
	ProviderVersion = "dev"
	// Branch is current provider git branch
	Branch = "dev"
)


func init() {
	if Branch != defaultBranch {
		ProviderVersion = strings.Replace(Branch, "/", "-", -1)
	}
}
