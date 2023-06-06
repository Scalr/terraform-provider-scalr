package scalr

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

var resourceScalrProviderConfigurationDefaultMutex sync.Mutex

func resourceScalrProviderConfigurationDefault() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalrProviderConfigurationDefaultCreate,
		ReadContext:   resourceScalrProviderConfigurationDefaultRead,
		DeleteContext: resourceScalrProviderConfigurationDefaultDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceScalrProviderConfigurationDefaultImport,
		},

		Schema: map[string]*schema.Schema{
			"provider_configuration_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalrProviderConfigurationDefaultImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	providerConfiguration, environment, err := getPCDLinkedResources(ctx, id, scalrClient)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil, fmt.Errorf("provider configuration default %s not found", id)
		}
		return nil, fmt.Errorf("error retrieving provider configuration default %s: %v", id, err)
	}

	_ = d.Set("provider_configuration_id", providerConfiguration.ID)
	_ = d.Set("environment_id", environment.ID)

	return []*schema.ResourceData{d}, nil
}

func resourceScalrProviderConfigurationDefaultCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceScalrProviderConfigurationDefaultMutex.Lock()
	defer resourceScalrProviderConfigurationDefaultMutex.Unlock()
	scalrClient := meta.(*scalr.Client)

	providerConfigurationID := d.Get("provider_configuration_id").(string)
	environmentID := d.Get("environment_id").(string)
	id := fmt.Sprintf("%s/%s", environmentID, providerConfigurationID)

	environment, err := scalrClient.Environments.Read(ctx, environmentID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("Environment %q not found", environmentID)
		}
		return diag.Errorf("error retrieving environment %s: %v", environmentID, err)
	}

	providerConfiguration, err := scalrClient.ProviderConfigurations.Read(ctx, providerConfigurationID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("Provider configuration %q not found", providerConfigurationID)
		}
		return diag.Errorf("Error retrieving provider configuration %s: %v", providerConfigurationID, err)
	}

	for _, pc := range environment.DefaultProviderConfigurations {
		if pc.ID == providerConfigurationID {
			return diag.Errorf("Provider configuration is already set as default for environment %q", environmentID)
		}
	}

	environment.DefaultProviderConfigurations = append(environment.DefaultProviderConfigurations, &scalr.ProviderConfiguration{ID: providerConfiguration.ID})
	updateOpts := scalr.EnvironmentUpdateOptions{
		DefaultProviderConfigurations: environment.DefaultProviderConfigurations,
		PolicyGroups:                  environment.PolicyGroups,
	}
	_, err = scalrClient.Environments.Update(ctx, environment.ID, updateOpts)
	if err != nil {
		return diag.Errorf("Error updating environment %s: %v", environment.ID, err)
	}
	d.SetId(id)

	return resourceScalrProviderConfigurationDefaultRead(ctx, d, meta)
}

func resourceScalrProviderConfigurationDefaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	providerConfiguration, environment, err := getPCDLinkedResources(ctx, id, scalrClient)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("provider_configuration_id", providerConfiguration.ID)
	_ = d.Set("environment_id", environment.ID)

	for _, pc := range environment.DefaultProviderConfigurations {
		if pc.ID == providerConfiguration.ID {
			return nil
		}
	}

	d.SetId("")

	return nil
}

func resourceScalrProviderConfigurationDefaultDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceScalrProviderConfigurationDefaultMutex.Lock()
	defer resourceScalrProviderConfigurationDefaultMutex.Unlock()
	scalrClient := meta.(*scalr.Client)

	providerConfigurationID := d.Get("provider_configuration_id").(string)
	environmentID := d.Get("environment_id").(string)

	environment, err := scalrClient.Environments.Read(ctx, environmentID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("Environment %q not found", environmentID)
		}
		return diag.Errorf("error retrieving environment %s: %v", environmentID, err)
	}

	found := false
	for i, pc := range environment.DefaultProviderConfigurations {
		if pc.ID == providerConfigurationID {
			environment.DefaultProviderConfigurations = append(environment.DefaultProviderConfigurations[:i], environment.DefaultProviderConfigurations[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return diag.Errorf("Provider configuration %q is not in environment %q default provider configuration", providerConfigurationID, environmentID)
	}

	updateOpts := scalr.EnvironmentUpdateOptions{
		DefaultProviderConfigurations: environment.DefaultProviderConfigurations,
		PolicyGroups:                  environment.PolicyGroups,
	}

	_, err = scalrClient.Environments.Update(ctx, environment.ID, updateOpts)
	if err != nil {
		return diag.Errorf("Error removing provider configuration %s from environment %s default provider configuration: %v", providerConfigurationID, environmentID, err)
	}

	return nil
}

func getPCDLinkedResources(ctx context.Context, id string, scalrClient *scalr.Client) (*scalr.ProviderConfiguration, *scalr.Environment, error) {
	environmentID, providerConfigurationID, err := parseProviderConfigurationDefaultID(id)
	if err != nil {
		return nil, nil, err
	}

	providerConfiguration, err := scalrClient.ProviderConfigurations.Read(ctx, providerConfigurationID)
	if err != nil {
		return nil, nil, err
	}

	environment, err := scalrClient.Environments.Read(ctx, environmentID)
	if err != nil {
		return nil, nil, err
	}

	return providerConfiguration, environment, nil
}

func parseProviderConfigurationDefaultID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid ID %q: expected {environment_id}/{provider_configuration_id}", id)
	}

	return parts[0], parts[1], nil
}
