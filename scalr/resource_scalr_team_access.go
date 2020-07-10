package scalr

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	scalr "github.com/scalr/go-scalr"
)

func resourceTFETeamAccess() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamAccessCreate,
		Read:   resourceTFETeamAccessRead,
		Delete: resourceTFETeamAccessDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFETeamAccessImporter,
		},

		Schema: map[string]*schema.Schema{
			"access": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.AccessAdmin),
						string(scalr.AccessRead),
						string(scalr.AccessPlan),
						string(scalr.AccessWrite),
					},
					false,
				),
			},

			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFETeamAccessCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get access and team ID.
	access := d.Get("access").(string)
	teamID := d.Get("team_id").(string)

	// Get organization and workspace.
	organization, workspace, err := unpackWorkspaceID(d.Get("workspace_id").(string))
	if err != nil {
		return fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	// Get the team.
	tm, err := scalrClient.Teams.Read(ctx, teamID)
	if err != nil {
		return fmt.Errorf("Error retrieving team %s: %v", teamID, err)
	}

	// Get the workspace.
	ws, err := scalrClient.Workspaces.Read(ctx, organization, workspace)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s from organization %s: %v", workspace, organization, err)
	}

	// Create a new options struct.
	options := scalr.TeamAccessAddOptions{
		Access:    scalr.Access(scalr.AccessType(access)),
		Team:      tm,
		Workspace: ws,
	}

	log.Printf("[DEBUG] Give team %s %s access to workspace: %s", tm.Name, access, ws.Name)
	tmAccess, err := scalrClient.TeamAccess.Add(ctx, options)
	if err != nil {
		return fmt.Errorf(
			"Error giving team %s %s access to workspace %s: %v", tm.Name, access, ws.Name, err)
	}

	d.SetId(tmAccess.ID)

	return resourceTFETeamAccessRead(d, meta)
}

func resourceTFETeamAccessRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Read configuration of team access: %s", d.Id())
	tmAccess, err := scalrClient.TeamAccess.Read(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			log.Printf("[DEBUG] Team access %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of team access %s: %v", d.Id(), err)
	}

	// Update config.
	d.Set("access", string(tmAccess.Access))

	if tmAccess.Team != nil {
		d.Set("team_id", tmAccess.Team.ID)
	} else {
		d.Set("team_id", "")
	}

	return nil
}

func resourceTFETeamAccessDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete team access: %s", d.Id())
	err := scalrClient.TeamAccess.Remove(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting team access %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFETeamAccessImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.SplitN(d.Id(), "/", 3)
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid team access import format: %s (expected <ORGANIZATION>/<WORKSPACE>/<TEAM ACCESS ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("workspace_id", s[0]+"/"+s[1])
	d.SetId(s[2])

	return []*schema.ResourceData{d}, nil
}
