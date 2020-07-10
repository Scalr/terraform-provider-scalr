package scalr

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	scalr "github.com/scalr/go-scalr"
)

func resourceTFESentinelPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFESentinelPolicyCreate,
		Read:   resourceTFESentinelPolicyRead,
		Update: resourceTFESentinelPolicyUpdate,
		Delete: resourceTFESentinelPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFESentinelPolicyImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"policy": {
				Type:     schema.TypeString,
				Required: true,
			},

			"enforce_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(scalr.EnforcementSoft),
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.EnforcementAdvisory),
						string(scalr.EnforcementHard),
						string(scalr.EnforcementSoft),
					},
					false,
				),
			},
		},
	}
}

func resourceTFESentinelPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create a new options struct.
	options := scalr.PolicyCreateOptions{
		Name: scalr.String(name),
		Enforce: []*scalr.EnforcementOptions{
			{
				Path: scalr.String(name + ".sentinel"),
				Mode: scalr.EnforcementMode(scalr.EnforcementLevel(d.Get("enforce_mode").(string))),
			},
		},
	}

	if desc, ok := d.GetOk("description"); ok {
		options.Description = scalr.String(desc.(string))
	}

	log.Printf("[DEBUG] Create sentinel policy %s for organization: %s", name, organization)
	policy, err := scalrClient.Policies.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating sentinel policy %s for organization %s: %v", name, organization, err)
	}

	d.SetId(policy.ID)

	log.Printf("[DEBUG] Upload sentinel policy %s for organization: %s", name, organization)
	err = scalrClient.Policies.Upload(ctx, policy.ID, []byte(d.Get("policy").(string)))
	if err != nil {
		return fmt.Errorf(
			"Error uploading sentinel policy %s for organization %s: %v", name, organization, err)
	}

	return resourceTFESentinelPolicyRead(d, meta)
}

func resourceTFESentinelPolicyRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Read sentinel policy: %s", d.Id())
	policy, err := scalrClient.Policies.Read(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			log.Printf("[DEBUG] Sentinel policy %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading sentinel policy %s: %v", d.Id(), err)
	}

	// Update the config.
	d.Set("name", policy.Name)
	d.Set("description", policy.Description)

	if len(policy.Enforce) == 1 {
		d.Set("enforce_mode", string(policy.Enforce[0].Mode))
	}

	content, err := scalrClient.Policies.Download(ctx, policy.ID)
	if err != nil {
		return fmt.Errorf("Error downloading sentinel policy %s: %v", d.Id(), err)
	}
	d.Set("policy", string(content))

	return nil
}

func resourceTFESentinelPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	if d.HasChange("description") || d.HasChange("enforce_mode") {
		// Create a new options struct.
		options := scalr.PolicyUpdateOptions{}

		if desc, ok := d.GetOk("description"); ok {
			options.Description = scalr.String(desc.(string))
		}

		if d.HasChange("enforce_mode") {
			options.Enforce = []*scalr.EnforcementOptions{
				{
					Path: scalr.String(d.Get("name").(string) + ".sentinel"),
					Mode: scalr.EnforcementMode(scalr.EnforcementLevel(d.Get("enforce_mode").(string))),
				},
			}
		}

		log.Printf("[DEBUG] Update configuration for sentinel policy: %s", d.Id())
		_, err := scalrClient.Policies.Update(ctx, d.Id(), options)
		if err != nil {
			return fmt.Errorf(
				"Error updating configuration for sentinel policy %s: %v", d.Id(), err)
		}
	}

	if d.HasChange("policy") {
		log.Printf("[DEBUG] Update sentinel policy: %s", d.Id())
		err := scalrClient.Policies.Upload(ctx, d.Id(), []byte(d.Get("policy").(string)))
		if err != nil {
			return fmt.Errorf("Error updating sentinel policy %s: %v", d.Id(), err)
		}

	}

	return resourceTFESentinelPolicyRead(d, meta)
}

func resourceTFESentinelPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete sentinel policy: %s", d.Id())
	err := scalrClient.Policies.Delete(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting sentinel policy %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFESentinelPolicyImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid Sentinel policy import format: %s (expected <ORGANIZATION>/<POLICY ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("organization", s[0])
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
