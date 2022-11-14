package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrIamTeam() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrIamTeamCreate,
		Read:   resourceScalrIamTeamRead,
		Update: resourceScalrIamTeamUpdate,
		Delete: resourceScalrIamTeamDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"identity_provider_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"users": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceScalrIamTeamCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	name := d.Get("name").(string)

	users, err := parseUserDefinitions(d)
	if err != nil {
		return err
	}

	opts := scalr.TeamCreateOptions{
		Name:  scalr.String(name),
		Users: users,
	}

	// Optional attributes
	if desc, ok := d.GetOk("description"); ok {
		opts.Description = scalr.String(desc.(string))
	}
	if accID, ok := d.GetOk("account_id"); ok {
		opts.Account = &scalr.Account{ID: accID.(string)}
	}
	if idpID, ok := d.GetOk("identity_provider_id"); ok {
		opts.IdentityProvider = &scalr.IdentityProvider{ID: idpID.(string)}
	}

	t, err := scalrClient.Teams.Create(ctx, opts)
	if err != nil {
		return fmt.Errorf("error creating team: %v", err)
	}

	d.SetId(t.ID)
	return resourceScalrIamTeamRead(d, meta)
}

func resourceScalrIamTeamRead(d *schema.ResourceData, meta interface{}) error {
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
		return fmt.Errorf("error reading configuration of team %s: %v", id, err)
	}

	// Update the configuration.
	d.Set("name", t.Name)
	d.Set("description", t.Description)
	d.Set("identity_provider_id", t.IdentityProvider.ID)
	if t.Account != nil {
		d.Set("account_id", t.Account.ID)
	}

	var users []string
	if len(t.Users) != 0 {
		for _, u := range t.Users {
			users = append(users, u.ID)
		}
	}
	d.Set("users", users)

	return nil
}

func resourceScalrIamTeamUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("users") {

		name := d.Get("name").(string)
		desc := d.Get("description").(string)
		users, err := parseUserDefinitions(d)
		if err != nil {
			return err
		}

		opts := scalr.TeamUpdateOptions{
			Name:        scalr.String(name),
			Description: scalr.String(desc),
			Users:       users,
		}

		log.Printf("[DEBUG] Update team %s", id)
		_, err = scalrClient.Teams.Update(ctx, id, opts)
		if err != nil {
			return fmt.Errorf("error updating team %s: %v", id, err)
		}
	}

	return resourceScalrIamTeamRead(d, meta)
}

func resourceScalrIamTeamDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete team %s", id)
	err := scalrClient.Teams.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Team %s not found", id)
			return nil
		}
		return fmt.Errorf("error deleting team %s: %v", id, err)
	}

	return nil
}
