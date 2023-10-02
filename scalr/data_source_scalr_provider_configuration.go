package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/scalr/go-scalr"
)

func dataSourceScalrProviderConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a single provider configuration.",
		ReadContext: dataSourceScalrProviderConfigurationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "The provider configuration ID, in the format `pcfg-xxxxxxxxxxx`.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"account_id": {
				Description: "The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"name": {
				Description:  "The name of a Scalr provider configuration.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"provider_name": {
				Description: "The name of a Terraform provider.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"environments": {
				Description: "The list of environment identifiers that the provider configuration is shared to, or `[\"*\"]` if shared with all environments.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
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
		ProviderConfiguration: providerID,
		AccountID:             accountID,
		Name:                  name,
		ProviderName:          providerName,
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

	if providerConfiguration.IsShared {
		_ = d.Set("environments", []string{"*"})
	} else {
		environments := make([]string, 0)
		for _, environment := range providerConfiguration.Environments {
			environments = append(environments, environment.ID)
		}
		_ = d.Set("environments", environments)
	}

	d.SetId(providerConfiguration.ID)

	return nil
}
