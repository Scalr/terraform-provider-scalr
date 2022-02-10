package scalr

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrAccountAllowedIps() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrAccountAllowedIpsCreate,
		Read:   resourceScalrAccountAllowedIpsRead,
		Update: resourceScalrAccountAllowedIpsUpdate,
		Delete: resourceScalrAccountAllowedIpsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
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

func resourceScalrAccountAllowedIpsCreate(d *schema.ResourceData, meta interface{}) error {
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
		return fmt.Errorf("Error updating allowed ips for account %s: %v", accountId, err)
	}

	d.SetId(account.ID)

	return resourceScalrAccountAllowedIpsRead(d, meta)
}

func resourceScalrAccountAllowedIpsRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the ID
	accountID := d.Id()

	log.Printf("[DEBUG] Read endpoint with ID: %s", accountID)
	account, err := scalrClient.Accounts.Read(ctx, accountID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			return fmt.Errorf("Could not find account %s: %v", accountID, err)
		}
		return fmt.Errorf("Error retrieving account: %v", err)
	}

	for i, ip := range account.AllowedIPs {
		account.AllowedIPs[i] = strings.TrimSuffix(ip, "/32")
	}

	// Update the config.
	d.Set("allowed_ips", account.AllowedIPs)

	return nil
}

func resourceScalrAccountAllowedIpsUpdate(d *schema.ResourceData, meta interface{}) error {
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
		return fmt.Errorf("Error updating allowed ips for %s: %v", d.Id(), err)
	}

	return resourceScalrAccountAllowedIpsRead(d, meta)
}

func resourceScalrAccountAllowedIpsDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete allowed ips for account: %s", d.Id())

	// Create a new options struct.
	options := scalr.AccountUpdateOptions{
		ID:         d.Id(),
		AllowedIPs: &[]string{},
	}
	_, err := scalrClient.Accounts.Update(ctx, d.Id(), options)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			return nil
		}
		return fmt.Errorf("Error deleting allowed ips for account %s: %v", d.Id(), err)
	}

	return nil
}
