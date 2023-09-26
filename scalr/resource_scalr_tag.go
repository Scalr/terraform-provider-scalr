package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
	"log"
)

func resourceScalrTag() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages the state of tags in Scalr.",
		CreateContext: resourceScalrTagCreate,
		ReadContext:   resourceScalrTagRead,
		UpdateContext: resourceScalrTagUpdate,
		DeleteContext: resourceScalrTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the tag.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"account_id": {
				Description: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
		},
	}
}

func resourceScalrTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("Error reading tag %s: %v", id, err)
	}

	// Update config.
	_ = d.Set("name", tag.Name)
	_ = d.Set("account_id", tag.Account.ID)

	return nil
}

func resourceScalrTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf(
			"Error creating tag %s for account %s: %v", name, accountID, err)
	}
	d.SetId(tag.ID)

	return resourceScalrTagRead(ctx, d, meta)
}

func resourceScalrTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
			return diag.Errorf("error updating tag %s: %v", id, err)
		}
	}

	return resourceScalrTagRead(ctx, d, meta)
}

func resourceScalrTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete tag %s", id)
	err := scalrClient.Tags.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting tag %s: %v", id, err)
	}

	return nil
}
