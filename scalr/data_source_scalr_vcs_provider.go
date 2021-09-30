package scalr

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
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
			"account": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"environment": {
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

	name := d.Get("name").(string)
	if name != "" {
		options.Query = &name
	}

	accountId := d.Get("account").(string)
	if name != "" {
		options.Account = &accountId
	}

	envId := d.Get("environment").(string)
	if envId != "" {
		options.Environment = &envId
	}

	vcsType := d.Get("vcs_type").(string)
	if vcsType != "" {
		vcsType := scalr.VcsType(vcsType)
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

	envIds := make([]string, 0)
	for _, env := range vcsProvider.Environments {
		envIds = append(envIds, env.ID)
	}

	// Update the configuration.
	d.Set("vcs_type", vcsProvider.VcsType)
	d.Set("name", vcsProvider.Name)
	d.Set("account", vcsProvider.Account.ID)
	d.Set("environments", envIds)
	d.SetId(vcsProvider.ID)

	return nil
}
