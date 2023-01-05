package scalr

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
	"log"
)

func dataSourceScalrServiceAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrServiceAccountRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"created_by": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"full_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceScalrServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	email := d.Get("email").(string)
	accountID := d.Get("account_id").(string)

	options := scalr.ServiceAccountListOptions{
		Email:   scalr.String(email),
		Account: scalr.String(accountID),
		Include: scalr.String("created-by"),
	}

	log.Printf("[DEBUG] Read service account: %s/%s", accountID, email)
	sas, err := scalrClient.ServiceAccounts.List(ctx, options)
	if err != nil {
		return diag.Errorf("Error retrieving service account: %v", err)
	}

	// Unlikely
	if sas.TotalCount > 1 {
		return diag.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if sas.TotalCount == 0 {
		return diag.Errorf("Could not find service account %s/%s", accountID, email)
	}

	sa := sas.Items[0]

	var createdBy []interface{}
	if sa.CreatedBy != nil {
		createdBy = append(createdBy, map[string]interface{}{
			"username":  sa.CreatedBy.Username,
			"email":     sa.CreatedBy.Email,
			"full_name": sa.CreatedBy.FullName,
		})
	}
	_ = d.Set("created_by", createdBy)

	d.SetId(sa.ID)

	return nil
}
