package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrAgentPool() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of an agent pool.",
		ReadContext: dataSourceScalrAgentPoolRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "ID of the agent pool.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"id", "name"},
			},
			"name": {
				Description:  "A name of the agent pool.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"vcs_enabled": {
				Description: "Indicates whether the VCS support is enabled for agents in the pool.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"account_id": {
				Description: "An identifier of the Scalr account.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},

			"environment_id": {
				Description: "An identifier of the Scalr environment.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"workspace_ids": {
				Description: "The list of IDs of linked workspaces.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceScalrAgentPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	agentPoolID := d.Get("id").(string)
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)
	envID := d.Get("environment_id").(string)

	options := scalr.AgentPoolListOptions{
		Account: scalr.String(accountID),
	}

	if agentPoolID != "" {
		options.AgentPool = agentPoolID
	}

	if name != "" {
		options.Name = name
	}

	if envID != "" {
		options.Environment = scalr.String(envID)
	}

	if vcsEnabled, ok := d.GetOkExists("vcs_enabled"); ok { //nolint:staticcheck
		options.VcsEnabled = scalr.Bool(vcsEnabled.(bool))
	}

	agentPoolsList, err := scalrClient.AgentPools.List(ctx, options)
	if err != nil {
		return diag.Errorf("Error retrieving agent pool: %v", err)
	}

	if len(agentPoolsList.Items) > 1 {
		return diag.FromErr(errors.New("Your query returned more than one result. Please try a more specific search criteria."))
	}

	if len(agentPoolsList.Items) == 0 {
		return diag.Errorf("Could not find agent pool with ID '%s', name '%s', account_id '%s', and environment_id '%s'", agentPoolID, name, accountID, envID)
	}

	agentPool := agentPoolsList.Items[0]

	workspaces := make([]string, 0)
	if len(agentPool.Workspaces) != 0 {
		for _, workspace := range agentPool.Workspaces {
			workspaces = append(workspaces, workspace.ID)
		}

		log.Printf("[DEBUG] agent pool %s workspaces: %+v", agentPool.ID, workspaces)
		_ = d.Set("workspace_ids", workspaces)
	}
	_ = d.Set("vcs_enabled", agentPool.VcsEnabled)
	_ = d.Set("name", agentPool.Name)
	d.SetId(agentPool.ID)

	return nil
}
