package provider

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrVcsProvider() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of a VCS provider.",
		ReadContext: dataSourceScalrVcsProviderRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "Identifier of the VCS provider.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"name": {
				Description:  "Name of the VCS provider.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"vcs_type": {
				Description: "Type of the VCS provider. For example, `github`.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"url": {
				Description: "The URL to the VCS provider installation.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"account_id": {
				Description: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"environment_id": {
				Description: "ID of the environment the VCS provider has to be linked to, in the format `env-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"agent_pool_id": {
				Description: "The id of the agent pool to connect Scalr to self-hosted VCS provider, in the format `apool-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"environments": {
				Description: "List of the identifiers of environments the VCS provider is linked to.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"draft_pr_runs_enabled": {
				Description: "Indicates whether the draft pull-request runs are enabled for this VCS provider.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		}}
}

func dataSourceScalrVcsProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	options := scalr.VcsProvidersListOptions{
		Account: ptr(d.Get("account_id").(string)),
	}

	if vcsProviderID, ok := d.GetOk("id"); ok {
		options.ID = ptr(vcsProviderID.(string))
	}

	if name, ok := d.GetOk("name"); ok {
		options.Query = ptr(name.(string))
	}

	if envId, ok := d.GetOk("environment_id"); ok {
		options.Environment = ptr(envId.(string))
	}

	if agentPoolID, ok := d.GetOk("agent_pool_id"); ok {
		options.AgentPool = ptr(agentPoolID.(string))
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

	envIds := make([]string, 0)
	for _, env := range vcsProvider.Environments {
		envIds = append(envIds, env.ID)
	}
	sort.Strings(envIds)

	// Update the configuration.
	_ = d.Set("vcs_type", vcsProvider.VcsType)
	_ = d.Set("name", vcsProvider.Name)
	_ = d.Set("url", vcsProvider.Url)
	_ = d.Set("environments", envIds)
	_ = d.Set("draft_pr_runs_enabled", vcsProvider.DraftPrRunsEnabled)
	if vcsProvider.AgentPool != nil {
		_ = d.Set("agent_pool_id", vcsProvider.AgentPool.ID)
	}
	d.SetId(vcsProvider.ID)

	return nil
}
