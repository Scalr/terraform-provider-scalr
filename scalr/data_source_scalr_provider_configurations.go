package scalr

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/scalr/go-scalr"
)

func dataSourceScalrProviderConfigurations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrProviderConfigurationsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
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

func dataSourceScalrProviderConfigurationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	accountID := d.Get("account_id").(string)
	name := d.Get("name").(string)
	providerName := d.Get("provider_name").(string)

	providersFilter := scalr.ProviderConfigurationFilter{
		AccountID:    accountID,
		Name:         name,
		ProviderName: providerName,
	}
	options := scalr.ProviderConfigurationsListOptions{
		Filter: &providersFilter,
	}

	var ids []string

	for {
		providerConfigurations, err := scalrClient.ProviderConfigurations.List(ctx, options)
		if err != nil {
			return diag.Errorf("Error retrieving provider configuration: %v", err)
		}

		for _, providerConfiguration := range providerConfigurations.Items {
			ids = append(ids, providerConfiguration.ID)
		}

		// Exit the loop when we've seen all pages.
		if providerConfigurations.CurrentPage >= providerConfigurations.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = providerConfigurations.NextPage
	}

	_ = d.Set("ids", ids)
	d.SetId(fmt.Sprintf("%d", schema.HashString(accountID+name+providerName)))

	return nil
}
