package scalr

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrVcsProvider() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrVcsProviderRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vcs_type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"environments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		}}
}

func dataSourceScalrVcsProviderRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	options := scalr.VcsProvidersListOptions{}

	if name, ok := d.GetOk("name"); ok {
		options.Query = scalr.String(name.(string))
	}

	if accountId, ok := d.GetOk("account_id"); ok {
		options.Account = scalr.String(accountId.(string))
	}

	if envId, ok := d.GetOk("environment_id"); ok {
		options.Environment = scalr.String(envId.(string))
	}

	if vcsType, ok := d.GetOk("vcs_type"); ok {
		vcsType := scalr.VcsType(vcsType.(string))
		options.VcsType = &vcsType
	}

	vcsProviders, err := scalrClient.VcsProviders.List(ctx, options)

	if err != nil {
		return fmt.Errorf("Error retrieving vcs provider: %s.", err)
	}

	if vcsProviders.TotalCount > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if vcsProviders.TotalCount == 0 {
		return fmt.Errorf("Could not find vcs provider matching you query.")
	}

	vcsProvider := vcsProviders.Items[0]

	envIds := []string{}
	for _, env := range vcsProvider.Environments {
		envIds = append(envIds, env.ID)
	}

	// Update the configuration.
	if vcsProvider.Account != nil {
		d.Set("account_id", vcsProvider.Account.ID)
	}
	d.Set("environments", envIds)
	d.Set("vcs_type", vcsProvider.VcsType)
	d.Set("name", vcsProvider.Name)
	d.Set("url", vcsProvider.Url)
	d.SetId(vcsProvider.ID)

	return nil
}
