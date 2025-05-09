package provider

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework/defaults"
)

const (
	dummyIdentifier = "-"
)

type GetEnvironmentByNameOptions struct {
	Name    *string
	Account *string
	Include *string
}

func GetEnvironmentByName(ctx context.Context, options GetEnvironmentByNameOptions, scalrClient *scalr.Client) (*scalr.Environment, error) {
	listOptions := scalr.EnvironmentListOptions{
		Include: options.Include,
		Filter: &scalr.EnvironmentFilter{
			Name:    options.Name,
			Account: options.Account,
		},
	}
	envl, err := scalrClient.Environments.List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving environments: %v", err)
	}

	if len(envl.Items) == 0 {
		return nil, fmt.Errorf("Environment with name '%s' not found or user unauthorized", *options.Name)
	}
	if len(envl.Items) > 1 {
		return nil, fmt.Errorf("Found more than one environment with name: %s, specify 'account_id' to search only for environments in specific account", *options.Name)
	}

	return envl.Items[0], nil
}

type GetWebhookByNameOptions struct {
	Name    *string
	Account *string
}

func GetWebhookByName(ctx context.Context, options GetWebhookByNameOptions, scalrClient *scalr.Client) (*scalr.WebhookIntegration, error) {
	listOptions := scalr.WebhookIntegrationListOptions{
		Query:   options.Name,
		Account: options.Account,
	}
	whl, err := scalrClient.WebhookIntegrations.List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving webhooks: %v", err)
	}

	if len(whl.Items) == 0 {
		return nil, fmt.Errorf("Webhook with name '%s' not found or user unauthorized", *options.Name)
	}

	var matchedWebhooks []*scalr.WebhookIntegration

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
		return nil, fmt.Errorf("Found more than one webhook with name: %s", *options.Name)

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

func InterfaceArrToTagRelationArr(arr []interface{}) []*scalr.TagRelation {
	tags := make([]*scalr.TagRelation, len(arr))
	for i, id := range arr {
		tags[i] = &scalr.TagRelation{ID: id.(string)}
	}
	return tags
}

func getDefaultScalrAccountID() (string, bool) {
	if v := os.Getenv(defaults.CurrentAccountIDEnvVar); v != "" {
		return v, true
	}
	return "", false
}

// scalrAccountIDDefaultFunc is a schema.SchemaDefaultFunc that returns default account id.
// If account info is not present, the error is returned.
func scalrAccountIDDefaultFunc() (interface{}, error) {
	if accID, ok := getDefaultScalrAccountID(); ok {
		return accID, nil
	}
	return nil, errors.New("Default value for `account_id` could not be computed." +
		"\nIf you are using Scalr Provider for local runs, please set the attribute in resources explicitly," +
		"\nor export `SCALR_ACCOUNT_ID` environment variable prior the run.")
}

// scalrAccountIDOptionalDefaultFunc is a schema.SchemaDefaultFunc that returns default account id
// or an empty (string) value, if account info is not present.
// Never returns non-nil error.
func scalrAccountIDOptionalDefaultFunc() (interface{}, error) {
	accID, _ := getDefaultScalrAccountID()
	return accID, nil
}

// ptr returns a pointer to the value passed.
func ptr[T any](v T) *T {
	return &v
}

// diff returns the added and removed elements between two slices.
func diff[T comparable](old, new []T) (added, removed []T) {
	newSet := make(map[T]struct{}, len(new))
	for _, v := range new {
		newSet[v] = struct{}{}
	}

	oldSet := make(map[T]struct{}, len(old))
	for _, v := range old {
		oldSet[v] = struct{}{}
		if _, found := newSet[v]; !found {
			removed = append(removed, v)
		}
	}

	for _, v := range new {
		if _, found := oldSet[v]; !found {
			added = append(added, v)
		}
	}

	return
}

// setsEqual reports whether two slices contain the same set of elements.
// The order of elements does not matter, and each element must be unique in the slice.
func setsEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	set := make(map[T]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}

	for _, v := range b {
		if _, ok := set[v]; !ok {
			return false
		}
	}

	return true
}
