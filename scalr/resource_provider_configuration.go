package scalr

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrProviderConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrProviderConfigurationCreate,
		Read:   resourceScalrProviderConfigurationRead,
		Update: resourceScalrProviderConfigurationUpdate,
		Delete: resourceScalrProviderConfigurationDelete,
		CustomizeDiff: customdiff.All(
			func(d *schema.ResourceDiff, meta interface{}) error {
				changedProviderTypes := 0
				providerTypeAttrs := []string{"aws", "google", "azurerm", "custom"}
				for _, providerTypeAttr := range providerTypeAttrs {
					if d.HasChange(providerTypeAttr) {
						changedProviderTypes += 1
					}
				}
				if changedProviderTypes > 1 {
					return fmt.Errorf("Provider type can't be changed.")
				}
				return nil
			},
		),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"export_shell_variables": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"aws": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"google", "azurerm", "custom"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key": { // TODO: required?
							Type:     schema.TypeString,
							Optional: true,
						},
						"secret_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},
			"google": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"aws", "azurerm", "custom"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"credentials": {
							Type:      schema.TypeString,
							Optional:  true, // TODO: required?
							Sensitive: true,
						},
					},
				},
			},
			"azurerm": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"aws", "google", "custom"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": { // TODO: required?
							Type:     schema.TypeString,
							Optional: true,
						},
						"client_secret": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"subscription_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tenant_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"custom": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"aws", "google", "azurerm"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"argument": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"sensitive": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceScalrProviderConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	configurationOptions := scalr.ProviderConfigurationCreateOptions{
		Name:                 scalr.String(name),
		Account:              &scalr.Account{ID: accountID},
		ExportShellVariables: scalr.Bool(d.Get("export_shell_variables").(bool)),
	}
	var createArgumentOptions []scalr.ProviderConfigurationParameterCreateOptions

	if v, ok := d.GetOk("aws"); ok {
		configurationOptions.ProviderType = scalr.String("aws")

		aws := v.([]interface{})[0].(map[string]interface{})
		if access_key, ok := aws["access_key"].(string); ok {
			configurationOptions.AwsAccessKey = scalr.String(access_key)
		}
		if secret_key, ok := aws["secret_key"].(string); ok {
			configurationOptions.AwsSecretKey = scalr.String(secret_key)
		}

	} else if v, ok := d.GetOk("google"); ok {
		configurationOptions.ProviderType = scalr.String("google")

		google := v.([]interface{})[0].(map[string]interface{})
		if project, ok := google["project"].(string); ok {
			configurationOptions.GoogleProject = scalr.String(project)
		}
		if credentials, ok := google["credentials"].(string); ok {
			configurationOptions.GoogleCredentials = scalr.String(credentials)
		}

	} else if v, ok := d.GetOk("azurerm"); ok {
		configurationOptions.ProviderType = scalr.String("azurerm")

		azurerm := v.([]interface{})[0].(map[string]interface{})
		if clientId, ok := azurerm["client_id"].(string); ok {
			configurationOptions.AzurermClientId = scalr.String(clientId)
		}
		if clientSecret, ok := azurerm["client_secret"].(string); ok {
			configurationOptions.AzurermClientSecret = scalr.String(clientSecret)
		}
		if subscriptionId, ok := azurerm["subscription_id"].(string); ok {
			configurationOptions.AzurermSubscriptionId = scalr.String(subscriptionId)
		}
		if tenantId, ok := azurerm["tenant_id"].(string); ok {
			configurationOptions.AzurermTenantId = scalr.String(tenantId)
		}

	} else if v, ok := d.GetOk("custom"); ok {
		custom := v.([]interface{})[0].(map[string]interface{})
		configurationOptions.ProviderType = scalr.String(custom["provider_type"].(string))

		for _, v := range custom["argument"].(*schema.Set).List() {
			argument := v.(map[string]interface{})
			createArgumentOption := scalr.ProviderConfigurationParameterCreateOptions{
				Key: scalr.String(argument["name"].(string)),
			}

			if v, ok := argument["value"]; ok {
				createArgumentOption.Value = scalr.String(v.(string))
			}
			if v, ok := argument["description"]; ok {
				createArgumentOption.Description = scalr.String(v.(string))
			}
			if v, ok := argument["sensitive"]; ok {
				createArgumentOption.Sensitive = scalr.Bool(v.(bool))
			}

			createArgumentOptions = append(createArgumentOptions, createArgumentOption)
		}
	}

	providerConfiguration, err := scalrClient.ProviderConfigurations.Create(ctx, configurationOptions)

	if err != nil {
		return fmt.Errorf(
			"Error creating provider configuration %s for account %s: %v", name, accountID, err)
	}
	d.SetId(providerConfiguration.ID)

	if len(createArgumentOptions) != 0 {
		_, err = scalrClient.ProviderConfigurations.CreateParameters(ctx, providerConfiguration.ID, &createArgumentOptions)
		if err != nil {
			defer scalrClient.ProviderConfigurations.Delete(ctx, providerConfiguration.ID)
			return fmt.Errorf(
				"Error creating provider configuration %s for account %s: %v", name, accountID, err)
		}
	}
	return resourceScalrProviderConfigurationRead(d, meta)
}

func resourceScalrProviderConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	providerConfiguration, err := scalrClient.ProviderConfigurations.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {

			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading provider configuration %s: %v", id, err)
	}

	d.Set("name", providerConfiguration.Name)
	d.Set("account_id", providerConfiguration.Account.ID)
	d.Set("export_shell_variables", providerConfiguration.ExportShellVariables)

	switch providerConfiguration.ProviderType {
	case "aws":
		stateAwsParameters := d.Get("aws").([]interface{})[0].(map[string]interface{})
		stateSecretKey := stateAwsParameters["secret_key"].(string)

		d.Set("aws", []map[string]interface{}{
			{
				"access_key": providerConfiguration.AwsAccessKey,
				"secret_key": stateSecretKey,
			},
		})
	case "google":
		stateGoogleParameters := d.Get("google").([]interface{})[0].(map[string]interface{})
		stateCredentials := stateGoogleParameters["credentials"].(string)

		d.Set("google", []map[string]interface{}{
			{
				"project":     providerConfiguration.GoogleProject,
				"credentials": stateCredentials,
			},
		})
	case "azurerm":
		stateAzurermParameters := d.Get("azurerm").([]interface{})[0].(map[string]interface{})
		stateClientSecret := stateAzurermParameters["client_secret"].(string)

		d.Set("azurerm", []map[string]interface{}{
			{
				"client_id":       providerConfiguration.AzurermClientId,
				"client_secret":   stateClientSecret,
				"subscription_id": providerConfiguration.AzurermSubscriptionId,
				"tenant_id":       providerConfiguration.AzurermTenantId,
			},
		})
	default:
		stateCustom := d.Get("custom").([]interface{})[0].(map[string]interface{})

		stateValues := make(map[string]string)
		for _, v := range stateCustom["argument"].(*schema.Set).List() {
			argument := v.(map[string]interface{})
			if value, ok := argument["value"]; ok {
				stateValues[argument["name"].(string)] = value.(string)
			}
		}

		var currentArguments []map[string]interface{}
		for _, argument := range providerConfiguration.Parameters {
			currentArgument := map[string]interface{}{
				"name":      argument.Key,
				"sensitive": argument.Sensitive,
				"value":     argument.Value,
			}

			if stateValue, ok := stateValues[argument.Key]; argument.Sensitive && ok {
				currentArgument["value"] = stateValue
			}

			currentArguments = append(currentArguments, currentArgument)
		}
		d.Set("custom", []map[string]interface{}{
			{
				"provider_type": providerConfiguration.ProviderType,
				"argument":      currentArguments,
			},
		})
	}
	return nil
}

func resourceScalrProviderConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("aws") || d.HasChange("google") || d.HasChange("azurerm") {
		configurationOptions := scalr.ProviderConfigurationUpdateOptions{
			Name: scalr.String(d.Get("name").(string)),
		}
		if v, ok := d.GetOk("aws"); d.HasChange("aws") && ok {
			aws := v.([]interface{})[0].(map[string]interface{})
			if access_key, ok := aws["access_key"].(string); ok {
				configurationOptions.AwsAccessKey = scalr.String(access_key)
			}
			if secret_key, ok := aws["secret_key"].(string); ok {
				configurationOptions.AwsSecretKey = scalr.String(secret_key)
			}

		} else if v, ok := d.GetOk("google"); d.HasChange("google") && ok {
			google := v.([]interface{})[0].(map[string]interface{})
			if project, ok := google["project"].(string); ok {
				configurationOptions.GoogleProject = scalr.String(project)
			}
			if credentials, ok := google["credentials"].(string); ok {
				configurationOptions.GoogleCredentials = scalr.String(credentials)
			}

		} else if v, ok := d.GetOk("azurerm"); d.HasChange("azurerm") && ok {
			azurerm := v.([]interface{})[0].(map[string]interface{})
			if clientId, ok := azurerm["client_id"].(string); ok {
				configurationOptions.AzurermClientId = scalr.String(clientId)
			}
			if clientSecret, ok := azurerm["client_secret"].(string); ok {
				configurationOptions.AzurermClientSecret = scalr.String(clientSecret)
			}
			if subscriptionId, ok := azurerm["subscription_id"].(string); ok {
				configurationOptions.AzurermSubscriptionId = scalr.String(subscriptionId)
			}
			if tenantId, ok := azurerm["tenant_id"].(string); ok {
				configurationOptions.AzurermTenantId = scalr.String(tenantId)
			}

			_, err := scalrClient.ProviderConfigurations.Update(ctx, id, configurationOptions)
			if err != nil {
				return fmt.Errorf(
					"Error updating provider configuration %s: %v", id, err)
			}

		}
	}

	if v, ok := d.GetOk("custom"); d.HasChange("custom") && ok {
		custom := v.([]interface{})[0].(map[string]interface{})

		err := syncArguments(id, custom, scalrClient)
		if err != nil {
			return fmt.Errorf(
				"Error updating provider configuration %s arguments: %v", id, err)
		}
	}

	return resourceScalrProviderConfigurationRead(d, meta)
}

func syncArguments(providerConfigurationId string, custom map[string]interface{}, client *scalr.Client) error {
	providerType := custom["provider_type"].(string)
	configArgumentsCreateOptions := make(map[string]scalr.ProviderConfigurationParameterCreateOptions)
	for _, configArgument := range custom["argument"].([]map[string]interface{}) {
		name := configArgument["name"].(string)
		parameterCreateOption := scalr.ProviderConfigurationParameterCreateOptions{
			Key: scalr.String(name),
		}
		if v, ok := configArgument["value"]; ok {
			parameterCreateOption.Value = scalr.String(v.(string))
		}
		if v, ok := configArgument["sensitive"]; ok {
			parameterCreateOption.Sensitive = scalr.Bool(v.(bool))
		}
		configArgumentsCreateOptions[name] = parameterCreateOption
	}

	providerConfiguration, err := client.ProviderConfigurations.Read(ctx, providerConfigurationId)
	if err != nil {
		return fmt.Errorf(
			"Error reading provider configuration %s: %v", providerConfigurationId, err)
	}

	if providerType != providerConfiguration.ProviderType {
		return fmt.Errorf(
			"Can't change provider configuration type '%s' to '%s'",
			providerConfiguration.ProviderType,
			providerType,
		)
	}

	currentArguments := make(map[string]scalr.ProviderConfigurationParameter)
	for _, argument := range providerConfiguration.Parameters {
		currentArguments[argument.Key] = *argument
	}

	var toCreate []scalr.ProviderConfigurationParameterCreateOptions
	var toUpdate []scalr.ProviderConfigurationParameterUpdateOptions
	for name, configArgumentCreateOption := range configArgumentsCreateOptions {
		currentArgument, exists := currentArguments[name]
		if !exists {
			toCreate = append(toCreate, configArgumentCreateOption)
		} else if currentArgument.Value != *configArgumentCreateOption.Value || currentArgument.Sensitive != *configArgumentCreateOption.Sensitive {
			toUpdate = append(toUpdate, scalr.ProviderConfigurationParameterUpdateOptions{
				ID:        currentArgument.ID,
				Sensitive: configArgumentCreateOption.Sensitive,
				Value:     configArgumentCreateOption.Value,
			})
		}
	}

	var toDelete []string
	for name, currentArgument := range currentArguments {
		if _, exists := configArgumentsCreateOptions[name]; exists {
			toDelete = append(toDelete, currentArgument.ID)
		}
	}
	_, _, _, err = client.ProviderConfigurations.ChangeParameters(
		ctx,
		providerConfigurationId,
		&toCreate,
		&toUpdate,
		&toDelete,
	)
	return err

}

func resourceScalrProviderConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	err := scalrClient.ProviderConfigurations.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return fmt.Errorf(
			"Error deleting provider configuration %s: %v", id, err)
	}

	return nil
}
