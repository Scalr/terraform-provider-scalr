package scalr

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrProviderConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrProviderConfigurationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"provider_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceScalrProviderConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	accountID := d.Get("account_id").(string)
	name := d.Get("name").(string)
	providerType := d.Get("provider_type").(string)

	providersFilter := scalr.ProviderConfigurationFilter{
		AccountID:    accountID,
		Name:         name,
		ProviderType: providerType,
	}
	options := scalr.ProviderConfigurationsListOptions{
		Filter: &providersFilter,
	}

	providerConfigurations, err := scalrClient.ProviderConfigurations.List(ctx, options)
	if err != nil {
		return fmt.Errorf("Error retrieving provider configuration: %v", err)
	}

	if len(providerConfigurations.Items) > 1 {
		return errors.New("Your query returned more than one result. Please try a more specific search criteria.")
	}
	if len(providerConfigurations.Items) == 0 {
		return fmt.Errorf("Could not find provider configuration with name '%s', account_id: '%s', and provider_type: '%s'", name, accountID, providerType)
	}

	providerConfiguration := providerConfigurations.Items[0]
	d.SetId(providerConfiguration.ID)

	return nil
}
