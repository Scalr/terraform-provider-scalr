package provider

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrAccessPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "This data source is used to retrieve details of a single access policy by id.",

		ReadContext: dataSourceScalrAccessPolicyRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The access policy ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"is_system": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"subject": {
				Description: "Defines the subject of the access policy.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The subject ID, `user-<RANDOM STRING>` for user, `team-<RANDOM STRING>` for team, `sa-<RANDOM STRING>` for service account.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"type": {
							Description: "The subject type, is one of `user`, `team`, or `service_account`.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"scope": {
				Description: "Defines the scope where access policy is applied.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The scope ID, `acc-<RANDOM STRING>` for account, `env-<RANDOM STRING>` for environment, `ws-<RANDOM STRING>` for workspace.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"type": {
							Description: "The scope identity type, is one of `account`, `environment`, or `workspace`.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"role_ids": {
				Description: "The list of the role IDs.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceScalrAccessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Get("id").(string)

	log.Printf("[DEBUG] Read configuration of access policy: %s", id)
	ap, err := scalrClient.AccessPolicies.Read(ctx, id)

	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("AccessPolicy '%s' not found", id)
		}
		return diag.Errorf("Error reading configuration of access policy %s: %v", id, err)
	}

	var subject [1]interface{}
	subjectEl := make(map[string]interface{})

	if ap.User != nil {
		subjectEl["type"] = User
		subjectEl["id"] = ap.User.ID
	} else if ap.Team != nil {
		subjectEl["type"] = Team
		subjectEl["id"] = ap.Team.ID
	} else if ap.ServiceAccount != nil {
		subjectEl["type"] = ServiceAccount
		subjectEl["id"] = ap.ServiceAccount.ID
	} else {
		return diag.Errorf("Unable to extract subject from access policy %s", ap.ID)
	}
	subject[0] = subjectEl
	_ = d.Set("subject", subject)

	var scope [1]interface{}
	scopeEl := make(map[string]interface{})

	if ap.Workspace != nil {
		scopeEl["type"] = Workspace
		scopeEl["id"] = ap.Workspace.ID
	} else if ap.Environment != nil {
		scopeEl["type"] = Environment
		scopeEl["id"] = ap.Environment.ID
	} else if ap.Account != nil {
		scopeEl["type"] = Account
		scopeEl["id"] = ap.Account.ID
	} else {
		return diag.Errorf("Unable to extract scope from access policy %s", ap.ID)
	}
	scope[0] = scopeEl
	_ = d.Set("scope", scope)

	roleIds := make([]interface{}, 0)
	for _, role := range ap.Roles {
		roleIds = append(roleIds, role.ID)
	}

	_ = d.Set("role_ids", roleIds)
	_ = d.Set("is_system", ap.IsSystem)
	d.SetId(ap.ID)

	return nil
}
