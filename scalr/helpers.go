package scalr

import (
	"errors"
	"fmt"

	scalr "github.com/scalr/go-scalr"
)

func GetEnvironmentByName(environmentName string, scalrClient *scalr.Client) (*scalr.Environment, error) {
	var environment *scalr.Environment

	envl, err := scalrClient.Environments.List(ctx)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error retrieving environments: %v", err))
	}

	for _, env := range envl.Items {
		if env.Name == environmentName {
			environment = env
			break
		}
	}
	if environment == nil {
		return nil, errors.New(fmt.Sprintf("Could not find environment with name: %s", environmentName))
	}

	return environment, nil
}
