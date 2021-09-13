package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceModuleVersion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceModuleVersionRead,
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

func dataSourceModuleVersionRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	source := d.Get("source").(string)
	module, err := scalrClient.Modules.ReadBySource(ctx, source)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			return fmt.Errorf("Could not find module with source %s", source)
		}
		return fmt.Errorf("Error retrieving module: %v", err)
	}
	log.Printf("[DEBUG] Download module by source: %s", source)

	var mv *scalr.ModuleVersion
	var version string
	if v, ok := d.GetOk("version"); ok {
		version = v.(string)
		mv, err = scalrClient.ModuleVersions.ReadBySemanticVersion(ctx, module.ID, version)
	} else {
		if module.LatestModuleVersion == nil {
			return errors.New("The module has no version tags")
		}
		mv, err = scalrClient.ModuleVersions.Read(ctx, module.LatestModuleVersion.ID)
	}

	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			return fmt.Errorf("Could not find module with source %s  and version %s", source, version)
		}
		return fmt.Errorf("Error retrieving module version: %v", err)
	}
	log.Printf("[DEBUG] Download module version by source %s version: %s", source, version)

	d.SetId(mv.ID)
	d.Set("version", mv.Version)
	return nil
}
