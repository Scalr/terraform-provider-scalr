package scalr

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func resourceTFESSHKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFESSHKeyCreate,
		Read:   resourceTFESSHKeyRead,
		Update: resourceTFESSHKeyUpdate,
		Delete: resourceTFESSHKeyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceTFESSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create a new options struct.
	options := scalr.SSHKeyCreateOptions{
		Name:  scalr.String(name),
		Value: scalr.String(d.Get("key").(string)),
	}

	log.Printf("[DEBUG] Create new SSH key for organization: %s", organization)
	sshKey, err := scalrClient.SSHKeys.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating SSH key %s for organization %s: %v", name, organization, err)
	}

	d.SetId(sshKey.ID)

	return resourceTFESSHKeyUpdate(d, meta)
}

func resourceTFESSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Read configuration of SSH key: %s", d.Id())
	sshKey, err := scalrClient.SSHKeys.Read(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			log.Printf("[DEBUG] SSH key %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of SSH key %s: %v", d.Id(), err)
	}

	// Update the config.
	d.Set("name", sshKey.Name)

	return nil
}

func resourceTFESSHKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Create a new options struct.
	options := scalr.SSHKeyUpdateOptions{
		Name:  scalr.String(d.Get("name").(string)),
		Value: scalr.String(d.Get("key").(string)),
	}

	log.Printf("[DEBUG] Update SSH key: %s", d.Id())
	_, err := scalrClient.SSHKeys.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating SSH key %s: %v", d.Id(), err)
	}

	return resourceTFESSHKeyRead(d, meta)
}

func resourceTFESSHKeyDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete SSH key: %s", d.Id())
	err := scalrClient.SSHKeys.Delete(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting SSH key %s: %v", d.Id(), err)
	}

	return nil
}
