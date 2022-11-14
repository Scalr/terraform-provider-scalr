package scalr

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
	"log"
)

func resourceScalrTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrTagCreate,
		Read:   resourceScalrTagRead,
		Update: resourceScalrTagUpdate,
		Delete: resourceScalrTagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalrTagRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Read tag: %s", id)
	tag, err := scalrClient.Tags.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Tag %s not found", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading tag %s: %v", id, err)
	}

	// Update config.
	d.Set("name", tag.Name)
	d.Set("account_id", tag.Account.ID)

	return nil
}

func resourceScalrTagCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the name and account_id.
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	options := scalr.TagCreateOptions{
		Name:    scalr.String(name),
		Account: &scalr.Account{ID: accountID},
	}

	log.Printf("[DEBUG] Create tag %s for account %s", name, accountID)
	tag, err := scalrClient.Tags.Create(ctx, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating tag %s for account %s: %v", name, accountID, err)
	}
	d.SetId(tag.ID)

	return resourceScalrTagRead(d, meta)
}

func resourceScalrTagUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()
	if d.HasChange("name") {
		name := d.Get("name").(string)
		opts := scalr.TagUpdateOptions{
			Name: scalr.String(name),
		}
		log.Printf("[DEBUG] Update tag %s", id)
		_, err := scalrClient.Tags.Update(ctx, id, opts)
		if err != nil {
			return fmt.Errorf("error updating tag %s: %v", id, err)
		}
	}

	return resourceScalrTagRead(d, meta)
}

func resourceScalrTagDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete tag %s", id)
	err := scalrClient.Tags.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting tag %s: %v", id, err)
	}

	return nil
}
