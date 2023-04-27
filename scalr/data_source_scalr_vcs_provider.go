package scalr

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrVcsProvider() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrVcsProviderRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
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
			"agent_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"environments": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		}}
}

func dataSourceScalrVcsProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	options := scalr.VcsProvidersListOptions{
		Account: scalr.String(d.Get("account_id").(string)),
	}

	if vcsProviderID, ok := d.GetOk("id"); ok {
		options.ID = scalr.String(vcsProviderID.(string))
	}

	if name, ok := d.GetOk("name"); ok {
		options.Query = scalr.String(name.(string))
	}

	if envId, ok := d.GetOk("environment_id"); ok {
		options.Environment = scalr.String(envId.(string))
	}

	if agentPoolID, ok := d.GetOk("agent_pool_id"); ok {
		options.AgentPool = scalr.String(agentPoolID.(string))
	}

	if vcsType, ok := d.GetOk("vcs_type"); ok {
		vcsType := scalr.VcsType(vcsType.(string))
		options.VcsType = &vcsType
	}

	vcsProviders, err := scalrClient.VcsProviders.List(ctx, options)

	if err != nil {
		return diag.Errorf("Error retrieving vcs provider: %s.", err)
	}

	if vcsProviders.TotalCount > 1 {
		return diag.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if vcsProviders.TotalCount == 0 {
		return diag.Errorf("Could not find vcs provider matching you query.")
	}

	vcsProvider := vcsProviders.Items[0]

	if vcsProvider.IsShared {
		_ = d.Set("environments", []string{"*"})
	} else {
		envIds := make([]string, 0)
		for _, environment := range vcsProvider.Environments {
			envIds = append(envIds, environment.ID)
		}
		_ = d.Set("environments", envIds)
	}
	// Update the configuration.
	_ = d.Set("vcs_type", vcsProvider.VcsType)
	_ = d.Set("name", vcsProvider.Name)
	_ = d.Set("url", vcsProvider.Url)
	if vcsProvider.AgentPool != nil {
		_ = d.Set("agent_pool_id", vcsProvider.AgentPool.ID)
	}
	d.SetId(vcsProvider.ID)

	return nil
}
