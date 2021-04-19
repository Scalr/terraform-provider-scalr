package version

import (
	"strings"
)

const defaultBranch = "develop"

var (
	// ProviderVersion is set to the release version of
	// the binary during the automated release process.
	ProviderVersion = "dev"
	// Branch is current provider git branch
	Branch = "dev"
)

func init() {
	// Empty branch could we if git repo was checkouted to tag
	if Branch != defaultBranch && Branch != "" {
		ProviderVersion = strings.Replace(Branch, "/", "-", -1)
	}
}
