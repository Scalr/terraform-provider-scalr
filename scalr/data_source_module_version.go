package scalr

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceModuleVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceModuleVersionRead,
		Schema: map[string]*schema.Schema{
			"source": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		}}
}

func dataSourceModuleVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	source := d.Get("source").(string)
	module, err := scalrClient.Modules.ReadBySource(ctx, source)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("Could not find module with source %s", source)
		}
		return diag.Errorf("Error retrieving module: %v", err)
	}
	log.Printf("[DEBUG] Download module by source: %s", source)

	var mv *scalr.ModuleVersion
	var version string
	if v, ok := d.GetOk("version"); ok {
		version = v.(string)
		mv, err = scalrClient.ModuleVersions.ReadBySemanticVersion(ctx, module.ID, version)
	} else {
		if module.ModuleVersion == nil {
			return diag.FromErr(errors.New("The module has no version tags"))
		}
		mv, err = scalrClient.ModuleVersions.Read(ctx, module.ModuleVersion.ID)
	}

	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("Could not find module with source %s  and version %s", source, version)
		}
		return diag.Errorf("Error retrieving module version: %v", err)
	}
	log.Printf("[DEBUG] Download module version by source %s version: %s", source, version)

	d.SetId(mv.ID)
	d.Set("version", mv.Version)
	return nil
}
