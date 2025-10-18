package provider

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

type Scope string
type Subject string

const (
	User           Subject = "user"
	Team           Subject = "team"
	ServiceAccount Subject = "service_account"
)

const (
	Workspace   Scope = "workspace"
	Environment Scope = "environment"
	Account     Scope = "account"
)

func (s Scope) IsValid() error {
	switch s {
	case Workspace, Environment, Account:
		return nil
	}
	return errors.New("Invalid scope type")
}

func (s Subject) IsValid() error {
	switch s {
	case User, Team, ServiceAccount:
		return nil
	}
	return errors.New("Invalid subject type")
}

func resourceScalrAccessPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages the Scalr IAM access policies. Create, update and destroy.",
		CreateContext: resourceScalrAccessPolicyCreate,
		ReadContext:   resourceScalrAccessPolicyRead,
		UpdateContext: resourceScalrAccessPolicyUpdate,
		DeleteContext: resourceScalrAccessPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"is_system": {
				Description: "The access policy is a built-in read-only policy that cannot be updated or deleted.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"subject": {
				Description: "Defines the subject of the access policy.",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The subject ID, `user-<RANDOM STRING>` for user, `team-<RANDOM STRING>` for team, `sa-<RANDOM STRING>` for service account.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"type": {
							Description: "The subject type, is one of `user`, `team`, or `service_account`.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if err := Subject(v).IsValid(); err != nil {
									errs = append(errs, fmt.Errorf("%s must be one of [user, team, service_account], got: %s", key, v))
								}
								return
							},
						},
					},
				},
			},
			"scope": {
				Description: "Defines the scope where access policy is applied.",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The scope ID, `acc-<RANDOM STRING>` for account, `env-<RANDOM STRING>` for environment, `ws-<RANDOM STRING>` for workspace.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"type": {
							Description: "The scope identity type, is one of `account`, `environment`, or `workspace`.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if err := Scope(v).IsValid(); err != nil {
									errs = append(errs, fmt.Errorf("%s must be one of [workspace, environment, account], got: %s", key, v))
								}
								return
							}},
					},
				},
			},
			"role_ids": {
				Description: "The list of the role IDs.",
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				MaxItems:    128,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func parseRoleIdDefinitions(d *schema.ResourceData) ([]*scalr.Role, error) {
	roles := make([]*scalr.Role, 0)

	roleIds := d.Get("role_ids").(*schema.Set).List()
	err := ValidateIDsDefinitions(roleIds)
	if err != nil {
		return nil, fmt.Errorf("Got error during parsing role ids: %s", err.Error())
	}

	for _, roleId := range roleIds {
		roles = append(roles, &scalr.Role{ID: roleId.(string)})
	}

	return roles, nil
}

func resourceScalrAccessPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	subject := d.Get("subject").([]interface{})[0].(map[string]interface{})
	subjectType := subject["type"].(string)
	subjectId := subject["id"].(string)

	scope := d.Get("scope").([]interface{})[0].(map[string]interface{})
	scopeType := scope["type"].(string)
	scopeId := scope["id"].(string)

	roles, err := parseRoleIdDefinitions(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create a new options struct.
	options := scalr.AccessPolicyCreateOptions{Roles: roles}

	switch Subject(subjectType) {
	case User:
		options.User = &scalr.User{ID: subjectId}
	case Team:
		options.Team = &scalr.Team{ID: subjectId}
	case ServiceAccount:
		options.ServiceAccount = &scalr.ServiceAccount{ID: subjectId}
	}

	switch Scope(scopeType) {
	case Workspace:
		options.Workspace = &scalr.Workspace{ID: scopeId}
	case Environment:
		options.Environment = &scalr.Environment{ID: scopeId}
	case Account:
		options.Account = &scalr.Account{ID: scopeId}
	}

	log.Printf("[DEBUG] Create access policy for %s %s on %s %s", subjectType, subjectId, scopeType, scopeId)
	ap, err := scalrClient.AccessPolicies.Create(ctx, options)
	if err != nil {
		return diag.Errorf(
			"Error creating access policy for %s %s on %s %s: %v", subjectType, subjectId, scopeType, scopeId, err)
	}
	d.SetId(ap.ID)
	return resourceScalrAccessPolicyRead(ctx, d, meta)
}

func resourceScalrAccessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Read configuration of access policy: %s", id)
	ap, err := scalrClient.AccessPolicies.Read(ctx, id)

	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] AccessPolicy %s not found", id)
			d.SetId("")
			return nil
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

func resourceScalrAccessPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("role_ids") {
		roles, err := parseRoleIdDefinitions(d)
		if err != nil {
			return diag.FromErr(err)
		}

		// Create a new options struct.
		options := scalr.AccessPolicyUpdateOptions{Roles: roles}

		log.Printf("[DEBUG] Update access policy %s", id)
		_, err = scalrClient.AccessPolicies.Update(ctx, id, options)
		if err != nil {
			return diag.Errorf(
				"Error updating access policy %s: %v", id, err)
		}
	}

	return resourceScalrAccessPolicyRead(ctx, d, meta)
}

func resourceScalrAccessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete access policy %s", id)
	err := scalrClient.AccessPolicies.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf(
			"Error deleting access policy %s: %v", id, err)
	}

	return nil
}
