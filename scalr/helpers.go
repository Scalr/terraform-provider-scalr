package scalr

import (
	"fmt"
	"math/rand"
	"time"

	scalr "github.com/scalr/go-scalr"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetEnvironmentByName(options scalr.EnvironmentListOptions, scalrClient *scalr.Client) (*scalr.Environment, error) {
	envl, err := scalrClient.Environments.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving environments: %v", err)
	}

	switch numberOfEnvironments := len(envl.Items); {
	case numberOfEnvironments == 0:
		return nil, fmt.Errorf("Environment with name '%s' not found or user unauthorized", *options.Name)
	case numberOfEnvironments > 1:
		// todo: update the error message.
		return nil, fmt.Errorf("Find more than one environment with name: %v, specify account id to be more specific", options.Name)
	default:
		return envl.Items[0], nil
	}
}

func GetRandomInteger() int {
	return rand.Int()
}

func ValidateIDsDefinitions(d []interface{}) error {
	for i, id := range d {
		id, ok := id.(string)
		if !ok || id == "" {
			return fmt.Errorf("%d-th value is empty", i)
		}
	}
	return nil
}
