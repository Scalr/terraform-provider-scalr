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
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"email"},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
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

	saID := d.Get("id").(string)
	email := d.Get("email").(string)
	accountID := d.Get("account_id").(string)

	options := scalr.ServiceAccountListOptions{
		Account: scalr.String(accountID),
		Include: scalr.String("created-by"),
	}

	if saID != "" {
		options.ServiceAccount = scalr.String(saID)
	}

	if email != "" {
		options.Email = scalr.String(email)
	}

	log.Printf("[DEBUG] Read service account with ID '%s', email '%s', and account_id '%s'", saID, email, accountID)
	sas, err := scalrClient.ServiceAccounts.List(ctx, options)
	if err != nil {
		return diag.Errorf("Error retrieving service account: %v", err)
	}

	// Unlikely
	if sas.TotalCount > 1 {
		return diag.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if sas.TotalCount == 0 {
		return diag.Errorf("Could not find service account with ID '%s', email '%s', and account_id '%s'", saID, email, accountID)
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
	_ = d.Set("name", sa.Name)
	_ = d.Set("email", sa.Email)
	_ = d.Set("description", sa.Description)
	_ = d.Set("status", sa.Status)
	_ = d.Set("created_by", createdBy)

	d.SetId(sa.ID)

	return nil
}
