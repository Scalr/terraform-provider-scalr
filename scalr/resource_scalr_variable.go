package scalr

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalrVariableCreate,
		ReadContext:   resourceScalrVariableRead,
		UpdateContext: resourceScalrVariableUpdate,
		DeleteContext: resourceScalrVariableDelete,
		CustomizeDiff: customdiff.All(
			func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
				// Reject change for key if variable is sensitive
				old, new := d.GetChange("key")
				sensitive := d.Get("sensitive")

				if sensitive.(bool) && (old.(string) != "" && old.(string) != new.(string)) {
					return fmt.Errorf("Error changing 'key' attribute for variable %s: immutable for sensitive variable", d.Id())
				}
				return nil
			},
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
				Type:     schema.TypeString,
				Required: true,
			},

			"value": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "",
				Sensitive: true,
			},

			"category": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"sensitive": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"final": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"force": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalrVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get key and category.
	key := d.Get("key").(string)
	category := scalr.CategoryType(d.Get("category").(string))

	// Create a new options struct.
	options := scalr.VariableCreateOptions{
		Key:          scalr.String(key),
		Value:        scalr.String(d.Get("value").(string)),
		Description:  scalr.String(d.Get("description").(string)),
		Category:     scalr.Category(category),
		HCL:          scalr.Bool(d.Get("hcl").(bool)),
		Sensitive:    scalr.Bool(d.Get("sensitive").(bool)),
		Final:        scalr.Bool(d.Get("final").(bool)),
		QueryOptions: &scalr.VariableWriteQueryOptions{Force: scalr.Bool(d.Get("force").(bool))},
	}

	// Get and check the workspace.
	if workspaceID, ok := d.GetOk("workspace_id"); ok {
		ws, err := scalrClient.Workspaces.ReadByID(ctx, workspaceID.(string))
		if err != nil {
			return diag.Errorf(
				"Error retrieving workspace %s: %v", workspaceID, err)
		}
		options.Workspace = ws
	} else {
		if category == scalr.CategoryTerraform {
			return diag.Errorf("Attribute 'workspace_id' is required for variable with category 'terraform'.")
		}
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

	// Get the account
	if accountId, ok := d.GetOk("account_id"); ok {
		options.Account = &scalr.Account{
			ID: accountId.(string),
		}
	}

	log.Printf("[DEBUG] Create %s variable: %s", category, key)
	log.Printf("[DEBUG] Description: %s", *options.Description)
	variable, err := scalrClient.Variables.Create(ctx, options)
	if err != nil {
		return diag.Errorf("Error creating %s variable %s: %v", category, key, err)
	}

	d.SetId(variable.ID)

	return resourceScalrVariableRead(ctx, d, meta)
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
	d.Set("key", variable.Key)
	d.Set("category", string(variable.Category))
	d.Set("hcl", variable.HCL)
	d.Set("sensitive", variable.Sensitive)
	d.Set("description", variable.Description)
	d.Set("final", variable.Final)
	_, exists := d.GetOk("force")
	if !exists {
		d.Set("force", false)
	}

	if variable.Workspace != nil {
		d.Set("workspace_id", variable.Workspace.ID)
	}

	if variable.Environment != nil {
		d.Set("environment_id", variable.Environment.ID)
	}

	if variable.Account != nil {
		d.Set("account_id", variable.Account.ID)
	}

	// Only set the value if it's not sensitive, as otherwise it will be empty.
	if !variable.Sensitive {
		d.Set("value", variable.Value)
	}

	return nil
}

func resourceScalrVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Create a new options struct.
	options := scalr.VariableUpdateOptions{
		Key:          scalr.String(d.Get("key").(string)),
		Value:        scalr.String(d.Get("value").(string)),
		HCL:          scalr.Bool(d.Get("hcl").(bool)),
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

	return resourceScalrVariableRead(ctx, d, meta)
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
