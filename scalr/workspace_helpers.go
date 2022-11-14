package scalr

import (
	"context"
	"fmt"
	"strings"

	scalr "github.com/scalr/go-scalr"
)

// fetchWorkspaceID returns the id for a workspace
// when given a workspace id of the form ENVIRONMENT_ID/WORKSPACE_NAME
func fetchWorkspaceID(ctx context.Context, id string, client *scalr.Client) (string, error) {
	environmentID, wsName, err := unpackWorkspaceID(id)
	if err != nil {
		return "", fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	workspace, err := client.Workspaces.Read(ctx, environmentID, wsName)
	if err != nil {
		return "", fmt.Errorf("Error reading configuration of workspace %s: %v", id, err)
	}

	return workspace.ID, nil
}

func unpackWorkspaceID(id string) (environmentID, name string, err error) {
	// Support the old ID format for backwards compatibility.
	if s := strings.SplitN(id, "|", 2); len(s) == 2 {
		return s[1], s[0], nil
	}

	s := strings.SplitN(id, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf(
			"invalid workspace ID format: %s (expected <ENVIRONMENT_ID>/<WORKSPACE>)", id)
	}

	return s[0], s[1], nil
}
