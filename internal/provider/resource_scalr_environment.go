package provider

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrEnvironment() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the state of environments in Scalr. Creates, updates and destroys.",
		CreateContext: resourceScalrEnvironmentCreate,
		ReadContext:   resourceScalrEnvironmentRead,
		DeleteContext: resourceScalrEnvironmentDelete,
		UpdateContext: resourceScalrEnvironmentUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the environment.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"status": {
				Description: "The status of the environment.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Details of the user that created the environment.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Description: "Username of creator.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"email": {
							Description: "Email address of creator.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"full_name": {
							Description: "Full name of creator.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"account_id": {
				Description: "ID of the environment account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"policy_groups": {
				Description: "List of the environment policy-groups IDs, in the format `pgrp-<RANDOM STRING>`.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"default_provider_configurations": {
				Description: "List of IDs of provider configurations, used in the environment workspaces by default.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"tag_ids": {
				Description: "List of tag IDs associated with the environment.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"remote_backend": {
				Description: "If Scalr exports the remote backend configuration and state storage for your infrastructure management. Disabling this feature will also prevent the ability to perform state locking, which ensures that concurrent operations do not conflict. Additionally, it will disable the capability to initiate CLI-driven runs through Scalr.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"mask_sensitive_output": {
				Description: "Enable masking of the sensitive console output. Defaults to `true`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceScalrEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	options := scalr.EnvironmentCreateOptions{
		Name:    ptr(name),
		Account: &scalr.Account{ID: accountID},
	}
	if remoteBackend, ok := d.GetOkExists("remote_backend"); ok { //nolint:staticcheck
		options.RemoteBackend = ptr(remoteBackend.(bool))
	}
	if maskOutput, ok := d.GetOkExists("mask_sensitive_output"); ok { //nolint:staticcheck
		options.MaskSensitiveOutput = ptr(maskOutput.(bool))
	}

	if defaultProviderConfigurationsI, ok := d.GetOk("default_provider_configurations"); ok {
		defaultProviderConfigurations := defaultProviderConfigurationsI.(*schema.Set).List()
		pcfgValues := make([]*scalr.ProviderConfiguration, 0)
		for _, pcfg := range defaultProviderConfigurations {
			pcfgValues = append(pcfgValues, &scalr.ProviderConfiguration{ID: pcfg.(string)})
		}
		options.DefaultProviderConfigurations = pcfgValues

	}
	if tagIDs, ok := d.GetOk("tag_ids"); ok {
		tagIDsList := tagIDs.(*schema.Set).List()
		tags := make([]*scalr.Tag, len(tagIDsList))
		for i, id := range tagIDsList {
			tags[i] = &scalr.Tag{ID: id.(string)}
		}
		options.Tags = tags
	}

	log.Printf("[DEBUG] Create Environment %s for account: %s", name, accountID)
	environment, err := scalrClient.Environments.Create(ctx, options)
	if err != nil {
		return diag.Errorf(
			"Error creating Environment %s for account %s: %v", name, accountID, err)
	}
	d.SetId(environment.ID)
	return resourceScalrEnvironmentRead(ctx, d, meta)
}

func resourceScalrEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	environmentID := d.Id()

	log.Printf("[DEBUG] Read configuration of environment: %s", environmentID)
	environment, err := scalrClient.Environments.Read(ctx, environmentID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			// If the resource isn't available, the function should set the ID
			// to an empty string so Terraform "destroys" the resource in state.
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading environment %s: %v", environmentID, err)
	}

	// Update the configuration.
	_ = d.Set("name", environment.Name)
	_ = d.Set("account_id", environment.Account.ID)
	_ = d.Set("remote_backend", environment.RemoteBackend)
	_ = d.Set("mask_sensitive_output", environment.MaskSensitiveOutput)
	_ = d.Set("status", environment.Status)

	defaultProviderConfigurations := make([]string, 0)
	for _, providerConfiguration := range environment.DefaultProviderConfigurations {
		defaultProviderConfigurations = append(defaultProviderConfigurations, providerConfiguration.ID)
	}
	_ = d.Set("default_provider_configurations", defaultProviderConfigurations)

	var createdBy []interface{}
	if environment.CreatedBy != nil {
		createdBy = append(createdBy, map[string]interface{}{
			"username":  environment.CreatedBy.Username,
			"email":     environment.CreatedBy.Email,
			"full_name": environment.CreatedBy.FullName,
		})
	}
	_ = d.Set("created_by", createdBy)

	policyGroups := make([]string, 0)
	if environment.PolicyGroups != nil {
		for _, group := range environment.PolicyGroups {
			policyGroups = append(policyGroups, group.ID)
		}
	}
	_ = d.Set("policy_groups", policyGroups)

	var tagIDs []string
	if len(environment.Tags) != 0 {
		for _, tag := range environment.Tags {
			tagIDs = append(tagIDs, tag.ID)
		}
	}
	_ = d.Set("tag_ids", tagIDs)

	return nil
}

func resourceScalrEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	var err error

	// Create a new options struct.
	options := scalr.EnvironmentUpdateOptions{
		Name: ptr(d.Get("name").(string)),
	}

	if maskOutput, ok := d.GetOkExists("mask_sensitive_output"); ok { //nolint:staticcheck
		options.MaskSensitiveOutput = ptr(maskOutput.(bool))
	}

	if defaultProviderConfigurationsI, ok := d.GetOk("default_provider_configurations"); ok {
		defaultProviderConfigurations := defaultProviderConfigurationsI.(*schema.Set).List()
		pcfgValues := make([]*scalr.ProviderConfiguration, 0)
		for _, pcfg := range defaultProviderConfigurations {
			pcfgValues = append(pcfgValues, &scalr.ProviderConfiguration{ID: pcfg.(string)})
		}
		options.DefaultProviderConfigurations = pcfgValues
	} else {
		options.DefaultProviderConfigurations = make([]*scalr.ProviderConfiguration, 0)
	}

	log.Printf("[DEBUG] Update environment: %s", d.Id())
	_, err = scalrClient.Environments.Update(ctx, d.Id(), options)
	if err != nil {
		return diag.Errorf("Error updating environment %s: %v", d.Id(), err)
	}

	if d.HasChange("tag_ids") {
		oldTags, newTags := d.GetChange("tag_ids")
		oldSet := oldTags.(*schema.Set)
		newSet := newTags.(*schema.Set)
		tagsToAdd := InterfaceArrToTagRelationArr(newSet.Difference(oldSet).List())
		tagsToDelete := InterfaceArrToTagRelationArr(oldSet.Difference(newSet).List())

		if len(tagsToAdd) > 0 {
			err := scalrClient.EnvironmentTags.Add(ctx, d.Id(), tagsToAdd)
			if err != nil {
				return diag.Errorf(
					"Error adding tags to environment %s: %v", d.Id(), err)
			}
		}

		if len(tagsToDelete) > 0 {
			err := scalrClient.EnvironmentTags.Delete(ctx, d.Id(), tagsToDelete)
			if err != nil {
				return diag.Errorf(
					"Error deleting tags from environment %s: %v", d.Id(), err)
			}
		}
	}

	return resourceScalrEnvironmentRead(ctx, d, meta)
}

func resourceScalrEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	environmentID := d.Id()

	log.Printf("[DEBUG] Delete environment %s", environmentID)
	err := scalrClient.Environments.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf(
			"Error deleting environment %s: %v", environmentID, err)
	}

	return nil
}
