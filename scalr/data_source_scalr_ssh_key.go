package scalr

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrSSHKey() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves details of a specific SSH key by ID or name.",

		ReadContext: dataSourceScalrSSHKeyRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "ID of the SSH key.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},

			"name": {
				Description:  "Name of the SSH key.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"account_id": {
				Description: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},

			"environments": {
				Description: "List of environment IDs where the SSH key is available, or `[\"*\"]` if shared with all environments.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceScalrSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	sshKeyID := d.Get("id").(string)
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	var sshKey *scalr.SSHKey
	var err error

	if sshKeyID != "" {
		// Search by ID
		sshKey, err = scalrClient.SSHKeys.Read(ctx, sshKeyID)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error retrieving SSH key by ID: %v", err))
		}
	} else {
		// Search by name
		options := scalr.SSHKeysListOptions{Filter: &scalr.SSHKeyFilter{Name: name, AccountID: accountID}}
		keys, err := scalrClient.SSHKeys.List(ctx, options)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error retrieving SSH key by name: %v", err))
		}
		if len(keys.Items) == 0 {
			return diag.FromErr(fmt.Errorf("no SSH key found with name: %s", name))
		}
		if len(keys.Items) > 1 {
			return diag.FromErr(errors.New("query returned more than one SSH key; specify an ID or a more unique name"))
		}
		sshKey = keys.Items[0]
	}

	if sshKey.IsShared {
		_ = d.Set("environments", []string{"*"})
	} else {
		environmentIDs := make([]string, 0)
		for _, environment := range sshKey.Environments {
			environmentIDs = append(environmentIDs, environment.ID)
		}
		_ = d.Set("environments", environmentIDs)
	}

	_ = d.Set("name", sshKey.Name)
	_ = d.Set("account_id", sshKey.Account.ID)

	d.SetId(sshKey.ID)

	return nil
}
