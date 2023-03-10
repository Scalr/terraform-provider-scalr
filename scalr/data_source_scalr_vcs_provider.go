package scalr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrVcsProvider() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrVcsProviderRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vcs_type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"environments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		}}
}

func dataSourceScalrVcsProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	options := GetVcsProviderOptions{
		Account: scalr.String(d.Get("account_id").(string)),
	}

	if name, ok := d.GetOk("name"); ok {
		options.Name = scalr.String(name.(string))
	}

	if envId, ok := d.GetOk("environment_id"); ok {
		options.Environment = scalr.String(envId.(string))
	}

	if vcsType, ok := d.GetOk("vcs_type"); ok {
		vcsType := scalr.VcsType(vcsType.(string))
		options.VcsType = &vcsType
	}

	vcsProvider, err := GetVcsProvider(ctx, options, scalrClient)
	if err != nil {
		return diag.Errorf("Error retrieving vcs provider: %s.", err)
	}

	envIds := make([]string, 0)
	for _, env := range vcsProvider.Environments {
		envIds = append(envIds, env.ID)
	}

	// Update the configuration.
	_ = d.Set("vcs_type", vcsProvider.VcsType)
	_ = d.Set("name", vcsProvider.Name)
	_ = d.Set("url", vcsProvider.Url)
	_ = d.Set("environments", envIds)
	d.SetId(vcsProvider.ID)

	return nil
}
