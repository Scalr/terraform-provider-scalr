package provider

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrIamTeam() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages the Scalr IAM teams: performs create, update and destroy actions.",
		CreateContext: resourceScalrIamTeamCreate,
		ReadContext:   resourceScalrIamTeamRead,
		UpdateContext: resourceScalrIamTeamUpdate,
		DeleteContext: resourceScalrIamTeamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "A name of the team.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "A verbose description of the team.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"account_id": {
				Description: "An identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"identity_provider_id": {
				Description: "An identifier of the login identity provider, in the format `idp-<RANDOM STRING>`. This is required when `account_id` is not specified.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"users": {
				Description: "A list of the user identifiers to add to the team.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func parseUserDefinitions(d *schema.ResourceData) ([]*scalr.User, error) {
	var users []*scalr.User

	userIDs := d.Get("users").([]interface{})
	err := ValidateIDsDefinitions(userIDs)
	if err != nil {
		return nil, fmt.Errorf("Got error during parsing users: %s", err.Error())
	}

	for _, userID := range userIDs {
		users = append(users, &scalr.User{ID: userID.(string)})
	}

	return users, nil
}

func resourceScalrIamTeamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	users, err := parseUserDefinitions(d)
	if err != nil {
		return diag.FromErr(err)
	}

	opts := scalr.TeamCreateOptions{
		Name:    &name,
		Account: &scalr.Account{ID: accountID},
		Users:   users,
	}

	// Optional attributes
	if desc, ok := d.GetOk("description"); ok {
		opts.Description = scalr.String(desc.(string))
	}
	if idpID, ok := d.GetOk("identity_provider_id"); ok {
		opts.IdentityProvider = &scalr.IdentityProvider{ID: idpID.(string)}
	}

	t, err := scalrClient.Teams.Create(ctx, opts)
	if err != nil {
		return diag.Errorf("error creating team: %v", err)
	}

	d.SetId(t.ID)
	return resourceScalrIamTeamRead(ctx, d, meta)
}

func resourceScalrIamTeamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()
	log.Printf("[DEBUG] Read configuration of team %s", id)
	t, err := scalrClient.Teams.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Team %s not found", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading configuration of team %s: %v", id, err)
	}

	// Update the configuration.
	_ = d.Set("name", t.Name)
	_ = d.Set("description", t.Description)
	_ = d.Set("identity_provider_id", t.IdentityProvider.ID)
	if t.Account != nil {
		_ = d.Set("account_id", t.Account.ID)
	}

	var users []string
	if len(t.Users) != 0 {
		for _, u := range t.Users {
			users = append(users, u.ID)
		}
	}
	_ = d.Set("users", users)

	return nil
}

func resourceScalrIamTeamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("users") {

		name := d.Get("name").(string)
		desc := d.Get("description").(string)
		users, err := parseUserDefinitions(d)
		if err != nil {
			return diag.FromErr(err)
		}

		opts := scalr.TeamUpdateOptions{
			Name:        scalr.String(name),
			Description: scalr.String(desc),
			Users:       users,
		}

		log.Printf("[DEBUG] Update team %s", id)
		_, err = scalrClient.Teams.Update(ctx, id, opts)
		if err != nil {
			return diag.Errorf("error updating team %s: %v", id, err)
		}
	}

	return resourceScalrIamTeamRead(ctx, d, meta)
}

func resourceScalrIamTeamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete team %s", id)
	err := scalrClient.Teams.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Team %s not found", id)
			return nil
		}
		return diag.Errorf("error deleting team %s: %v", id, err)
	}

	return nil
}
