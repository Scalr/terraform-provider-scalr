package provider

import (
	"context"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const CurrentEnvironmentIdEnvVar = "SCALR_ENVIRONMENT_ID"

func dataSourceScalrCurrentEnvironment() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the identifier of current environment when using Scalr remote backend." +
			"\n\nNo arguments are required. The data source returns ID of the current environment",
		ReadContext: dataSourceScalrCurrentEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The identifier of the account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceScalrCurrentEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//scalrClient := meta.(*scalr.Client)

	envId, ok := os.LookupEnv(CurrentEnvironmentIdEnvVar)
	if !ok {
		log.Printf("[DEBUG] %s not is set", CurrentEnvironmentIdEnvVar)
		return diag.Errorf("Current environmnet is not set. `%s` OS environment variable must be set", CurrentEnvironmentIdEnvVar)
	}

	d.SetId(envId)
	return nil
}
