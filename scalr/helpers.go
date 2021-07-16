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

func GetEnvironmentByName(environmentName string, scalrClient *scalr.Client) (*scalr.Environment, error) {
	var environment *scalr.Environment

	envl, err := scalrClient.Environments.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving environments: %v", err)
	}

	for _, env := range envl.Items {
		if env.Name == environmentName {
			environment = env
			break
		}
	}
	if environment == nil {
		return nil, fmt.Errorf("Could not find environment with name: %s", environmentName)
	}

	return environment, nil
}

func GetRandomInteger() int {
	return rand.Int()
}
