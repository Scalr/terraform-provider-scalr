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
	Account *string
	Include *string
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
		return nil, fmt.Errorf("Found more than one environment with name: %s, specify 'account_id' to search only for environments in specific account", *options.Name)

	default:
		return matchedEnvironments[0], nil

	}
}

type GetEndpointByNameOptions struct {
	Name    *string
	Account *string
}

func GetEndpointByName(options GetEndpointByNameOptions, scalrClient *scalr.Client) (*scalr.Endpoint, error) {
	listOptions := scalr.EndpointListOptions{
		Name:    options.Name,
		Account: options.Account,
	}
	endpl, err := scalrClient.Endpoints.List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving endpoints: %v", err)
	}

	if len(endpl.Items) == 0 {
		return nil, fmt.Errorf("Endpoint with name '%s' not found or user unauthorized", *options.Name)
	}

	var matchedEndpoints []*scalr.Endpoint

	// filter in endpoint search endpoints that contains query string, this is why we need to do exact match on our side.
	for _, endp := range endpl.Items {
		if endp.Name == *options.Name {
			matchedEndpoints = append(matchedEndpoints, endp)
		}
	}

	switch numberOfMatch := len(matchedEndpoints); {
	case numberOfMatch == 0:
		return nil, fmt.Errorf("Endpoint with name '%s' not found", *options.Name)

	case numberOfMatch > 1:
		return nil, fmt.Errorf("Found more than one endpoint with name: %s, specify 'account_id' to search only for endpoints in specific account", *options.Name)

	default:
		return matchedEndpoints[0], nil

	}
}

type GetWebhookByNameOptions struct {
	Name        *string
	Environment *string
}

func GetWebhookByName(options GetWebhookByNameOptions, scalrClient *scalr.Client) (*scalr.Webhook, error) {
	listOptions := scalr.WebhookListOptions{
		Name:        options.Name,
		Environment: options.Environment,
	}
	whl, err := scalrClient.Webhooks.List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving webhooks: %v", err)
	}

	if len(whl.Items) == 0 {
		return nil, fmt.Errorf("Webhook with name '%s' not found or user unauthorized", *options.Name)
	}

	var matchedWebhooks []*scalr.Webhook

	// filter in endpoint search endpoints that contains query string, this is why we need to do exact match on our side.
	for _, wh := range whl.Items {
		if wh.Name == *options.Name {
			matchedWebhooks = append(matchedWebhooks, wh)
		}
	}

	switch numberOfMatch := len(matchedWebhooks); {
	case numberOfMatch == 0:
		return nil, fmt.Errorf("Webhook with name '%s' not found", *options.Name)

	case numberOfMatch > 1:
		return nil, fmt.Errorf("Found more than one webhook with name: %s, specify 'environment_id' to search only for webhooks in specific environment", *options.Name)

	default:
		return matchedWebhooks[0], nil

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
