package scalr

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrPolicyGroupLinkage() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrPolicyGroupLinkageCreate,
		Read:   resourceScalrPolicyGroupLinkageRead,
		Delete: resourceScalrPolicyGroupLinkageDelete,
		Importer: &schema.ResourceImporter{
			State: resourceScalrPolicyGroupLinkageImport,
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

func resourceScalrPolicyGroupLinkageImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	policyGroup, environment, err := getLinkedResources(id, scalrClient)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			return nil, fmt.Errorf("policy group linkage %s not found", id)
		}
		return nil, fmt.Errorf("error retrieving policy group linkage %s: %v", id, err)
	}

	d.Set("policy_group_id", policyGroup.ID)
	d.Set("environment_id", environment.ID)

	return []*schema.ResourceData{d}, nil
}

func resourceScalrPolicyGroupLinkageCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	pgID := d.Get("policy_group_id").(string)
	envID := d.Get("environment_id").(string)
	id := packPolicyGroupLinkageID(pgID, envID)

	environment, err := scalrClient.Environments.Read(ctx, envID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			return fmt.Errorf("environment %s not found", envID)
		}
		return fmt.Errorf("error creating policy group linkage %s: %v", id, err)
	}

	// existing policy groups of the environment plus the new one
	policyGroups := append(environment.PolicyGroups, &scalr.PolicyGroup{ID: pgID})

	opts := scalr.EnvironmentUpdateOptions{PolicyGroups: policyGroups}
	_, err = scalrClient.Environments.Update(ctx, envID, opts)
	if err != nil {
		return fmt.Errorf("error creating policy group linkage %s: %v", id, err)
	}

	d.SetId(id)
	return resourceScalrPolicyGroupLinkageRead(d, meta)
}

func resourceScalrPolicyGroupLinkageRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	policyGroup, environment, err := getLinkedResources(id, scalrClient)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			log.Printf("[DEBUG] Policy group linkage %s not found", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error retrieving policy group linkage %s: %v", id, err)
	}

	d.Set("policy_group_id", policyGroup.ID)
	d.Set("environment_id", environment.ID)

	return nil
}

func resourceScalrPolicyGroupLinkageDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	policyGroup, environment, err := getLinkedResources(id, scalrClient)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			log.Printf("[DEBUG] Policy group linkage %s not found", id)
			return nil
		}
		return fmt.Errorf("error deleting policy group linkage %s: %v", id, err)
	}

	// existing policy groups of the environment that will remain linked
	var policyGroups []*scalr.PolicyGroup
	for _, pg := range environment.PolicyGroups {
		if pg.ID != policyGroup.ID {
			policyGroups = append(policyGroups, pg)
		}
	}

	opts := scalr.EnvironmentUpdateOptions{PolicyGroups: policyGroups}
	_, err = scalrClient.Environments.Update(ctx, environment.ID, opts)
	if err != nil {
		return fmt.Errorf("error deleting policy group linkage %s: %v", id, err)
	}

	return nil
}

// getLinkedResources verifies existence of the linkage
// and returns associated policy group and environment.
func getLinkedResources(id string, scalrClient *scalr.Client) (
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
		return nil, nil, scalr.ErrResourceNotFound{}
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
