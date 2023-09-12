package scalr

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

const (
	currentRunIDEnvVar = "SCALR_RUN_ID"
)

// Note: The structure is similar to one from policy-check phase:
// https://iacp.docs.scalr.com/en/latest/working-with-iacp/opa.html#policy-checking-process
func dataSourceScalrCurrentRun() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrCurrentRunRead,
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
										Type:     schema.TypeList,
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

func dataSourceScalrCurrentRunRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	runID, exists := os.LookupEnv(currentRunIDEnvVar)
	if !exists {
		log.Printf("[DEBUG] %s is not set", currentRunIDEnvVar)
		d.SetId(dummyIdentifier)
		return nil
	}

	log.Printf("[DEBUG] Read configuration of run: %s", runID)
	run, err := scalrClient.Runs.Read(ctx, runID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("Could not find run %s", runID)
		}
		return diag.Errorf("Error retrieving run: %v", err)
	}

	log.Printf("[DEBUG] Read workspace of run: %s", runID)
	workspace, err := scalrClient.Workspaces.ReadByID(ctx, run.Workspace.ID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("Could not find workspace %s", run.Workspace.ID)
		}
		return diag.Errorf("Error retrieving workspace: %v", err)
	}

	// Update the config
	_ = d.Set("source", run.Source)
	_ = d.Set("message", run.Message)
	_ = d.Set("is_destroy", run.IsDestroy)
	_ = d.Set("is_dry", run.Apply == nil)

	_ = d.Set("workspace_name", workspace.Name)
	_ = d.Set("environment_id", workspace.Environment.ID)

	if workspace.VCSRepo != nil {
		log.Printf("[DEBUG] Read vcs revision attributes of run: %s", runID)
		var vcsConfig []map[string]interface{}
		vcs := map[string]interface{}{
			"repository_id": workspace.VCSRepo.Identifier,
			"branch":        workspace.VCSRepo.Branch,
			"commit":        []map[string]interface{}{},
		}

		if run.VcsRevision != nil {
			vcs["commit"] = []map[string]interface{}{
				{
					"sha":     run.VcsRevision.CommitSha,
					"message": run.VcsRevision.CommitMessage,
					"author": []interface{}{
						map[string]string{
							"username": run.VcsRevision.SenderUsername,
						},
					},
				},
			}
		}

		_ = d.Set("vcs", append(vcsConfig, vcs))
	}

	d.SetId(runID)

	return nil
}
