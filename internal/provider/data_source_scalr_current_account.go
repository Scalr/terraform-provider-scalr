package provider

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework/defaults"
)

func dataSourceScalrCurrentAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of current account when using Scalr remote backend." +
			"\n\nNo arguments are required. The data source returns details of the current account based on the" +
			" `SCALR_ACCOUNT_ID` environment variable that is automatically exported in the Scalr remote backend.",
		ReadContext: dataSourceScalrCurrentAccountRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The identifier of the account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceScalrCurrentAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	accID, ok := getDefaultScalrAccountID()
	if !ok {
		log.Printf("[DEBUG] %s is not set", defaults.CurrentAccountIDEnvVar)
		return diag.Errorf("Current account is not set")
	}

	log.Printf("[DEBUG] Read configuration of account: %s", accID)
	acc, err := scalrClient.Accounts.Read(ctx, accID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("Could not find account %s", accID)
		}
		return diag.Errorf("Error retrieving account: %v", err)
	}

	_ = d.Set("name", acc.Name)
	d.SetId(accID)

	return nil
}
