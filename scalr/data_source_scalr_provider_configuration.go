package scalr

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/scalr/go-scalr"
)

func dataSourceScalrProviderConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrProviderConfigurationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceScalrProviderConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	providerID := d.Get("id").(string)
	accountID := d.Get("account_id").(string)
	name := d.Get("name").(string)
	providerName := d.Get("provider_name").(string)

	providersFilter := scalr.ProviderConfigurationFilter{
		ProviderID:   providerID,
		AccountID:    accountID,
		Name:         name,
		ProviderName: providerName,
	}
	options := scalr.ProviderConfigurationsListOptions{
		Filter: &providersFilter,
	}

	providerConfigurations, err := scalrClient.ProviderConfigurations.List(ctx, options)
	if err != nil {
		return diag.Errorf("Error retrieving provider configuration: %v", err)
	}

	if len(providerConfigurations.Items) > 1 {
		return diag.FromErr(errors.New("Your query returned more than one result. Please try a more specific search criteria."))
	}
	if len(providerConfigurations.Items) == 0 {
		return diag.Errorf("Could not find provider configuration with ID '%s', name '%s', account_id '%s', and provider_name '%s'", providerID, name, accountID, providerName)
	}

	providerConfiguration := providerConfigurations.Items[0]
	d.SetId(providerConfiguration.ID)

	return nil
}
