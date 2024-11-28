package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrEnvironments() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a list of environment ids by name or tags.",
		ReadContext: dataSourceScalrEnvironmentsRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Description: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"name": {
				Description: "The query used in a Scalr environment name filter.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tag_ids": {
				Description: "List of tag IDs associated with the environment.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"ids": {
				Description: "The list of environment IDs, in the format [`env-xxxxxxxxxxx`, `env-yyyyyyyyy`].",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
		},
	}
}

func dataSourceScalrEnvironmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	accountId := d.Get("account_id").(string)

	options := scalr.EnvironmentListOptions{
		Filter: &scalr.EnvironmentFilter{Account: &accountId},
	}

	id := strings.Builder{} // holds the string to build a unique resource id hash
	id.WriteString(accountId)

	ids := make([]string, 0)

	if name, ok := d.GetOk("name"); ok {
		id.WriteString(name.(string))
		options.Filter.Name = ptr(name.(string))
	}

	if tagIDsI, ok := d.GetOk("tag_ids"); ok {
		tagIDs := make([]string, 0)
		for _, t := range tagIDsI.(*schema.Set).List() {
			id.WriteString(t.(string))
			tagIDs = append(tagIDs, t.(string))
		}
		if len(tagIDs) > 0 {
			options.Filter.Tag = ptr("in:" + strings.Join(tagIDs, ","))
		}
	}

	for {
		el, err := scalrClient.Environments.List(ctx, options)
		if err != nil {
			return diag.Errorf("Error retrieving environments: %v", err)
		}

		for _, e := range el.Items {
			ids = append(ids, e.ID)
		}

		if el.CurrentPage >= el.TotalPages {
			break
		}
		options.PageNumber = el.NextPage
	}

	_ = d.Set("ids", ids)
	d.SetId(fmt.Sprintf("%d", schema.HashString(id.String())))

	return nil
}
