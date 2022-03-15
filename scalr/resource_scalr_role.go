package scalr

import (
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrRoleCreate,
		Read:   resourceScalrRoleRead,
		Update: resourceScalrRoleUpdate,
		Delete: resourceScalrRoleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceScalrRoleResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrRoleStateUpgradeV0,
				Version: 0,
			},
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

			"is_system": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"permissions": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 128,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func parsePermissionDefinitions(d *schema.ResourceData) ([]*scalr.Permission, error) {
	permissions := make([]*scalr.Permission, 0)

	permissionIds := d.Get("permissions").([]interface{})
	err := ValidateIDsDefinitions(permissionIds)
	if err != nil {
		return nil, fmt.Errorf("Got error during parsing permissions: %s", err.Error())
	}

	for _, permID := range permissionIds {
		permissions = append(permissions, &scalr.Permission{ID: permID.(string)})
	}

	return permissions, nil
}

func resourceScalrRoleCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get required options
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	accountID := d.Get("account_id").(string)

	// Get optional attributes
	permissions, err := parsePermissionDefinitions(d)
	if err != nil {
		return err
	}

	// Create a new options struct
	options := scalr.RoleCreateOptions{
		Name:        scalr.String(name),
		Account:     &scalr.Account{ID: accountID},
		Description: scalr.String(description),
		Permissions: permissions,
	}

	log.Printf("[DEBUG] Create role %s for account: %s", name, accountID)
	role, err := scalrClient.Roles.Create(ctx, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating role %s for account %s: %v", name, accountID, err)
	}
	d.SetId(role.ID)
	return resourceScalrRoleRead(d, meta)
}

func resourceScalrRoleRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()
	log.Printf("[DEBUG] Read configuration of role: %s", id)
	role, err := scalrClient.Roles.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			log.Printf("[DEBUG] Role %s not found", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of role %s: %v", id, err)
	}
	log.Printf("[DEBUG] role permissions: %+v", role.Permissions)

	// Update the config.
	d.Set("name", role.Name)
	d.Set("description", role.Description)
	d.Set("account_id", role.Account.ID)
	d.Set("is_system", role.IsSystem)

	remotePermissions := make([]string, 0)
	if len(role.Permissions) != 0 {
		for _, permission := range role.Permissions {
			remotePermissions = append(remotePermissions, permission.ID)
		}
		sort.Strings(remotePermissions)
	}

	log.Printf("[DEBUG] remote permissions: %+v", remotePermissions)

	d.Set("permissions", remotePermissions)

	return nil
}

func resourceScalrRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("permissions") {
		permissions, err := parsePermissionDefinitions(d)
		if err != nil {
			return err
		}

		// Create a new options struct
		options := scalr.RoleUpdateOptions{
			Name:        scalr.String(d.Get("name").(string)),
			Description: scalr.String(d.Get("description").(string)),
			Permissions: permissions,
		}

		log.Printf("[DEBUG] Update role %s", id)
		_, err = scalrClient.Roles.Update(ctx, id, options)
		if err != nil {
			return fmt.Errorf(
				"Error updating role %s: %v", id, err)
		}
	}

	return resourceScalrRoleRead(d, meta)
}

func resourceScalrRoleDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete role %s", id)
	err := scalrClient.Roles.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			return nil
		}
		return fmt.Errorf(
			"Error deleting role %s: %v", id, err)
	}

	return nil
}
