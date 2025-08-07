package provider

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrAgentPool() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the state of agent pools in Scalr. Create, update and destroy.",
		CreateContext: resourceScalrAgentPoolCreate,
		ReadContext:   resourceScalrAgentPoolRead,
		UpdateContext: resourceScalrAgentPoolUpdate,
		DeleteContext: resourceScalrAgentPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the agent pool.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"account_id": {
				Description: "ID of the account.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Deprecated: "Attribute `account_id` is deprecated, the account id is calculated from the " +
					"API request context.",
			},
			"environment_id": {
				Description: "ID of the environment.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Deprecated:  "The attribute `environment_id` is deprecated.",
			},
			"vcs_enabled": {
				Description: "Indicates whether the VCS support is enabled for agents in the pool.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"api_gateway_url": {
				Description:      "HTTP(s) destination URL for pool webhook.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},
			"header": {
				Description: "Additional headers to set in the agent pool webhook request.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description:  "The name of the header.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"value": {
							Description:  "The value of the header.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"sensitive": {
							Description: "Whether the header value is a secret.",
							Type:        schema.TypeBool,
							Optional:    true,
						},
					},
				},
			},
			"environments": {
				Description: "The list of the environment identifiers that the agent pool is shared to. Use `[\"*\"]` to share with all environments.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				DefaultFunc: func() (interface{}, error) {
					return []string{"*"}, nil
				},
			},
		},
	}
}

func parsePoolHeaders(d *schema.ResourceData) []*scalr.AgentPoolHeader {
	headers := d.Get("header").(*schema.Set)
	headerValues := make([]*scalr.AgentPoolHeader, 0)
	for _, headerI := range headers.List() {
		header := headerI.(map[string]interface{})
		headerValues = append(headerValues, &scalr.AgentPoolHeader{
			Name:      header["name"].(string),
			Value:     header["value"].(string),
			Sensitive: header["sensitive"].(bool),
		})
	}
	return headerValues
}

func resourceScalrAgentPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	var envID string

	// Get required options
	name := d.Get("name").(string)
	vcsEnabled := d.Get("vcs_enabled").(bool)

	// Create a new options struct
	options := scalr.AgentPoolCreateOptions{
		Name:       ptr(name),
		VcsEnabled: ptr(vcsEnabled),
	}

	if v, ok := d.GetOk("environment_id"); ok {
		envID = v.(string)
		options.Environment = &scalr.Environment{
			ID: v.(string),
		}
	}

	if v, ok := d.GetOk("api_gateway_url"); ok {
		options.WebhookUrl = ptr(v.(string))
		options.WebhookEnabled = ptr(true)
	}

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			options.IsShared = ptr(true)
		} else if len(environments) > 0 {
			environmentValues := make([]*scalr.Environment, 0)
			for _, env := range environments {
				if env.(string) == "*" {
					return diag.Errorf(
						"You cannot simultaneously enable the agent poool for all and a limited list of environments. Please remove either wildcard or environment identifiers.",
					)
				}
				environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
			}
			options.Environments = environmentValues
			options.IsShared = ptr(false)
			if envID != "" {
				return diag.Errorf(
					"Environmnet scope agent pool cannot have environments linkage.",
				)
			}
		}
	}

	if _, ok := d.GetOk("header"); ok {
		options.WebhookHeaders = parsePoolHeaders(d)
	}

	log.Printf("[DEBUG] Creating agent pool %s. Environment: %s", name, envID)
	agentPool, err := scalrClient.AgentPools.Create(ctx, options)
	if err != nil {
		return diag.Errorf(
			"Error creating agent pool %s. Environment %s: %v", name, envID, err)
	}
	d.SetId(agentPool.ID)
	return resourceScalrAgentPoolRead(ctx, d, meta)
}

func resourceScalrAgentPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()
	log.Printf("[DEBUG] Read configuration of agent pool: %s", id)
	agentPool, err := scalrClient.AgentPools.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] agent pool %s not found", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading configuration of agent pool %s: %v", id, err)
	}

	// Update the config.
	_ = d.Set("name", agentPool.Name)
	_ = d.Set("account_id", agentPool.Account.ID)
	_ = d.Set("vcs_enabled", agentPool.VcsEnabled)

	if agentPool.Environment != nil {
		_ = d.Set("environment_id", agentPool.Environment.ID)
	} else {
		_ = d.Set("environment_id", nil)
	}

	if agentPool.IsShared {
		_ = d.Set("environments", []string{"*"})
	} else {
		environmentIDs := make([]string, 0)
		for _, environment := range agentPool.Environments {
			environmentIDs = append(environmentIDs, environment.ID)
		}
		_ = d.Set("environments", environmentIDs)
	}

	if agentPool.WebhookEnabled {
		_ = d.Set("api_gateway_url", agentPool.WebhookUrl)
		headers := make([]map[string]interface{}, 0)
		if agentPool.WebhookHeaders != nil {
			_, doesConfigHasHeaders := d.GetOk("header")
			for _, header := range agentPool.WebhookHeaders {
				if header.Sensitive && doesConfigHasHeaders {
					for _, headerI := range d.Get("header").(*schema.Set).List() {
						configHeader := headerI.(map[string]interface{})
						if header.Name == configHeader["name"] {
							header.Value = configHeader["value"].(string)
						}
					}
				}

				headers = append(headers, map[string]interface{}{
					"name":      header.Name,
					"value":     header.Value,
					"sensitive": header.Sensitive,
				})
			}
		}
		_ = d.Set("header", headers)
	}

	return nil
}

func resourceScalrAgentPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("vcs_enabled") {
		return diag.Errorf("Error updating agentPool %s: %v", id, "vcs_enabled attribute is readonly.")
	}

	options := scalr.AgentPoolUpdateOptions{
		Name: ptr(d.Get("name").(string)),
	}

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			options.IsShared = ptr(true)
			options.Environments = make([]*scalr.Environment, 0)
		} else {
			options.IsShared = ptr(false)
			environmentValues := make([]*scalr.Environment, 0)
			for _, env := range environments {
				if env.(string) == "*" {
					return diag.Errorf(
						"You cannot simultaneously enable the agent pool for all and a limited list of environments. Please remove either wildcard or environment identifiers.",
					)
				}
				environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
			}
			options.Environments = environmentValues
		}

		if _, ok := d.GetOk("environment_id"); ok {
			return diag.Errorf(
				"Environmnet scope agent pool cannot have environments linkage.",
			)
		}

	} else {
		options.IsShared = ptr(false)
		options.Environments = make([]*scalr.Environment, 0)
	}

	if d.HasChange("api_gateway_url") {
		options.WebhookUrl = ptr(d.Get("api_gateway_url").(string))
		options.WebhookEnabled = ptr(true)
	}

	if d.HasChange("header") {
		options.WebhookHeaders = parsePoolHeaders(d)
	}

	log.Printf("[DEBUG] Update agent pool %s", id)
	_, err := scalrClient.AgentPools.Update(ctx, id, options)
	if err != nil {
		return diag.Errorf(
			"Error updating agentPool %s: %v", id, err)
	}

	return resourceScalrAgentPoolRead(ctx, d, meta)
}

func resourceScalrAgentPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete agent pool %s", id)
	err := scalrClient.AgentPools.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf(
			"Error deleting agent pool %s: %v", id, err)
	}

	return nil
}
