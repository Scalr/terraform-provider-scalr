package scalr

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
	"strings"
)

func dataSourceScalrEnvironments() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrEnvironmentsRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tag_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
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
		options.Filter.Name = scalr.String(name.(string))
	}

	if tagIDsI, ok := d.GetOk("tag_ids"); ok {
		tagIDs := make([]string, 0)
		for _, t := range tagIDsI.(*schema.Set).List() {
			id.WriteString(t.(string))
			tagIDs = append(tagIDs, t.(string))
		}
		if len(tagIDs) > 0 {
			options.Filter.Tag = scalr.String("in:" + strings.Join(tagIDs, ","))
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
