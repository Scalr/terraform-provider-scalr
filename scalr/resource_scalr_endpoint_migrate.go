package scalr

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceScalrEndpointResourceV0() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrEndpointCreate,
		Read:   resourceScalrEndpointRead,
		Update: resourceScalrEndpointUpdate,
		Delete: resourceScalrEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				Required:  true,
				Sensitive: true,
			},

			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceScalrEndpointStateUpgradeV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	return rawState, nil
}

func resourceScalrEndpointResourceV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrEndpointCreate,
		Read:   resourceScalrEndpointRead,
		Update: resourceScalrEndpointUpdate,
		Delete: resourceScalrEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				Required:  true,
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
