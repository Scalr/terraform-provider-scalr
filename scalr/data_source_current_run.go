package scalr

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

// Note: The structure is similar to one from policy-check phase:
// https://iacp.docs.scalr.com/en/latest/working-with-iacp/opa.html#policy-checking-process
func dataSourceScalrCurrentRun() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrCurrentRunRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vcs": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						// TODO: add path
						"branch": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"commit": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sha": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"message": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"author": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"username": {
													Type:     schema.TypeString,
													Computed: true,
												},
												// TODO: add email and name
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_destroy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_dry": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			// TODO: add cost_estimate, credentials(?), created_by
		},
	}
}

func dataSourceScalrCurrentRunRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	runID, exists := os.LookupEnv("TFE_RUN_ID")
	if !exists {
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Read configuration of run: %s", runID)
	run, err := scalrClient.Runs.Read(ctx, runID)
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return fmt.Errorf("Could not find run %s", runID)
		}
		return fmt.Errorf("Error retrieving run: %v", err)
	}

	log.Printf("[DEBUG] Read workspace of run: %s", runID)
	workspace, err := scalrClient.Workspaces.ReadByID(ctx, run.Workspace.ID)
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return fmt.Errorf("Could not find workspace %s", run.Workspace.ID)
		}
		return fmt.Errorf("Error retrieving workspace: %v", err)
	}

	// Update the config
	d.Set("source", run.Source)
	d.Set("message", run.Message)
	d.Set("is_destroy", run.IsDestroy)
	d.Set("is_dry", run.Apply == nil)

	d.Set("workspace_name", workspace.Name)
	d.Set("environment_id", workspace.Organization.Name)

	if workspace.VCSRepo != nil {
		log.Printf("[DEBUG] Read ingress attributes of run: %s", runID)
		ingressAttributes, err := scalrClient.ConfigurationVersions.ReadIngressAttributes(ctx, run.ConfigurationVersion.ID)
		if err != nil {
			if err == scalr.ErrResourceNotFound {
				return fmt.Errorf("Could not find configuration version %s", run.ConfigurationVersion.ID)
			}
			return fmt.Errorf("Error retrieving ingress attributes: %v", err)
		}

		var commitConfig []map[string]interface{}
		commit := map[string]interface{}{
			"sha":     ingressAttributes.CommitSha,
			"message": ingressAttributes.CommitMessage,
			"author": map[string]interface{}{
				"username": ingressAttributes.SenderUsername,
			},
		}

		var vcsConfig []map[string]interface{}
		vcs := map[string]interface{}{
			"repository_id": workspace.VCSRepo.Identifier,
			"branch":        workspace.VCSRepo.Branch,
			"commit":        append(commitConfig, commit),
		}
		d.Set("vcs", append(vcsConfig, vcs))
	}

	d.SetId(runID)

	return nil
}
