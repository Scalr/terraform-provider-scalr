package scalr

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrAccountAllowedIps() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalrAccountAllowedIpsCreate,
		ReadContext:   resourceScalrAccountAllowedIpsRead,
		UpdateContext: resourceScalrAccountAllowedIpsUpdate,
		DeleteContext: resourceScalrAccountAllowedIpsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},

			"allowed_ips": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
				Required: true,
			},
		},
	}
}

func preprocessAllowedIps(allowedIps []interface{}) []string {
	ips := make([]string, 0)
	for _, v := range allowedIps {
		ips = append(ips, v.(string))
	}

	return ips
}

func resourceScalrAccountAllowedIpsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get attributes.
	accountId := d.Get("account_id").(string)

	allowedIps := preprocessAllowedIps(d.Get("allowed_ips").([]interface{}))
	// Create a new options struct.
	options := scalr.AccountUpdateOptions{
		ID:         accountId,
		AllowedIPs: &allowedIps,
	}

	log.Printf("[DEBUG] Update allowed ips: %s", accountId)
	account, err := scalrClient.Accounts.Update(ctx, accountId, options)
	if err != nil {
		return diag.Errorf("Error updating allowed ips for account %s: %v", accountId, err)
	}

	d.SetId(account.ID)

	return resourceScalrAccountAllowedIpsRead(ctx, d, meta)
}

func resourceScalrAccountAllowedIpsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get the ID
	accountID := d.Id()

	log.Printf("[DEBUG] Read endpoint with ID: %s", accountID)
	account, err := scalrClient.Accounts.Read(ctx, accountID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("Could not find account %s: %v", accountID, err)
		}
		return diag.Errorf("Error retrieving account: %v", err)
	}

	allowedIpsAsString := fmt.Sprintf("%v", d.Get("allowed_ips").([]interface{}))
	for i, ip := range account.AllowedIPs {
		if !strings.Contains(allowedIpsAsString, ip) {
			ip = strings.TrimSuffix(ip, "/32")
		}
		account.AllowedIPs[i] = ip
	}

	// Update the config.
	_ = d.Set("allowed_ips", account.AllowedIPs)
	_ = d.Set("account_id", accountID)

	return nil
}

func resourceScalrAccountAllowedIpsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get attributes.
	allowedIps := preprocessAllowedIps(d.Get("allowed_ips").([]interface{}))

	// Create a new options struct.
	options := scalr.AccountUpdateOptions{
		ID:         d.Id(),
		AllowedIPs: &allowedIps,
	}

	log.Printf("[DEBUG] Update allowed ips for account: %s", d.Id())
	_, err := scalrClient.Accounts.Update(ctx, d.Id(), options)
	if err != nil {
		return diag.Errorf("Error updating allowed ips for %s: %v", d.Id(), err)
	}

	return resourceScalrAccountAllowedIpsRead(ctx, d, meta)
}

func resourceScalrAccountAllowedIpsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete allowed ips for account: %s", d.Id())

	// Create a new options struct.
	options := scalr.AccountUpdateOptions{
		ID:         d.Id(),
		AllowedIPs: &[]string{},
	}
	_, err := scalrClient.Accounts.Update(ctx, d.Id(), options)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting allowed ips for account %s: %v", d.Id(), err)
	}

	return nil
}
