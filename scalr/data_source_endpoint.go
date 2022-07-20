package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrEndpointRead,

		Schema: map[string]*schema.Schema{

			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"max_attempts": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"secret_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func dataSourceScalrEndpointRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the ID
	endpointID := d.Get("id").(string)
	endpointName := d.Get("name").(string)

	if endpointID == "" && endpointName == "" {
		return fmt.Errorf("At least one argument 'id' or 'name' is required, but no definitions was found")
	}

	if endpointID != "" && endpointName != "" {
		return fmt.Errorf("Attributes 'name' and 'id' can not be set at the same time")
	}

	environmentID := d.Get("environment_id").(string)
	var endpoint *scalr.Endpoint
	var err error

	if endpointID != "" {
		log.Printf("[DEBUG] Read endpoint with ID: %s", endpointID)
		endpoint, err = scalrClient.Endpoints.Read(ctx, endpointID)
	} else {
		log.Printf("[DEBUG] Read configuration of endpoint: %s", endpointName)
		options := GetEndpointByNameOptions{
			Name: &endpointName,
		}
		if environmentID != "" {
			options.Environment = &environmentID
		}
		endpoint, err = GetEndpointByName(options, scalrClient)
	}

	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return fmt.Errorf("Could not find endpoint %s: %v", endpointID, err)
		}
		return fmt.Errorf("Error retrieving endpoint: %v", err)
	}

	// Update the config.
	d.Set("name", endpoint.Name)
	d.Set("timeout", endpoint.Timeout)
	d.Set("max_attempts", endpoint.MaxAttempts)
	d.Set("secret_key", endpoint.SecretKey)
	d.Set("url", endpoint.Url)
	if endpoint.Environment != nil {
		d.Set("environment_id", endpoint.Environment.ID)
	}
	d.SetId(endpoint.ID)

	return nil
}
