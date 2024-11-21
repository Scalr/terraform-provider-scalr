package provider

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
)

func resourceScalrVariable() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the state of the variables in Scalr. Create, update and destroy.",
		CreateContext: resourceScalrVariableCreate,
		ReadContext:   resourceScalrVariableRead,
		UpdateContext: resourceScalrVariableUpdate,
		DeleteContext: resourceScalrVariableDelete,
		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIf(
				"key",
				func(ctx context.Context, d *schema.ResourceDiff, meta any) bool {
					// Force new when updating the `key` value of a sensitive variable.
					// To do this we check the `sensitive` value before the change,
					// as it might be changed in new configuration as well.
					oldSens, _ := d.GetChange("sensitive")
					return oldSens.(bool)
				},
			),
			customdiff.ForceNewIfChange(
				"sensitive",
				func(ctx context.Context, old, new, meta any) bool {
					// Force new when updating the `sensitive` value from true to false.
					return old.(bool)
				},
			),
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 3,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceScalrVariableResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrVariableStateUpgradeV0,
				Version: 0,
			},
			{
				Type:    resourceScalrVariableResourceV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrVariableStateUpgradeV1,
				Version: 1,
			},

			{
				Type:    resourceScalrVariableResourceV2().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrVariableStateUpgradeV2,
				Version: 2,
			},
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Key of the variable.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"value": {
				Description: "Variable value.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Sensitive:   true,
			},

			"category": {
				Description: "Indicates if this is a Terraform or shell variable. Allowed values are `terraform` or `shell`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.CategoryEnv),
						string(scalr.CategoryTerraform),
						string(scalr.CategoryShell),
					},
					false,
				),
			},

			"hcl": {
				Description: "Set (true/false) to configure the variable as a string of HCL code. Has no effect for `category = \"shell\"` variables. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"sensitive": {
				Description: "Set (true/false) to configure as sensitive. Sensitive variable values are not visible after being set. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"description": {
				Description: "Variable verbose description, defaults to empty string.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"final": {
				Description: "Set (true/false) to configure as final. Indicates whether the variable can be overridden on a lower scope down the Scalr organizational model. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"force": {
				Description: "Set (true/false) to configure as force. Allows creating final variables on higher scope, even if the same variable exists on lower scope (lower is to be deleted). Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"workspace_id": {
				Description: "The workspace that owns the variable, specified as an ID, in the format `ws-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"environment_id": {
				Description: "The environment that owns the variable, specified as an ID, in the format `env-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"account_id": {
				Description: "The account that owns the variable, specified as an ID, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},

			"updated_at": {
				Description: "Date/time the variable was updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"updated_by_email": {
				Description: "Email of the user who updated the variable last time.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"updated_by": {
				Description: "Details of the user that updated the variable last time.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Description: "Username of editor.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"email": {
							Description: "Email address of editor.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"full_name": {
							Description: "Full name of editor.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func resourceScalrVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	scalrClient := meta.(*scalr.Client)

	// Get key and category.
	key := d.Get("key").(string)
	category := scalr.CategoryType(d.Get("category").(string))
	hcl := d.Get("hcl").(bool)
	diags = append(diags, validateCategoryHCL(category, hcl)...)

	// Create a new options struct.
	options := scalr.VariableCreateOptions{
		Key:          scalr.String(key),
		Value:        scalr.String(d.Get("value").(string)),
		Description:  scalr.String(d.Get("description").(string)),
		Category:     scalr.Category(category),
		HCL:          scalr.Bool(hcl),
		Sensitive:    scalr.Bool(d.Get("sensitive").(bool)),
		Final:        scalr.Bool(d.Get("final").(bool)),
		QueryOptions: &scalr.VariableWriteQueryOptions{Force: scalr.Bool(d.Get("force").(bool))},
		Account:      &scalr.Account{ID: d.Get("account_id").(string)},
	}

	// Get and check the workspace.
	if workspaceID, ok := d.GetOk("workspace_id"); ok {
		ws, err := scalrClient.Workspaces.ReadByID(ctx, workspaceID.(string))
		if err != nil {
			return diag.Errorf(
				"Error retrieving workspace %s: %v", workspaceID, err)
		}
		options.Workspace = ws
	}

	// Get and check the environment
	if environmentId, ok := d.GetOk("environment_id"); ok {
		env, err := scalrClient.Environments.Read(ctx, environmentId.(string))
		if err != nil {
			return diag.Errorf(
				"Error retrieving environment %s: %v", environmentId, err)
		}
		options.Environment = env
	}

	log.Printf("[DEBUG] Create %s variable: %s", category, key)
	log.Printf("[DEBUG] Description: %s", *options.Description)
	variable, err := scalrClient.Variables.Create(ctx, options)
	if err != nil {
		return diag.Errorf("Error creating %s variable %s: %v", category, key, err)
	}

	d.SetId(variable.ID)

	return append(diags, resourceScalrVariableRead(ctx, d, meta)...)
}

func resourceScalrVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Read variable: %s", d.Id())
	variable, err := scalrClient.Variables.Read(ctx, d.Id())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Variable %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading variable %s: %v", d.Id(), err)
	}

	// Update config.
	_ = d.Set("key", variable.Key)
	_ = d.Set("category", string(variable.Category))
	_ = d.Set("hcl", variable.HCL)
	_ = d.Set("sensitive", variable.Sensitive)
	_ = d.Set("description", variable.Description)
	_ = d.Set("final", variable.Final)
	_ = d.Set("updated_by_email", variable.UpdatedByEmail)

	if variable.UpdatedAt != nil {
		_ = d.Set("updated_at", variable.UpdatedAt.Format(time.RFC3339))
	}

	var updatedBy []interface{}
	if variable.UpdatedBy != nil {
		updatedBy = append(updatedBy, map[string]interface{}{
			"username":  variable.UpdatedBy.Username,
			"email":     variable.UpdatedBy.Email,
			"full_name": variable.UpdatedBy.FullName,
		})
	}
	_ = d.Set("updated_by", updatedBy)

	_, exists := d.GetOk("force")
	if !exists {
		_ = d.Set("force", false)
	}

	if variable.Workspace != nil {
		_ = d.Set("workspace_id", variable.Workspace.ID)
	}

	if variable.Environment != nil {
		_ = d.Set("environment_id", variable.Environment.ID)
	}

	if variable.Account != nil {
		_ = d.Set("account_id", variable.Account.ID)
	}

	// Only set the value if it's not sensitive, as otherwise it will be empty.
	if !variable.Sensitive {
		_ = d.Set("value", variable.Value)
	}

	return nil
}

func resourceScalrVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	scalrClient := meta.(*scalr.Client)

	category := scalr.CategoryType(d.Get("category").(string))
	hcl := d.Get("hcl").(bool)
	diags = append(diags, validateCategoryHCL(category, hcl)...)

	// Create a new options struct.
	options := scalr.VariableUpdateOptions{
		Key:          scalr.String(d.Get("key").(string)),
		Value:        scalr.String(d.Get("value").(string)),
		HCL:          scalr.Bool(hcl),
		Sensitive:    scalr.Bool(d.Get("sensitive").(bool)),
		Description:  scalr.String(d.Get("description").(string)),
		Final:        scalr.Bool(d.Get("final").(bool)),
		QueryOptions: &scalr.VariableWriteQueryOptions{Force: scalr.Bool(d.Get("force").(bool))},
	}

	log.Printf("[DEBUG] Update variable: %s", d.Id())
	_, err := scalrClient.Variables.Update(ctx, d.Id(), options)
	if err != nil {
		return diag.Errorf("Error updating variable %s: %v", d.Id(), err)
	}

	return append(diags, resourceScalrVariableRead(ctx, d, meta)...)
}

func resourceScalrVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete variable: %s", d.Id())
	err := scalrClient.Variables.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting variable%s: %v", d.Id(), err)
	}

	return nil
}

// validateCategoryHCL checks if HCL is set to true for a category that is not 'terraform'
// and issues a warning if it is.
func validateCategoryHCL(c scalr.CategoryType, hcl bool) (diags diag.Diagnostics) {
	if c != scalr.CategoryTerraform && hcl {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       "HCL is not supported for shell variables",
			Detail:        "Setting 'hcl' attribute to 'true' for shell variable is now deprecated.",
			AttributePath: cty.GetAttrPath("hcl"),
		})
	}
	return
}
