package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrAgentPool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrAgentPoolRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"workspace_ids": {
				Type:     schema.TypeList,
				Computed: true,
				MinItems: 0,
				MaxItems: 128,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceScalrAgentPoolRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	var envID string

	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)
	options := scalr.AgentPoolListOptions{
		Name:    name,
		Account: scalr.String(accountID),
	}

	if envID, ok := d.GetOk("environment_id"); ok {
		options.Environment = scalr.String(envID.(string))
	}

	agentPoolsList, err := scalrClient.AgentPools.List(ctx, options)
	if err != nil {
		return fmt.Errorf("Error retrieving agent pool: %v", err)
	}

	if len(agentPoolsList.Items) > 1 {
		return errors.New("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if len(agentPoolsList.Items) == 0 {
		return fmt.Errorf("Could not find agent pool with name '%s', account_id: '%s', and environment_id: '%s'", name, accountID, envID)
	}

	agentPool := agentPoolsList.Items[0]

	workspaces := make([]string, 0)
	if len(agentPool.Workspaces) != 0 {
		for _, workspace := range agentPool.Workspaces {
			workspaces = append(workspaces, workspace.ID)
		}

		log.Printf("[DEBUG] agent pool %s workspaces: %+v", agentPool.ID, workspaces)
		d.Set("workspace_ids", workspaces)
	}
	d.SetId(agentPool.ID)

	return nil
}
