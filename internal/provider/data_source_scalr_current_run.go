package provider

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
		Description: "Allows you to get information about the current Terraform run" +
			" when using a Scalr remote backend workspace, including VCS (Git) metadata." +
			"\n\nNo arguments are required. The data source returns details of the current run based on the" +
			" `SCALR_RUN_ID` shell variable that is automatically exported in the Scalr remoted backend.",
		ReadContext: dataSourceScalrCurrentRunRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the run, in the format `run-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"environment_id": {
				Description: "The ID of the environment, in the format `env-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"workspace_name": {
				Description: "Workspace name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vcs": {
				Description: "Contains details of the VCS configuration if the workspace is linked to a VCS repo.",
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_id": {
							Description: "ID of the VCS repo in the format `:org/:repo`.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						// TODO: add path
						"branch": {
							Description: "The linked VCS repo branch.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"commit": {
							Description: "Details of the last commit to the linked VCS repo.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sha": {
										Description: "SHA of the last commit.",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"message": {
										Description: "Message for the last commit.",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"author": {
										Description: "Details of the author of the last commit.",
										Type:        schema.TypeList,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"username": {
													Description: "Username of the author in the VCS.",
													Type:        schema.TypeString,
													Computed:    true,
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
				Description: "The source of the run (VCS, API, Manual).",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"message": {
				Description: "Message describing how the run was triggered.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"is_destroy": {
				Description: "Boolean indicates if this is a \"destroy\" run.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"is_dry": {
				Description: "Boolean indicates if this is a dry run, i.e. triggered by a Pull Request (PR). No apply phase if this is true.",
				Type:        schema.TypeBool,
				Computed:    true,
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
