package scalr

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrPolicyGroupLinkage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalrPolicyGroupLinkageCreate,
		ReadContext:   resourceScalrPolicyGroupLinkageRead,
		DeleteContext: resourceScalrPolicyGroupLinkageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceScalrPolicyGroupLinkageImport,
		},

		Schema: map[string]*schema.Schema{
			"policy_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalrPolicyGroupLinkageImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	policyGroup, environment, err := getLinkedResources(ctx, id, scalrClient)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil, fmt.Errorf("policy group linkage %s not found", id)
		}
		return nil, fmt.Errorf("error retrieving policy group linkage %s: %v", id, err)
	}

	_ = d.Set("policy_group_id", policyGroup.ID)
	_ = d.Set("environment_id", environment.ID)

	return []*schema.ResourceData{d}, nil
}

func resourceScalrPolicyGroupLinkageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	pgID := d.Get("policy_group_id").(string)
	envID := d.Get("environment_id").(string)
	id := packPolicyGroupLinkageID(pgID, envID)

	opts := scalr.PolicyGroupEnvironmentsCreateOptions{
		PolicyGroupID:           pgID,
		PolicyGroupEnvironments: []*scalr.PolicyGroupEnvironment{{ID: envID}},
	}
	err := scalrClient.PolicyGroupEnvironments.Create(ctx, opts)
	if err != nil {
		return diag.Errorf("error creating policy group linkage %s: %v", id, err)
	}

	d.SetId(id)
	return resourceScalrPolicyGroupLinkageRead(ctx, d, meta)
}

func resourceScalrPolicyGroupLinkageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	policyGroup, environment, err := getLinkedResources(ctx, id, scalrClient)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Policy group linkage %s not found", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving policy group linkage %s: %v", id, err)
	}

	_ = d.Set("policy_group_id", policyGroup.ID)
	_ = d.Set("environment_id", environment.ID)

	return nil
}

func resourceScalrPolicyGroupLinkageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	policyGroup, environment, err := getLinkedResources(ctx, id, scalrClient)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Policy group linkage %s not found", id)
			return nil
		}
		return diag.Errorf("error deleting policy group linkage %s: %v", id, err)
	}

	opts := scalr.PolicyGroupEnvironmentDeleteOptions{PolicyGroupID: policyGroup.ID, EnvironmentID: environment.ID}
	err = scalrClient.PolicyGroupEnvironments.Delete(ctx, opts)
	if err != nil {
		return diag.Errorf("error deleting policy group linkage %s: %v", id, err)
	}

	return nil
}

// getLinkedResources verifies existence of the linkage
// and returns associated policy group and environment.
func getLinkedResources(ctx context.Context, id string, scalrClient *scalr.Client) (
	policyGroup *scalr.PolicyGroup, environment *scalr.Environment, err error,
) {
	pgID, envID, err := unpackPolicyGroupLinkageID(id)
	if err != nil {
		return
	}

	environment, err = scalrClient.Environments.Read(ctx, envID)
	if err != nil {
		return
	}

	for _, pg := range environment.PolicyGroups {
		if pg.ID == pgID {
			policyGroup = pg
			break
		}
	}
	if policyGroup == nil {
		return nil, nil, scalr.ErrResourceNotFound
	}

	return
}

func packPolicyGroupLinkageID(pgID, envID string) string {
	return pgID + "/" + envID
}

func unpackPolicyGroupLinkageID(id string) (pgID, envID string, err error) {
	if s := strings.SplitN(id, "/", 2); len(s) == 2 {
		return s[0], s[1], nil
	}
	return "", "", fmt.Errorf(
		"invalid policy group linkage ID format: %s (expected <policy_group_id>/<environment_id>", id,
	)
}
