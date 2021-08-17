package scalr

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform/helper/schema"
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
		SchemaVersion: 0,
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
				Optional: true,
				MinItems: 1,
				MaxItems: 128,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceScalrRoleCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get required options
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	// Create a new options struct.
	options := scalr.RoleCreateOptions{
		Name:    scalr.String(name),
		Account: &scalr.Account{ID: accountID},
	}

	// Process all optional fields.
	if value, ok := d.GetOk("permissions"); ok {
		permissionNames := value.([]interface{})
		permissions := make([]*scalr.Permission, 0)

		for _, id := range permissionNames {
			permissions = append(permissions, &scalr.Permission{ID: id.(string)})
		}
		options.Permissions = permissions
	}

	if description, ok := d.GetOk("description"); ok {
		options.Description = scalr.String(description.(string))
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

	schemaPermissions := make([]string, 0)
	if value, ok := d.GetOk("permissions"); ok {
		permissionNames := value.([]interface{})

		for _, id := range permissionNames {
			schemaPermissions = append(schemaPermissions, id.(string))
		}
	}
	sort.Strings(schemaPermissions)
	log.Printf("[DEBUG] schema permissions: %+v", schemaPermissions)

	remotePermissions := make([]string, 0)
	if len(role.Permissions) != 0 {
		for _, permission := range role.Permissions {
			remotePermissions = append(remotePermissions, permission.ID)
		}
		sort.Strings(remotePermissions)
	}
	log.Printf("[DEBUG] remote permissions: %+v", remotePermissions)

	// ignore permission ordering from the remote server
	if !reflect.DeepEqual(remotePermissions, schemaPermissions) {
		d.Set("permissions", remotePermissions)
	}

	return nil
}

func resourceScalrRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("permissions") {
		// Create a new options struct.
		options := scalr.RoleUpdateOptions{
			Name:        scalr.String(d.Get("name").(string)),
			Description: scalr.String(d.Get("description").(string)),
		}

		// Process all configured options.
		if value, ok := d.GetOk("permissions"); ok {
			permissionNames := value.([]interface{})
			permissions := make([]*scalr.Permission, 0)

			for _, id := range permissionNames {
				permissions = append(permissions, &scalr.Permission{ID: id.(string)})
			}
			options.Permissions = permissions
		}
		log.Printf("[DEBUG] Update role %s", id)
		_, err := scalrClient.Roles.Update(ctx, id, options)
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
