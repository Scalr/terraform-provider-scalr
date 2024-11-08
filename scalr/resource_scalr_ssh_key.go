package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrSSHKey() *schema.Resource {
	return &schema.Resource{
		Description:   "A resource to manage Scalr SSH keys with options for sharing and environment linkage.",
		CreateContext: resourceScalrSSHKeyCreate,
		ReadContext:   resourceScalrSSHKeyRead,
		UpdateContext: resourceScalrSSHKeyUpdate,
		DeleteContext: resourceScalrSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the SSH key. Must be unique within an account.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"private_key": {
				Description: "The private key for the SSH key.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"account_id": {
				Description: "The account ID to which the SSH key belongs.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"environments": {
				Description: "The environments where the SSH key can be used. Use `["*"]` to share with all environments.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceScalrSSHKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	accountID := d.Get("account_id").(string)
	name := d.Get("name").(string)
	privateKey := d.Get("private_key").(string)

	sshKeyOptions := scalr.SSHKeyCreateOptions{
		Account:    &scalr.Account{ID: accountID},
		Name:       scalr.String(name),
		PrivateKey: scalr.String(privateKey),
	}

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			sshKeyOptions.IsShared = scalr.Bool(true)
		} else if len(environments) > 0 {
			environmentValues := make([]*scalr.Environment, 0)
			for _, env := range environments {
				environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
			}
			sshKeyOptions.Environments = environmentValues
		}
	}

	sshKey, err := scalrClient.SSHKeys.Create(ctx, sshKeyOptions)
	if err != nil {
		return diag.Errorf("Error creating SSH key: %v", err)
	}

	d.SetId(sshKey.ID)
	return resourceScalrSSHKeyRead(ctx, d, meta)
}

func resourceScalrSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	sshKey, err := scalrClient.SSHKeys.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading SSH key: %v", err)
	}

	_ = d.Set("name", sshKey.Name)
	_ = d.Set("account_id", sshKey.Account.ID)

	if sshKey.IsShared {
		allEnvironments := []string{"*"}
		_ = d.Set("environments", allEnvironments)
	} else {
		environmentIDs := make([]string, 0)
		for _, environment := range sshKey.Environments {
			environmentIDs = append(environmentIDs, environment.ID)
		}
		_ = d.Set("environments", environmentIDs)
	}

	return nil
}

func resourceScalrSSHKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	if d.HasChange("name") || d.HasChange("private_key") || d.HasChange("environments") {
		sshKeyUpdateOptions := scalr.SSHKeyUpdateOptions{
			Name:       scalr.String(d.Get("name").(string)),
			PrivateKey: scalr.String(d.Get("private_key").(string)),
		}

		if environmentsI, ok := d.GetOk("environments"); ok {
			environments := environmentsI.(*schema.Set).List()
			if (len(environments) == 1) && (environments[0].(string) == "*") {
				sshKeyUpdateOptions.IsShared = scalr.Bool(true)
				sshKeyUpdateOptions.Environments = make([]*scalr.Environment, 0)
			} else {
				sshKeyUpdateOptions.IsShared = scalr.Bool(false)
				environmentValues := make([]*scalr.Environment, 0)
				for _, env := range environments {
					environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
				}
				sshKeyUpdateOptions.Environments = environmentValues
			}
		} else {
			sshKeyUpdateOptions.IsShared = scalr.Bool(false)
			sshKeyUpdateOptions.Environments = make([]*scalr.Environment, 0)
		}

		_, err := scalrClient.SSHKeys.Update(ctx, id, sshKeyUpdateOptions)
		if err != nil {
			return diag.Errorf("Error updating SSH key: %v", err)
		}
	}

	return resourceScalrSSHKeyRead(ctx, d, meta)
}

func resourceScalrSSHKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	err := scalrClient.SSHKeys.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf(
			"Error deleting SSH key with ID %s: %v", id, err)
	}

	return nil
}
