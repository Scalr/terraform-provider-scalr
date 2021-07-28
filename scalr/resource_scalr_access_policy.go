package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
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
		Create: resourceScalrAccessPolicyCreate,
		Read:   resourceScalrAccessPolicyRead,
		Update: resourceScalrAccessPolicyUpdate,
		Delete: resourceScalrAccessPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"is_system": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"subject": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
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
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
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
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 128,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceScalrAccessPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	subject := d.Get("subject").([]interface{})[0].(map[string]interface{})
	subjectType := subject["type"].(string)
	subjectId := subject["id"].(string)

	scope := d.Get("scope").([]interface{})[0].(map[string]interface{})
	scopeType := scope["type"].(string)
	scopeId := scope["id"].(string)

	var roles []*scalr.Role
	for _, roleId := range d.Get("role_ids").([]interface{}) {
		roles = append(roles, &scalr.Role{ID: roleId.(string)})
	}

	// Create a new options struct.
	options := scalr.AccessPolicyCreateOptions{Roles: roles}

	switch subjectType {
	case "user":
		options.User = &scalr.User{ID: subjectId}
	case "team":
		options.Team = &scalr.Team{ID: subjectId}
	case "service_account":
		options.ServiceAccount = &scalr.ServiceAccount{ID: subjectId}
	}

	switch scopeType {
	case "workspace":
		options.Workspace = &scalr.Workspace{ID: scopeId}
	case "environment":
		options.Environment = &scalr.Environment{ID: scopeId}
	case "account":
		options.Account = &scalr.Account{ID: scopeId}
	}

	log.Printf("[DEBUG] Create access policy for %s %s on %s %s", subjectType, subjectId, scopeType, scopeId)
	ap, err := scalrClient.AccessPolicies.Create(ctx, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating access policy for %s %s on %s %s", subjectType, subjectId, scopeType, scopeId)
	}
	d.SetId(ap.ID)
	return resourceScalrAccessPolicyRead(d, meta)
}

func resourceScalrAccessPolicyRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Read configuration of access policy: %s", id)
	ap, err := scalrClient.AccessPolicies.Read(ctx, id)

	if err != nil {
		if err == scalr.ErrResourceNotFound {
			log.Printf("[DEBUG] AccessPolicy %s not found", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of access policy %s: %v", id, err)
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
		return fmt.Errorf("Unable to extract subject from access policy %s", ap.ID)
	}
	subject[0] = subjectEl
	d.Set("subject", subject)

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
		return fmt.Errorf("Unable to extract scope from access policy %s", ap.ID)
	}
	scope[0] = scopeEl
	d.Set("scope", scope)

	roleIds := make([]interface{}, 0)
	for _, role := range ap.Roles {
		roleIds = append(roleIds, role.ID)
	}

	d.Set("role_ids", roleIds)
	d.Set("is_system", ap.IsSystem)
	d.SetId(ap.ID)

	return nil
}

func resourceScalrAccessPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("role_ids") {
		var roles []*scalr.Role
		for _, roleId := range d.Get("role_ids").([]interface{}) {
			roles = append(roles, &scalr.Role{ID: roleId.(string)})
		}

		// Create a new options struct.
		options := scalr.AccessPolicyUpdateOptions{Roles: roles}

		log.Printf("[DEBUG] Update access policy %s", id)
		_, err := scalrClient.AccessPolicies.Update(ctx, id, options)
		if err != nil {
			return fmt.Errorf(
				"Error updating access policy %s: %v", id, err)
		}
	}

	return resourceScalrAccessPolicyRead(d, meta)
}

func resourceScalrAccessPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete access policy %s", id)
	err := scalrClient.AccessPolicies.Delete(ctx, id)
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf(
			"Error deleting access policy %s: %v", id, err)
	}

	return nil
}
