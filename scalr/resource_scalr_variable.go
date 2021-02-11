package scalr

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrVariableCreate,
		Read:   resourceScalrVariableRead,
		Update: resourceScalrVariableUpdate,
		Delete: resourceScalrVariableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceScalrVariableResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrVariableStateUpgradeV0,
				Version: 0,
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
				ForceNew: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalrVariableCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get key and category.
	key := d.Get("key").(string)
	category := scalr.CategoryType(d.Get("category").(string))

	// Create a new options struct.
	options := scalr.VariableCreateOptions{
		Key:          scalr.String(key),
		Value:        scalr.String(d.Get("value").(string)),
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
			return fmt.Errorf(
				"Error retrieving workspace %s: %v", workspaceID, err)
		}
		options.Workspace = ws
	} else {
		if category == scalr.CategoryTerraform {
			return fmt.Errorf("Only environment variables should be multi-scoped.")
		}
	}

	// Get and check the environment
	if environmentId, ok := d.GetOk("environment_id"); ok {
		env, err := scalrClient.Environments.Read(ctx, environmentId.(string))
		if err != nil {
			return fmt.Errorf(
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
	variable, err := scalrClient.Variables.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("Error creating %s variable %s: %v", category, key, err)
	}

	d.SetId(variable.ID)

	return resourceScalrVariableRead(d, meta)
}

func resourceScalrVariableRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Read variable: %s", d.Id())
	variable, err := scalrClient.Variables.Read(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading variable %s: %v", d.Id(), err)
	}

	// Update config.
	d.Set("key", variable.Key)
	d.Set("category", string(variable.Category))
	d.Set("hcl", variable.HCL)
	d.Set("sensitive", variable.Sensitive)
	d.Set("final", variable.Final)
	_, exists := d.GetOk("force")
	if !exists {
		d.Set("force", false)
	}
	if variable.Account != nil {
		d.Set("account_id", variable.Account.ID)
	}
	if variable.Environment != nil {
		d.Set("environment_id", variable.Environment.ID)
	}
	if variable.Workspace != nil {
		d.Set("workspace_id", variable.Workspace.ID)
	}

	// Only set the value if its not sensitive, as otherwise it will be empty.
	if !variable.Sensitive {
		d.Set("value", variable.Value)
	}

	return nil
}

func resourceScalrVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Create a new options struct.
	options := scalr.VariableUpdateOptions{
		Key:          scalr.String(d.Get("key").(string)),
		Value:        scalr.String(d.Get("value").(string)),
		HCL:          scalr.Bool(d.Get("hcl").(bool)),
		Sensitive:    scalr.Bool(d.Get("sensitive").(bool)),
		Final:        scalr.Bool(d.Get("final").(bool)),
		QueryOptions: &scalr.VariableWriteQueryOptions{Force: scalr.Bool(d.Get("force").(bool))},
	}

	log.Printf("[DEBUG] Update variable: %s", d.Id())
	_, err := scalrClient.Variables.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating variable %s: %v", d.Id(), err)
	}

	return resourceScalrVariableRead(d, meta)
}

func resourceScalrVariableDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete variable: %s", d.Id())
	err := scalrClient.Variables.Delete(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting variable%s: %v", d.Id(), err)
	}

	return nil
}
