package scalr

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

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
	scalrClient := meta.(*scalr.Client)

	providerConfigurationID := d.Get("provider_configuration_id").(string)
	environmentID := d.Get("environment_id").(string)
	id := fmt.Sprintf("%s/%s", environmentID, providerConfigurationID)

	opts := scalr.ProviderConfigurationDefaultsCreateOptions{
		EnvironmentID:           environmentID,
		ProviderConfigurationID: providerConfigurationID,
	}
	err := scalrClient.ProviderConfigurationDefaults.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(err)
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

	return nil
}

func resourceScalrProviderConfigurationDefaultDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	providerConfigurationID := d.Get("provider_configuration_id").(string)
	environmentID := d.Get("environment_id").(string)

	opts := scalr.ProviderConfigurationDefaultsDeleteOptions{
		EnvironmentID:           environmentID,
		ProviderConfigurationID: providerConfigurationID,
	}

	err := scalrClient.ProviderConfigurationDefaults.Delete(ctx, opts)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getPCDLinkedResources(ctx context.Context, id string, scalrClient *scalr.Client) (*scalr.ProviderConfiguration, *scalr.Environment, error) {
	environmentID, providerConfigurationID, err := parseID(id)
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

func parseID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid ID %q: expected {environment_id}/{provider_configuration_id}", id)
	}

	return parts[0], parts[1], nil
}
