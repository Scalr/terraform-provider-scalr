package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrModule() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrModuleCreate,
		Read:   resourceScalrModuleRead,
		Delete: resourceScalrModuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"module_provider": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vcs_repo": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"tag_prefix": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"vcs_provider_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"environment_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceScalrModuleCreate(d *schema.ResourceData, meta interface{}) error {
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
		VCSRepo:     vcsOpt,
		VcsProvider: &scalr.VcsProvider{ID: d.Get("vcs_provider_id").(string)},
	}

	if accID, ok := d.GetOk("account_id"); ok {
		opt.Account = &scalr.Account{ID: accID.(string)}
	}

	if envID, ok := d.GetOk("environment_id"); ok {
		opt.Environment = &scalr.Environment{ID: envID.(string)}
	}

	m, err := scalrClient.Modules.Create(ctx, opt)
	if err != nil {
		return fmt.Errorf("Error creating module: %v", err)
	}

	d.SetId(m.ID)
	return resourceScalrModuleRead(d, meta)
}

func resourceScalrModuleRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()
	log.Printf("[DEBUG] Read configuration of module: %s", id)
	m, err := scalrClient.Modules.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			log.Printf("[DEBUG] Module %s no longer exists", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of module %s: %v", id, err)
	}

	// Update the config.
	d.Set("name", m.Name)
	d.Set("provider", m.Provider)
	d.Set("status", m.Status)
	d.Set("source", m.Source)
	d.Set("vcs_repo", []map[string]interface{}{{
		"identifier": m.VCSRepo.Identifier,
		"path":       m.VCSRepo.Path,
		"tag_prefix": m.VCSRepo.TagPrefix,
	}})
	d.Set("vcs_provider_id", m.VcsProvider.ID)

	if m.Account != nil {
		d.Set("account_id", m.Account.ID)
	}
	if m.Environment != nil {
		d.Set("environment_id", m.Environment.ID)
	}

	return nil
}

func resourceScalrModuleDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete module %s", id)
	err := scalrClient.Modules.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			return nil
		}
		return fmt.Errorf("Error deleting module %s: %v", id, err)
	}

	return nil
}
