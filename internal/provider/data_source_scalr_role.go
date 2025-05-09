package provider

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrRole() *schema.Resource {
	return &schema.Resource{
		Description: "This data source is used to retrieve details of a single role.",
		ReadContext: dataSourceScalrRoleRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "ID of the role.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},
			"name": {
				Description:  "Name of the role.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"account_id": {
				Description: "ID of the account.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDOptionalDefaultFunc,
			},

			"is_system": {
				Description: "Boolean indicates if the role can be edited.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"description": {
				Description: "Verbose description of the role.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"permissions": {
				Description: "Array of permission names.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceScalrRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// required fields
	roleID := d.Get("id").(string)
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	options := scalr.RoleListOptions{
		Account: ptr("in:null," + accountID),
	}

	if roleID != "" {
		options.Role = roleID
	}

	if name != "" {
		options.Name = name
	}

	log.Printf("[DEBUG] Read configuration of role with ID '%s', name '%s', and account_id '%s'", roleID, name, accountID)
	roles, err := scalrClient.Roles.List(ctx, options)
	if err != nil {
		return diag.Errorf("Error retrieving role: %v", err)
	}

	// Unlikely
	if roles.TotalCount > 1 {
		return diag.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if roles.TotalCount == 0 {
		return diag.Errorf("Could not find role with ID '%s', name '%s', and account_id '%s'", roleID, name, accountID)
	}

	role := roles.Items[0]

	// Update the config.
	_ = d.Set("name", role.Name)
	_ = d.Set("is_system", role.IsSystem)
	_ = d.Set("description", role.Description)
	d.SetId(role.ID)

	if len(role.Permissions) != 0 {
		permissionNames := make([]string, 0)

		for _, permission := range role.Permissions {
			permissionNames = append(permissionNames, permission.ID)
		}
		_ = d.Set("permissions", permissionNames)
	}

	if role.Account == nil {
		_ = d.Set("account_id", nil)
	}

	return nil
}
