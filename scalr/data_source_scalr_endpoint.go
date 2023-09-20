package scalr

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrEndpoint() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of a webhook endpoint.",
		DeprecationMessage: "Datasource `scalr_endpoint` is deprecated, the endpoint information" +
			" is included in the `scalr_webhook` resource.",

		ReadContext: dataSourceScalrEndpointRead,

		Schema: map[string]*schema.Schema{

			"id": {
				Description:  "The endpoint ID, in the format `ep-<RANDOM STRING>`.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},

			"name": {
				Description:  "Name of the endpoint.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"max_attempts": {
				Description: "Max delivery attempts of the payload.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"secret_key": {
				Description: "Secret key to sign the webhook payload.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},

			"url": {
				Description: "Endpoint URL.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"timeout": {
				Description: "Endpoint timeout (in seconds).",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"account_id": {
				Description: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},

			"environment_id": {
				Description: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceScalrEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get IDs
	endpointID := d.Get("id").(string)
	endpointName := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	var endpoint *scalr.Endpoint
	var err error

	log.Printf("[DEBUG] Read endpoint with ID '%s' and name '%s'", endpointID, endpointName)
	if endpointID != "" {
		endpoint, err = scalrClient.Endpoints.Read(ctx, endpointID)
		if err != nil {
			return diag.Errorf("Error retrieving endpoint: %v", err)
		}
		if endpointName != "" && endpointName != endpoint.Name {
			return diag.Errorf("Could not find endpoint with ID '%s' and name '%s'", endpointID, endpointName)
		}
	} else {
		options := GetEndpointByNameOptions{
			Name:    &endpointName,
			Account: &accountID,
		}
		endpoint, err = GetEndpointByName(ctx, options, scalrClient)
		if err != nil {
			return diag.Errorf("Error retrieving endpoint: %v", err)
		}
		if endpointID != "" && endpointID != endpoint.ID {
			return diag.Errorf("Could not find endpoint with ID '%s' and name '%s'", endpointID, endpointName)
		}
	}

	// Update the config.
	_ = d.Set("name", endpoint.Name)
	_ = d.Set("timeout", endpoint.Timeout)
	_ = d.Set("max_attempts", endpoint.MaxAttempts)
	_ = d.Set("secret_key", endpoint.SecretKey)
	_ = d.Set("url", endpoint.Url)
	if endpoint.Environment != nil {
		_ = d.Set("environment_id", endpoint.Environment.ID)
	}
	d.SetId(endpoint.ID)

	return nil
}
