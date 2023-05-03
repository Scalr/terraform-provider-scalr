package scalr

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
	"strings"
)

func dataSourceScalrWorkspaces() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrWorkspacesRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
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

func dataSourceScalrWorkspacesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	accountId := d.Get("account_id").(string)

	options := scalr.WorkspaceListOptions{
		Filter: &scalr.WorkspaceFilter{Account: &accountId},
	}

	id := strings.Builder{} // holds the string to build a unique resource id hash
	id.WriteString(accountId)

	ids := make([]string, 0)

	if env, ok := d.GetOk("environment_id"); ok {
		id.WriteString(env.(string))
		options.Filter.Environment = scalr.String(env.(string))
	}

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
		wl, err := scalrClient.Workspaces.List(ctx, options)
		if err != nil {
			return diag.Errorf("Error retrieving workspaces: %v", err)
		}

		for _, w := range wl.Items {
			ids = append(ids, w.ID)
		}

		if wl.CurrentPage >= wl.TotalPages {
			break
		}
		options.PageNumber = wl.NextPage
	}

	_ = d.Set("ids", ids)
	d.SetId(fmt.Sprintf("%d", schema.HashString(id.String())))

	return nil
}
