package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrRole() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the Scalr IAM roles. Create, update and destroy.",
		CreateContext: resourceScalrRoleCreate,
		ReadContext:   resourceScalrRoleRead,
		UpdateContext: resourceScalrRoleUpdate,
		DeleteContext: resourceScalrRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				Description: "Name of the role.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"account_id": {
				Description: "ID of the account.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Deprecated: "Attribute `account_id` is deprecated, the account id is calculated from the " +
					"API request context.",
			},

			"is_system": {
				Description: "Boolean indicates if the role can be edited. System roles are maintained by Scalr and cannot be changed.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"description": {
				Description: "Verbose description of the role.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"permissions": {
				Description: "Array of permission names.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				MaxItems:    128,
				Elem:        &schema.Schema{Type: schema.TypeString},
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

func resourceScalrRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get required options
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	// Get optional attributes
	permissions, err := parsePermissionDefinitions(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create a new options struct
	options := scalr.RoleCreateOptions{
		Name:        ptr(name),
		Description: ptr(description),
		Permissions: permissions,
	}

	log.Printf("[DEBUG] Creating role %s", name)
	role, err := scalrClient.Roles.Create(ctx, options)
	if err != nil {
		return diag.Errorf(
			"Error creating role %s: %v", name, err)
	}
	d.SetId(role.ID)
	return resourceScalrRoleRead(ctx, d, meta)
}

func resourceScalrRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()
	log.Printf("[DEBUG] Read configuration of role: %s", id)
	role, err := scalrClient.Roles.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Role %s not found", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading configuration of role %s: %v", id, err)
	}
	log.Printf("[DEBUG] role permissions: %+v", role.Permissions)

	// Update the config.
	_ = d.Set("name", role.Name)
	_ = d.Set("description", role.Description)
	_ = d.Set("account_id", role.Account.ID)
	_ = d.Set("is_system", role.IsSystem)

	uniqueSchemaPermissions := []string{}
	if value, ok := d.GetOk("permissions"); ok {
		schemaSet := make(map[string]struct{}, len(value.([]interface{})))
		for _, id := range value.([]interface{}) {
			p := id.(string)
			schemaSet[p] = struct{}{}
		}
		for p := range schemaSet {
			uniqueSchemaPermissions = append(uniqueSchemaPermissions, p)
		}
		sort.Strings(uniqueSchemaPermissions)
	}
	log.Printf("[DEBUG] unique schema permissions: %+v", uniqueSchemaPermissions)

	remotePermissions := make([]string, 0)
	if len(role.Permissions) != 0 {
		for _, permission := range role.Permissions {
			remotePermissions = append(remotePermissions, permission.ID)
		}
		sort.Strings(remotePermissions)
	}
	log.Printf("[DEBUG] remote permissions: %+v", remotePermissions)

	// ignore permission ordering from the remote server
	if !reflect.DeepEqual(remotePermissions, uniqueSchemaPermissions) {
		_ = d.Set("permissions", remotePermissions)
	}

	return nil
}

func resourceScalrRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("permissions") {
		permissions, err := parsePermissionDefinitions(d)
		if err != nil {
			return diag.FromErr(err)
		}

		// Create a new options struct
		options := scalr.RoleUpdateOptions{
			Name:        ptr(d.Get("name").(string)),
			Description: ptr(d.Get("description").(string)),
			Permissions: permissions,
		}

		log.Printf("[DEBUG] Update role %s", id)
		_, err = scalrClient.Roles.Update(ctx, id, options)
		if err != nil {
			return diag.Errorf(
				"Error updating role %s: %v", id, err)
		}
	}

	return resourceScalrRoleRead(ctx, d, meta)
}

func resourceScalrRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete role %s", id)
	err := scalrClient.Roles.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf(
			"Error deleting role %s: %v", id, err)
	}

	return nil
}
