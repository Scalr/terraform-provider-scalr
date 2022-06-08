package scalr

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

const NUM_PARALLEL = 10

func resourceScalrProviderConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrProviderConfigurationCreate,
		Read:   resourceScalrProviderConfigurationRead,
		Update: resourceScalrProviderConfigurationUpdate,
		Delete: resourceScalrProviderConfigurationDelete,
		CustomizeDiff: customdiff.All(
			func(d *schema.ResourceDiff, meta interface{}) error {
				changedProviderNames := 0
				providerNameAttrs := []string{"aws", "google", "azurerm", "custom"}
				for _, providerNameAttr := range providerNameAttrs {
					if d.HasChange(providerNameAttr) {
						changedProviderNames += 1
					}
				}

				if changedProviderNames > 1 {
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
				Default:  false,
			},
			"aws": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"google", "azurerm", "custom"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key": {
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
							Optional:  true,
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
						"client_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"client_secret": {
							Type:     schema.TypeString,
							Required: true,
						},
						"tenant_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"subscription_id": {
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
						"provider_name": {
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
										Default:  false,
									},
									"description": {
										Type:     schema.TypeString,
										Optional: true,
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

	if _, ok := d.GetOk("aws"); ok {
		configurationOptions.ProviderName = scalr.String("aws")

		if v, ok := d.GetOk("aws.0.access_key"); ok {
			configurationOptions.AwsAccessKey = scalr.String(v.(string))
		}
		if v, ok := d.GetOk("aws.0.secret_key"); ok {
			configurationOptions.AwsSecretKey = scalr.String(v.(string))
		}

	} else if _, ok := d.GetOk("google"); ok {
		configurationOptions.ProviderName = scalr.String("google")

		if v, ok := d.GetOk("google.0.project"); ok {
			configurationOptions.GoogleProject = scalr.String(v.(string))
		}
		if v, ok := d.GetOk("google.0.credentials"); ok {
			configurationOptions.GoogleCredentials = scalr.String(v.(string))
		}

	} else if _, ok := d.GetOk("azurerm"); ok {
		configurationOptions.ProviderName = scalr.String("azurerm")
		configurationOptions.AzurermClientId = scalr.String(d.Get("azurerm.0.client_id").(string))
		configurationOptions.AzurermClientSecret = scalr.String(d.Get("azurerm.0.client_secret").(string))
		configurationOptions.AzurermSubscriptionId = scalr.String(d.Get("azurerm.0.subscription_id").(string))
		if v, ok := d.GetOk("azurerm.0.tenant_id"); ok {
			configurationOptions.AzurermTenantId = scalr.String(v.(string))
		}

	} else if v, ok := d.GetOk("custom"); ok {
		custom := v.([]interface{})[0].(map[string]interface{})
		configurationOptions.ProviderName = scalr.String(custom["provider_name"].(string))

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
		_, err = createParameters(ctx, scalrClient, providerConfiguration.ID, &createArgumentOptions)
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

	switch providerConfiguration.ProviderName {
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
				"name":        argument.Key,
				"sensitive":   argument.Sensitive,
				"value":       argument.Value,
				"description": argument.Description,
			}

			if stateValue, ok := stateValues[argument.Key]; argument.Sensitive && ok {
				currentArgument["value"] = stateValue
			}

			currentArguments = append(currentArguments, currentArgument)
		}
		d.Set("custom", []map[string]interface{}{
			{
				"provider_name": providerConfiguration.ProviderName,
				"argument":      currentArguments,
			},
		})
	}
	return nil
}

func resourceScalrProviderConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("export_shell_variables") || d.HasChange("aws") || d.HasChange("google") || d.HasChange("azurerm") {
		configurationOptions := scalr.ProviderConfigurationUpdateOptions{
			Name:                 scalr.String(d.Get("name").(string)),
			ExportShellVariables: scalr.Bool(d.Get("export_shell_variables").(bool)),
		}

		if d.HasChange("aws") {
			if v, ok := d.GetOk("aws.0.access_key"); ok {
				configurationOptions.AwsAccessKey = scalr.String(v.(string))
			}
			if v, ok := d.GetOk("aws.0.secret_key"); ok {
				configurationOptions.AwsSecretKey = scalr.String(v.(string))
			}
		} else if d.HasChange("google") {
			if v, ok := d.GetOk("google.0.project"); ok {
				configurationOptions.GoogleProject = scalr.String(v.(string))
			}
			if v, ok := d.GetOk("google.0.credentials"); ok {
				configurationOptions.GoogleCredentials = scalr.String(v.(string))
			}
		} else if _, ok := d.GetOk("azurerm"); ok {
			configurationOptions.AzurermClientId = scalr.String(d.Get("azurerm.0.client_id").(string))
			configurationOptions.AzurermClientSecret = scalr.String(d.Get("azurerm.0.client_secret").(string))
			configurationOptions.AzurermSubscriptionId = scalr.String(d.Get("azurerm.0.subscription_id").(string))
			if v, ok := d.GetOk("azurerm.0.tenant_id"); ok {
				configurationOptions.AzurermTenantId = scalr.String(v.(string))
			}
		}
		_, err := scalrClient.ProviderConfigurations.Update(ctx, id, configurationOptions)
		if err != nil {
			return fmt.Errorf(
				"Error updating provider configuration %s: %v", id, err)
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
	providerName := custom["provider_name"].(string)
	configArgumentsCreateOptions := make(map[string]scalr.ProviderConfigurationParameterCreateOptions)
	for _, v := range custom["argument"].(*schema.Set).List() {
		configArgument := v.(map[string]interface{})
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
		if v, ok := configArgument["description"]; ok {
			parameterCreateOption.Description = scalr.String(v.(string))
		}
		configArgumentsCreateOptions[name] = parameterCreateOption
	}

	providerConfiguration, err := client.ProviderConfigurations.Read(ctx, providerConfigurationId)
	if err != nil {
		return fmt.Errorf(
			"Error reading provider configuration %s: %v", providerConfigurationId, err)
	}

	if providerName != providerConfiguration.ProviderName {
		return fmt.Errorf(
			"Can't change provider configuration type '%s' to '%s'",
			providerConfiguration.ProviderName,
			providerName,
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
		} else if currentArgument.Value != *configArgumentCreateOption.Value || currentArgument.Sensitive != *configArgumentCreateOption.Sensitive || currentArgument.Description != *configArgumentCreateOption.Description {
			toUpdate = append(toUpdate, scalr.ProviderConfigurationParameterUpdateOptions{
				ID:          currentArgument.ID,
				Sensitive:   configArgumentCreateOption.Sensitive,
				Value:       configArgumentCreateOption.Value,
				Description: configArgumentCreateOption.Description,
			})
		}
	}

	var toDelete []string
	for name, currentArgument := range currentArguments {
		if _, exists := configArgumentsCreateOptions[name]; !exists {
			toDelete = append(toDelete, currentArgument.ID)
		}
	}
	_, _, _, err = changeParameters(
		ctx,
		client,
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

// changeParameters is used to change parameters for provider configuratio.
func changeParameters(
	ctx context.Context,
	client *scalr.Client,
	configurationID string,
	toCreate *[]scalr.ProviderConfigurationParameterCreateOptions,
	toUpdate *[]scalr.ProviderConfigurationParameterUpdateOptions,
	toDelete *[]string,
) (
	created []scalr.ProviderConfigurationParameter,
	updated []scalr.ProviderConfigurationParameter,
	deleted []string,
	err error,
) {

	done := make(chan struct{})
	defer close(done)

	type result struct {
		created *scalr.ProviderConfigurationParameter
		updated *scalr.ProviderConfigurationParameter
		deleted *string
		err     error
	}
	type task struct {
		createOption *scalr.ProviderConfigurationParameterCreateOptions
		updateOption *scalr.ProviderConfigurationParameterUpdateOptions
		deleteId     *string
	}

	inputCh := make(chan task)
	var tasks []task

	if toDelete != nil {
		for i := range *toDelete {
			tasks = append(tasks, task{deleteId: &(*toDelete)[i]})
		}
	}
	if toUpdate != nil {
		for i := range *toUpdate {
			tasks = append(tasks, task{updateOption: &(*toUpdate)[i]})
		}
	}
	if toCreate != nil {
		for i := range *toCreate {
			tasks = append(tasks, task{createOption: &(*toCreate)[i]})
		}
	}

	if tasks == nil {
		return
	}

	go func() {
		defer close(inputCh)
		for _, t := range tasks {
			select {
			case inputCh <- t:

			case <-done:
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(NUM_PARALLEL)

	resultCh := make(chan result)

	for i := 0; i < NUM_PARALLEL; i++ {
		go func() {
			for t := range inputCh {
				if t.createOption != nil {
					parameter, err := client.ProviderConfigurationParameters.Create(ctx, configurationID, *t.createOption)
					resultCh <- result{created: parameter, err: err}
				} else if t.updateOption != nil {
					parameter, err := client.ProviderConfigurationParameters.Update(ctx, t.updateOption.ID, *t.updateOption)
					resultCh <- result{updated: parameter, err: err}
				} else {
					err := client.ProviderConfigurationParameters.Delete(ctx, *t.deleteId)
					resultCh <- result{deleted: t.deleteId, err: err}
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for result := range resultCh {
		if result.err != nil {
			err = result.err
			break
		} else if result.created != nil {
			created = append(created, *result.created)
		} else if result.updated != nil {
			updated = append(updated, *result.updated)
		} else {
			deleted = append(deleted, *result.deleted)
		}
	}

	return
}

// createParameters is used to create parameters for provider configuratio.
func createParameters(
	ctx context.Context,
	client *scalr.Client,
	configurationID string,
	optionsList *[]scalr.ProviderConfigurationParameterCreateOptions,
) (
	created []scalr.ProviderConfigurationParameter,
	err error,
) {
	created, _, _, err = changeParameters(
		ctx, client, configurationID, optionsList, nil, nil,
	)
	return
}
