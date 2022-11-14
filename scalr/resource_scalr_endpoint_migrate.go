package scalr

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceScalrEndpointResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"http_method": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"max_attempts": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"url": {
				Type:     schema.TypeString,
				Required: true,
			},

			"secret_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},

			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceScalrEndpointStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	delete(rawState, "http_method")
	return rawState, nil
}
