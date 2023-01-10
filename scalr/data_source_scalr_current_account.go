package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
	"log"
)

func dataSourceScalrCurrentAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrCurrentAccountRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceScalrCurrentAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	accID, ok := getDefaultScalrAccountID()
	if !ok {
		log.Printf("[DEBUG] %s is not set", currentAccountIDEnvVar)
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
