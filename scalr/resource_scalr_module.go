package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrModule() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages the state of a module in the Private Modules Registry. Create and destroy operations are available only.",
		CreateContext: resourceScalrModuleCreate,
		ReadContext:   resourceScalrModuleRead,
		DeleteContext: resourceScalrModuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the module, e.g. `rds`, `compute`, `kubernetes-engine`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"module_provider": {
				Description: "Module provider name, e.g `aws`, `azurerm`, `google`, etc.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "A system status of the Module.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"source": {
				Description: "The source of a remote module in the private registry, e.g `env-xxxx/aws/vpc`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vcs_repo": {
				Description: "Source configuration of a VCS repository.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				MaxItems:    1,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Description: "The identifier of a VCS repository in the format `:org/:repo` (`:org/:project/:name` is used for Azure DevOps). It refers to an organization and a repository name in a VCS provider.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"path": {
							Description: "The path to the root module folder. It is expected to have the format `<path>/terraform-<provider_name>-<module_name>`, where `<path>` stands for any folder within the repository inclusively a repository root.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"tag_prefix": {
							Description: "Registry ignores tags which do not match specified prefix, e.g. `aws/`.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
					},
				},
			},
			"vcs_provider_id": {
				Description: "The identifier of a VCS provider in the format `vcs-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"account_id": {
				Description: "The identifier of the account in the format `acc-<RANDOM STRING>`. If it is not specified the module will be registered globally and available across the whole installation.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"environment_id": {
				Description: "The identifier of an environment in the format `env-<RANDOM STRING>`. If it is not specified the module will be registered at the account level and available across all environments within the account specified in `account_id` attribute.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
			},
		},
	}
}

func resourceScalrModuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	vcsRepo := d.Get("vcs_repo").([]interface{})[0].(map[string]interface{})
	vcsOpt := &scalr.ModuleVCSRepo{
		Identifier: *scalr.String(vcsRepo["identifier"].(string)),
	}
	if path, ok := vcsRepo["path"].(string); ok && path != "" {
		vcsOpt.Path = scalr.String(path)
	}
	if prefix, ok := vcsRepo["tag_prefix"].(string); ok && prefix != "" {
		vcsOpt.TagPrefix = scalr.String(prefix)
	}

	opt := scalr.ModuleCreateOptions{
		Account:     &scalr.Account{ID: d.Get("account_id").(string)},
		VCSRepo:     vcsOpt,
		VcsProvider: &scalr.VcsProvider{ID: d.Get("vcs_provider_id").(string)},
	}

	if envID, ok := d.GetOk("environment_id"); ok {
		opt.Environment = &scalr.Environment{ID: envID.(string)}
	}

	m, err := scalrClient.Modules.Create(ctx, opt)
	if err != nil {
		return diag.Errorf("Error creating module: %v", err)
	}

	d.SetId(m.ID)
	return resourceScalrModuleRead(ctx, d, meta)
}

func resourceScalrModuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()
	log.Printf("[DEBUG] Read configuration of module: %s", id)
	m, err := scalrClient.Modules.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Module %s no longer exists", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading configuration of module %s: %v", id, err)
	}

	// Update the config.
	_ = d.Set("name", m.Name)
	_ = d.Set("module_provider", m.Provider)
	_ = d.Set("status", m.Status)
	_ = d.Set("source", m.Source)
	_ = d.Set("vcs_repo", []map[string]interface{}{{
		"identifier": m.VCSRepo.Identifier,
		"path":       m.VCSRepo.Path,
		"tag_prefix": m.VCSRepo.TagPrefix,
	}})
	_ = d.Set("vcs_provider_id", m.VcsProvider.ID)

	if m.Account != nil {
		_ = d.Set("account_id", m.Account.ID)
	}
	if m.Environment != nil {
		_ = d.Set("environment_id", m.Environment.ID)
	}

	return nil
}

func resourceScalrModuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete module %s", id)
	err := scalrClient.Modules.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting module %s: %v", id, err)
	}

	return nil
}
