package scalr

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/scalr/go-scalr"
)

func resourceScalrVcsProviderV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vcs_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.Github),
						string(scalr.GithubEnterprise),
						string(scalr.Gitlab),
						string(scalr.GitlabEnterprise),
						string(scalr.BitbucketEnterprise),
					},
					false,
				),
			},
			"token": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalrVcsProviderStateUpgradeV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	rawState["username"] = ""
	return rawState, nil
}
