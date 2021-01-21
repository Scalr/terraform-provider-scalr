package scalr

import (
	"fmt"
	"log"
	"strings"

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
			State: resourceScalrVariableImporter,
		},

		SchemaVersion: 2,
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
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceScalrVariableCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get key and category.
	key := d.Get("key").(string)
	category := d.Get("category").(string)

	// Create a new options struct.
	options := scalr.VariableCreateOptions{
		Key:       scalr.String(key),
		Value:     scalr.String(d.Get("value").(string)),
		Category:  scalr.Category(scalr.CategoryType(category)),
		HCL:       scalr.Bool(d.Get("hcl").(bool)),
		Sensitive: scalr.Bool(d.Get("sensitive").(bool)),
		Final:     scalr.Bool(d.Get("final").(bool)),
		Force:     scalr.Bool(d.Get("force").(bool)),
	}

	// Get and check the workspace.
	if workspaceID, ok := d.GetOk("workspace_id"); ok {
		ws, err := scalrClient.Workspaces.ReadByID(ctx, workspaceID.(string))
		if err != nil {
			return fmt.Errorf(
				"Error retrieving workspace %s: %v", workspaceID, err)
		}
		options.Workspace = ws
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
	d.Set("force", variable.Force)

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
		Key:       scalr.String(d.Get("key").(string)),
		Value:     scalr.String(d.Get("value").(string)),
		HCL:       scalr.Bool(d.Get("hcl").(bool)),
		Sensitive: scalr.Bool(d.Get("sensitive").(bool)),
		Final:     scalr.Bool(d.Get("final").(bool)),
		Force:     scalr.Bool(d.Get("force").(bool)),
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

func resourceScalrVariableImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	scalrClient := meta.(*scalr.Client)
	s := strings.SplitN(d.Id(), "/", 3)
	log.Printf("[DEBUG] in resourceScalrVariableImporter: %s", s[0])
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid variable import format: %s (expected <ENVIRONMENT ID>/<WORKSPACE NAME>/<VARIABLE ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	workspaceID, err := fetchWorkspaceID(s[0]+"/"+s[1], scalrClient)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving workspace %s from environment %s: %v", s[1], s[0], err)
	}
	d.Set("workspace_id", workspaceID)
	d.SetId(s[2])

	return []*schema.ResourceData{d}, nil
}
