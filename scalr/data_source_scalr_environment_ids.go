package scalr

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
	"strings"
)

func dataSourceScalrEnvironmentIDs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrEnvironmentIDsRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"names": {
				Type:         schema.TypeList,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Optional:     true,
				AtLeastOneOf: []string{"tag_ids"},
			},
			"tag_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"exact_match": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ids": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceScalrEnvironmentIDsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	accountId := d.Get("account_id").(string)
	exact := d.Get("exact_match").(bool)
	var id string
	ids := make(map[string]string, 0)
	options := scalr.EnvironmentListOptions{Account: &accountId}

	names := make(map[string]bool)
	if namesI, ok := d.GetOk("names"); ok {
		for _, name := range namesI.([]interface{}) {
			id += name.(string)
			names[name.(string)] = true
		}
	}

	if tagIDsI, ok := d.GetOk("tag_ids"); ok {
		tagIDs := make([]string, 0)
		for _, t := range tagIDsI.(*schema.Set).List() {
			id += t.(string)
			tagIDs = append(tagIDs, t.(string))
		}
		if len(tagIDs) > 0 {
			options.Tag = scalr.String("in:" + strings.Join(tagIDs, ","))
		}
	}

	for {
		el, err := scalrClient.Environments.List(ctx, options)
		if err != nil {
			return diag.Errorf("Error retrieving environments: %v", err)
		}

		for _, e := range el.Items {
			if len(names) > 0 {
				if names["*"] || (exact && names[e.Name]) || (!exact && matchesPattern(e.Name, names)) {
					ids[e.Name] = e.ID
				}
			} else {
				ids[e.Name] = e.ID
			}
		}

		if el.CurrentPage >= el.TotalPages {
			break
		}
		options.PageNumber = el.NextPage
	}

	_ = d.Set("ids", ids)
	d.SetId(fmt.Sprintf("%s/%d", accountId, schema.HashString(id)))

	return nil
}
