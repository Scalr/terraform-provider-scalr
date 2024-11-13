package scalr

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceModuleVersions() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a list of module versions by module source or module id.",
		ReadContext: dataSourceModuleVersionsRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "The identifier of а module. Example: `mod-xxxx`",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "source"},
			},
			"source": {
				Description: "The source of a module.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"versions": {
				Description: "The list of semantic versions.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The identifier of а module version. Example: `modver-xxxx`",
							Type:        schema.TypeString,
							Required:    true,
						},
						"version": {
							Description: "The semantic version. Example: `1.2.3`.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"status": {
							Description: "The status of a module version. Possible values: `ok`, `pending`, `not_uploaded`, `errored`",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		}}
}

func dataSourceModuleVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	moduleID := d.Get("id").(string)
	moduleSource := d.Get("source").(string)

	var module *scalr.Module

	if moduleID != "" {
		log.Printf("[DEBUG] Read module with ID: %s", moduleID)
		var err error
		module, err = scalrClient.Modules.Read(ctx, moduleID)
		if err != nil {
			if errors.Is(err, scalr.ErrResourceNotFound) {
				return diag.Errorf("Could not find module with ID '%s'", moduleID)
			}
			return diag.Errorf("Error retrieving module: %v", err)
		}

		if moduleSource == "" {
			_ = d.Set("source", module.Source)
		} else if module.Source != moduleSource {
			return diag.Errorf("Could not find module with ID '%s' and source '%s'", moduleID, moduleSource)
		}

	} else if moduleSource != "" {
		log.Printf("[DEBUG] Read module with source: %s", moduleSource)
		var err error
		module, err = scalrClient.Modules.ReadBySource(ctx, moduleSource)
		if err != nil {
			if errors.Is(err, scalr.ErrResourceNotFound) {
				return diag.Errorf("Could not find module with source '%s'", moduleSource)
			}
			return diag.Errorf("Error retrieving module: %v", err)
		}

	} else {
		return diag.Errorf("Error retrieving module: either 'id' or 'source' is required")
	}

	log.Printf("[DEBUG] Read versions of module: %s", module.ID)

	versions := make([]map[string]interface{}, 0)
	options := scalr.ModuleVersionListOptions{Module: module.ID}
	for {
		page, err := scalrClient.ModuleVersions.List(ctx, options)
		if err != nil {
			return diag.Errorf("Error retrieving versions of module with ID '%s': %v", module.ID, err)
		}
		for _, version := range page.Items {
			versionItem := map[string]interface{}{
				"id":      version.ID,
				"version": version.Version,
				"status":  string(version.Status),
			}
			versions = append(versions, versionItem)
		}
		if page.CurrentPage >= page.TotalPages {
			break
		}
		options.PageNumber = page.NextPage
	}
	_ = d.Set("versions", versions)

	d.SetId(module.ID)
	return nil
}
