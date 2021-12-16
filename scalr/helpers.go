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

type GetEnvironmentByNameOptions struct {
	Name    *string
	Account *string `json:",omitempty"`
	Include *string `json:",omitempty"`
}

func GetEnvironmentByName(options GetEnvironmentByNameOptions, scalrClient *scalr.Client) (*scalr.Environment, error) {
	listOptions := scalr.EnvironmentListOptions{
		Name:    options.Name,
		Account: options.Account,
		Include: options.Include,
	}
	envl, err := scalrClient.Environments.List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving environments: %v", err)
	}

	if len(envl.Items) == 0 {
		return nil, fmt.Errorf("Environment with name '%s' not found or user unauthorized", *options.Name)
	}

	var matchedEnvironments []*scalr.Environment

	// filter in endpoint search environments that contains quering string, this is why we need to do exeact match on our side.
	for _, env := range envl.Items {
		if env.Name == *options.Name {
			matchedEnvironments = append(matchedEnvironments, env)
		}
	}

	switch numberOfMatch := len(matchedEnvironments); {
	case numberOfMatch == 0:
		return nil, fmt.Errorf("Environment with name '%s' not found", *options.Name)

	case numberOfMatch > 1:
		return nil, fmt.Errorf("Find more than one environment with name: %v, specify `account_id` to search only for environments in specific account", options.Name)

	default:
		return matchedEnvironments[0], nil

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
