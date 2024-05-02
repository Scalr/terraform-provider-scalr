package scalr

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

const numParallel = 10

func resourceScalrProviderConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "A provider configuration helps organizations manage provider secrets in a centralized way." +
			" It natively supports the management of the major providers like Scalr, AWS, AzureRM," +
			" and Google Cloud Platform, but also allows registering any custom provider." +
			" Please have a look at the basic usage examples for each provider type.",
		CreateContext: resourceScalrProviderConfigurationCreate,
		ReadContext:   resourceScalrProviderConfigurationRead,
		UpdateContext: resourceScalrProviderConfigurationUpdate,
		DeleteContext: resourceScalrProviderConfigurationDelete,
		CustomizeDiff: customdiff.All(
			func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
				changedProviderNames := 0
				providerNameAttrs := []string{"aws", "google", "azurerm", "scalr", "custom"}
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
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"account_id": {
				Description: "The account that owns the object, specified as an ID.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the Scalr provider configuration. This field is unique for the account.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"export_shell_variables": {
				Description: "Export provider variables into the run environment. This option is available for built-in (Scalr, AWS, AzureRM, Google) providers only.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"environments": {
				Description: "The list of environment identifiers that the provider configuration is shared to. Use `[\"*\"]` to share with all environments.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"aws": {
				Description:  "Settings for the aws provider configuration. Exactly one of the following attributes must be set: `scalr`, `aws`, `google`, `azurerm`, `custom`.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"google", "azurerm", "scalr", "custom"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_type": {
							Description: "The type of AWS account, available options: `regular`, `gov-cloud`, `cn-cloud`.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "regular",
						},
						"credentials_type": {
							Description: "The type of AWS credentials, available options: `access_keys`, `role_delegation`, `oidc`.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"trusted_entity_type": {
							Description: "Trusted entity type, available options: `aws_account`, `aws_service`. This option is required with `role_delegation` credentials type.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"role_arn": {
							Description: "Amazon Resource Name (ARN) of the IAM Role to assume. This option is required with the `role_delegation` and `oidc` credentials type.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"external_id": {
							Description: "External identifier to use when assuming the role. This option is required with `role_delegation` credentials type and `aws_account` trusted entity type.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"access_key": {
							Description: "AWS access key. This option is required with `access_keys` credentials type.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"secret_key": {
							Description: "AWS secret key. This option is required with `access_keys` credentials type.",
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
						},
						"audience": {
							Description: "The value of the `aud` claim for the identity token. This option is required with `oidc` credentials type.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"google": {
				Description:  "Settings for the google provider configuration. Exactly one of the following attributes must be set: `scalr`, `aws`, `google`, `azurerm`, `custom`.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"aws", "azurerm", "scalr", "custom"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_type": {
							Description: "Authentication type, either `service-account-key` (default) or `oidc`.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "service-account-key",
						},
						"project": {
							Description: "The default project to manage resources in. If another project is specified on a resource, it will take precedence.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"use_default_project": {
							Description: "If the project a credential is created in will be used by default.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"credentials": {
							Description: "Service account key file in JSON format, required when `auth_type` is `service-account-key`.",
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
						},
						"service_account_email": {
							Description: "The service account email used to authenticate to GCP, required when `auth_type` is `oidc`.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"workload_provider_name": {
							Description: "The canonical name of the workload identity provider, required when `auth_type` is `oidc`.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"azurerm": {
				Description:  "Settings for the azurerm provider configuration. Exactly one of the following attributes must be set: `scalr`, `aws`, `google`, `azurerm`, `custom`.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"aws", "google", "scalr", "custom"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_type": {
							Description: "Authentication type, either `client-secrets` (default) or `oidc`.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "client-secrets",
						},
						"audience": {
							Description: "The value of the `aud` claim for the identity token. This option is required with `oidc` authentication type.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"client_id": {
							Description: "The Client ID that should be used.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"client_secret": {
							Description: "The Client Secret that should be used, required when `auth_type` is `client-secrets`.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"tenant_id": {
							Description: "The Tenant ID that should be used.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"subscription_id": {
							Description: "The Subscription ID that should be used. If skipped, it must be set as a shell variable in the workspace or as a part of the source configuration.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"scalr": {
				Description:  "Settings for the Scalr provider configuration. Exactly one of the following attributes must be set: `scalr`, `aws`, `google`, `azurerm`, `custom`.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"aws", "google", "azurerm", "custom"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Description: "The Scalr hostname which should be used.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"token": {
							Description: "The Scalr token which should be used.",
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
			"custom": {
				Description:  "Settings for the provider configuration that does not have scalr support as a built-in provider. Exactly one of the following attributes must be set: `scalr`, `aws`, `google`, `azurerm`, `custom`.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"aws", "google", "azurerm", "scalr"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider_name": {
							Description: "The name of a Terraform provider.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"argument": {
							Description: "The provider configuration argument. Multiple instances are allowed per block.",
							Type:        schema.TypeSet,
							Required:    true,
							MinItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Description: "The name of the provider configuration argument.",
										Type:        schema.TypeString,
										Required:    true,
									},
									"value": {
										Description: "The value of the provider configuration argument.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"sensitive": {
										Description: "Set (true/false) to configure as sensitive. Default `false`.",
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
									},
									"description": {
										Description: "The description of the provider configuration argument.",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
					},
				},
			},
			"owners": {
				Description: "The teams, the provider configuration belongs to.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceScalrProviderConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	configurationOptions := scalr.ProviderConfigurationCreateOptions{
		Name:                 scalr.String(name),
		Account:              &scalr.Account{ID: accountID},
		ExportShellVariables: scalr.Bool(d.Get("export_shell_variables").(bool)),
	}

	if owners, ok := d.GetOk("owners"); ok {
		ownerResources := make([]*scalr.Team, 0)
		for _, ownerId := range owners.(*schema.Set).List() {
			ownerResources = append(ownerResources, &scalr.Team{ID: ownerId.(string)})
		}
		configurationOptions.Owners = ownerResources
	}

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			configurationOptions.IsShared = scalr.Bool(true)
		} else if len(environments) > 0 {
			environmentValues := make([]*scalr.Environment, 0)
			for _, env := range environments {
				environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
			}
			configurationOptions.Environments = environmentValues
		}
	}

	var createArgumentOptions []scalr.ProviderConfigurationParameterCreateOptions

	if _, ok := d.GetOk("aws"); ok {
		configurationOptions.ProviderName = scalr.String("aws")

		configurationOptions.AwsAccountType = scalr.String(d.Get("aws.0.account_type").(string))
		configurationOptions.AwsCredentialsType = scalr.String(d.Get("aws.0.credentials_type").(string))

		accessKeyIdI, accessKeyIdExists := d.GetOk("aws.0.access_key")
		accessKeyIdExists = accessKeyIdExists && len(accessKeyIdI.(string)) > 0
		accessSecretKeyI, accessSecretKeyExists := d.GetOk("aws.0.secret_key")
		accessSecretKeyExists = accessSecretKeyExists && len(accessSecretKeyI.(string)) > 0

		if accessKeyIdExists && accessSecretKeyExists {
			configurationOptions.AwsAccessKey = scalr.String(accessKeyIdI.(string))
			configurationOptions.AwsSecretKey = scalr.String(accessSecretKeyI.(string))
		} else if accessKeyIdExists || accessSecretKeyExists {
			return diag.Errorf("'access_key' and 'secret_key' fields can be used only together")
		}

		if *configurationOptions.AwsCredentialsType == "role_delegation" {
			configurationOptions.AwsTrustedEntityType = scalr.String(d.Get("aws.0.trusted_entity_type").(string))
			configurationOptions.AwsRoleArn = scalr.String(d.Get("aws.0.role_arn").(string))
			externalIdI, externalIdExists := d.GetOk("aws.0.external_id")
			if externalIdExists {
				configurationOptions.AwsExternalId = scalr.String(externalIdI.(string))
			}
			if len(*configurationOptions.AwsTrustedEntityType) == 0 {
				return diag.Errorf("'trusted_entity_type' field is required for 'role_delegation' credentials type of aws provider configuration")
			}
			if len(*configurationOptions.AwsRoleArn) == 0 {
				return diag.Errorf("'role_arn' field is required for 'role_delegation' credentials type of aws provider configuration")
			}
			if *configurationOptions.AwsTrustedEntityType == "aws_account" && (!externalIdExists || (len(externalIdI.(string)) == 0)) {
				return diag.Errorf("'external_id' field is required for 'role_delegation' credentials type with 'aws_account' trusted entity type of aws provider configuration")
			}
		} else if *configurationOptions.AwsCredentialsType == "oidc" {
			configurationOptions.AwsRoleArn = scalr.String(d.Get("aws.0.role_arn").(string))
			configurationOptions.AwsAudience = scalr.String(d.Get("aws.0.audience").(string))
			if len(*configurationOptions.AwsRoleArn) == 0 {
				return diag.Errorf("'role_arn' field is required for 'oidc' credentials type of aws provider configuration")
			}
			if len(*configurationOptions.AwsAudience) == 0 {
				return diag.Errorf("'audience' field is required for 'oidc' credentials type of aws provider configuration")
			}
		} else if *configurationOptions.AwsCredentialsType != "access_keys" {
			return diag.Errorf("unknown aws provider configuration credentials type: %s, allowed: 'role_delegation', 'access_keys', 'oidc'", *configurationOptions.AwsCredentialsType)
		} else if !accessKeyIdExists || !accessSecretKeyExists {
			return diag.Errorf("'access_key' and 'secret_key' fields are required for 'access_keys' credentials type of aws provider configuration")
		}

	} else if _, ok := d.GetOk("google"); ok {
		configurationOptions.ProviderName = scalr.String("google")
		configurationOptions.GoogleAuthType = scalr.String(d.Get("google.0.auth_type").(string))
		configurationOptions.GoogleUseDefaultProject = scalr.Bool(d.Get("google.0.use_default_project").(bool))

		googleCredentials, googleCredentialsExists := d.GetOk("google.0.credentials")
		googleCredentialsExists = googleCredentialsExists && len(googleCredentials.(string)) > 0
		serviceAccountEmail, serviceAccountEmailExists := d.GetOk("google.0.service_account_email")
		serviceAccountEmailExists = serviceAccountEmailExists && len(serviceAccountEmail.(string)) > 0
		workloadProviderName, workloadProviderNameExists := d.GetOk("google.0.workload_provider_name")
		workloadProviderNameExists = workloadProviderNameExists && len(workloadProviderName.(string)) > 0

		if *configurationOptions.GoogleAuthType == "service-account-key" {
			if !googleCredentialsExists {
				return diag.Errorf("'credentials' field is required for 'service-account-key' auth type of google provider configuration")
			}
			if serviceAccountEmailExists || workloadProviderNameExists {
				return diag.Errorf("'service_account_email' and 'workload_provider_name' fields of google provider configuration can be used only with 'oidc' auth type")
			}
			configurationOptions.GoogleCredentials = scalr.String(googleCredentials.(string))
		} else if *configurationOptions.GoogleAuthType == "oidc" {
			if !serviceAccountEmailExists || !workloadProviderNameExists {
				return diag.Errorf("'service_account_email' and 'workload_provider_name' fields are required for 'oidc' auth type of google provider configuration")
			}
			if googleCredentialsExists {
				return diag.Errorf("'credentials' field of google provider configuration can be used only with 'service-account-key' auth type")
			}
			configurationOptions.GoogleServiceAccountEmail = scalr.String(serviceAccountEmail.(string))
			configurationOptions.GoogleWorkloadProviderName = scalr.String(workloadProviderName.(string))
		} else {
			return diag.Errorf("unknown google provider configuration auth type: '%s', allowed: 'service-account-key', 'oidc'", *configurationOptions.GoogleAuthType)
		}

		if v, ok := d.GetOk("google.0.project"); ok {
			configurationOptions.GoogleProject = scalr.String(v.(string))
		}

	} else if _, ok := d.GetOk("azurerm"); ok {
		configurationOptions.ProviderName = scalr.String("azurerm")
		configurationOptions.AzurermClientId = scalr.String(d.Get("azurerm.0.client_id").(string))
		configurationOptions.AzurermSubscriptionId = scalr.String(d.Get("azurerm.0.subscription_id").(string))
		configurationOptions.AzurermTenantId = scalr.String(d.Get("azurerm.0.tenant_id").(string))

		authType := d.Get("azurerm.0.auth_type").(string)
		if authType == "oidc" {
			audience, audienceExists := d.GetOk("azurerm.0.audience")
			if !audienceExists {
				return diag.Errorf("'audience' field is required for 'oidc' auth type of azurerm provider configuration")
			}
			configurationOptions.AzurermAudience = scalr.String(audience.(string))
			configurationOptions.AzurermAuthType = scalr.String("oidc")
		} else if authType == "client-secrets" {
			client_secret, secretExists := d.GetOk("azurerm.0.client_secret")
			if !secretExists {
				return diag.Errorf("'client_secret' field is required for 'client-secrets' auth type of azurerm provider configuration")
			}
			configurationOptions.AzurermClientSecret = scalr.String(client_secret.(string))
			configurationOptions.AzurermAuthType = scalr.String("client-secrets")
		} else {
			return diag.Errorf("unknown azurerm provider configuration auth type: '%s', allowed: 'client-secrets', 'oidc'", authType)
		}

	} else if _, ok := d.GetOk("scalr"); ok {
		configurationOptions.ProviderName = scalr.String("scalr")
		configurationOptions.ScalrHostname = scalr.String(d.Get("scalr.0.hostname").(string))
		configurationOptions.ScalrToken = scalr.String(d.Get("scalr.0.token").(string))

	} else if v, ok := d.GetOk("custom"); ok {
		custom := v.([]interface{})[0].(map[string]interface{})
		configurationOptions.ProviderName = scalr.String(custom["provider_name"].(string))
		configurationOptions.IsCustom = scalr.Bool(true)

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
		return diag.Errorf(
			"Error creating provider configuration %s for account %s: %v", name, accountID, err)
	}
	d.SetId(providerConfiguration.ID)

	if len(createArgumentOptions) != 0 {
		_, err = createParameters(ctx, scalrClient, providerConfiguration.ID, &createArgumentOptions)
		if err != nil {
			defer func(ctx context.Context, configurationID string) {
				_ = scalrClient.ProviderConfigurations.Delete(ctx, configurationID)
			}(ctx, providerConfiguration.ID)
			return diag.Errorf(
				"Error creating provider configuration %s for account %s: %v", name, accountID, err)
		}
	}
	return resourceScalrProviderConfigurationRead(ctx, d, meta)
}

func resourceScalrProviderConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	providerConfiguration, err := scalrClient.ProviderConfigurations.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {

			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading provider configuration %s: %v", id, err)
	}

	_ = d.Set("name", providerConfiguration.Name)
	_ = d.Set("account_id", providerConfiguration.Account.ID)
	_ = d.Set("export_shell_variables", providerConfiguration.ExportShellVariables)

	if providerConfiguration.IsShared {
		allEnvironments := []string{"*"}
		_ = d.Set("environments", allEnvironments)
	} else {
		environmentIDs := make([]string, 0)
		for _, environment := range providerConfiguration.Environments {
			environmentIDs = append(environmentIDs, environment.ID)
		}
		_ = d.Set("environments", environmentIDs)
	}

	owners := make([]string, 0)
	for _, owner := range providerConfiguration.Owners {
		owners = append(owners, owner.ID)
	}
	_ = d.Set("owners", owners)

	if providerConfiguration.IsCustom {
		var currentArguments []map[string]interface{}

		if stateCustomI, ok := d.GetOk("custom"); ok {
			stateCustom := stateCustomI.([]interface{})
			if len(stateCustom) > 0 {
				stateCustomMap := stateCustom[0].(map[string]interface{})

				stateValues := make(map[string]string)
				for _, v := range stateCustomMap["argument"].(*schema.Set).List() {
					argument := v.(map[string]interface{})
					if value, ok := argument["value"]; ok {
						stateValues[argument["name"].(string)] = value.(string)
					}
				}

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
			}
		} else {
			for _, argument := range providerConfiguration.Parameters {
				currentArgument := map[string]interface{}{
					"name":        argument.Key,
					"sensitive":   argument.Sensitive,
					"value":       argument.Value,
					"description": argument.Description,
				}
				currentArguments = append(currentArguments, currentArgument)
			}
		}
		_ = d.Set("custom", []map[string]interface{}{
			{
				"provider_name": providerConfiguration.ProviderName,
				"argument":      currentArguments,
			},
		})
	} else {
		switch providerConfiguration.ProviderName {
		case "aws":
			aws := make(map[string]interface{})

			aws["account_type"] = providerConfiguration.AwsAccountType
			aws["credentials_type"] = providerConfiguration.AwsCredentialsType

			if stateSecretKeyI, ok := d.GetOk("aws.0.secret_key"); ok {
				aws["secret_key"] = stateSecretKeyI.(string)
			}

			if len(providerConfiguration.AwsAccessKey) > 0 {
				aws["access_key"] = providerConfiguration.AwsAccessKey
			}
			if len(providerConfiguration.AwsTrustedEntityType) > 0 {
				aws["trusted_entity_type"] = providerConfiguration.AwsTrustedEntityType
			}
			if len(providerConfiguration.AwsRoleArn) > 0 {
				aws["role_arn"] = providerConfiguration.AwsRoleArn
			}
			if len(providerConfiguration.AwsExternalId) > 0 {
				aws["external_id"] = providerConfiguration.AwsExternalId
			}
			if len(providerConfiguration.AwsAudience) > 0 {
				aws["audience"] = providerConfiguration.AwsAudience
			}

			_ = d.Set("aws", []map[string]interface{}{aws})
		case "google":
			google := make(map[string]interface{})

			google["auth_type"] = providerConfiguration.GoogleAuthType
			google["use_default_project"] = providerConfiguration.GoogleUseDefaultProject

			var stateCredentials string
			if stateGoogleParametersI, ok := d.GetOk("google"); ok {
				stateGoogleParameters := stateGoogleParametersI.([]interface{})
				if len(stateGoogleParameters) > 0 {
					stateCredentials = stateGoogleParameters[0].(map[string]interface{})["credentials"].(string)
					google["credentials"] = stateCredentials
				}
			}

			if len(providerConfiguration.GoogleProject) > 0 {
				google["project"] = providerConfiguration.GoogleProject
			}
			if len(providerConfiguration.GoogleServiceAccountEmail) > 0 {
				google["service_account_email"] = providerConfiguration.GoogleServiceAccountEmail
			}
			if len(providerConfiguration.GoogleWorkloadProviderName) > 0 {
				google["workload_provider_name"] = providerConfiguration.GoogleWorkloadProviderName
			}

			_ = d.Set("google", []map[string]interface{}{google})
		case "scalr":
			var stateToken string
			if stateScalrParametersI, ok := d.GetOk("scalr"); ok {
				stateScalrParameters := stateScalrParametersI.([]interface{})
				if len(stateScalrParameters) > 0 {
					stateToken = stateScalrParameters[0].(map[string]interface{})["token"].(string)
				}
			}

			_ = d.Set("scalr", []map[string]interface{}{
				{
					"hostname": providerConfiguration.ScalrHostname,
					"token":    stateToken,
				},
			})

		case "azurerm":
			var stateClientSecret string
			if stateAzurermParametersI, ok := d.GetOk("azurerm"); ok {
				stateAzurermParameters := stateAzurermParametersI.([]interface{})
				if len(stateAzurermParameters) > 0 {
					stateClientSecret = stateAzurermParameters[0].(map[string]interface{})["client_secret"].(string)
				}
			}
			auth_type := "client-secrets"
			if len(providerConfiguration.AzurermAuthType) > 0 {
				auth_type = providerConfiguration.AzurermAuthType
			}

			_ = d.Set("azurerm", []map[string]interface{}{
				{
					"client_id":       providerConfiguration.AzurermClientId,
					"client_secret":   stateClientSecret,
					"subscription_id": providerConfiguration.AzurermSubscriptionId,
					"tenant_id":       providerConfiguration.AzurermTenantId,
					"audience":        providerConfiguration.AzurermAudience,
					"auth_type":       auth_type,
				},
			})
		}
	}
	return nil
}

func resourceScalrProviderConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") ||
		d.HasChange("export_shell_variables") ||
		d.HasChange("aws") ||
		d.HasChange("google") ||
		d.HasChange("azurerm") ||
		d.HasChange("scalr") ||
		d.HasChange("custom") ||
		d.HasChange("environments") ||
		d.HasChange("owners") {
		configurationOptions := scalr.ProviderConfigurationUpdateOptions{
			Name:                 scalr.String(d.Get("name").(string)),
			ExportShellVariables: scalr.Bool(d.Get("export_shell_variables").(bool)),
		}
		if environmentsI, ok := d.GetOk("environments"); ok {
			environments := environmentsI.(*schema.Set).List()
			if (len(environments) == 1) && (environments[0].(string) == "*") {
				configurationOptions.IsShared = scalr.Bool(true)
				configurationOptions.Environments = make([]*scalr.Environment, 0)
			} else {
				configurationOptions.IsShared = scalr.Bool(false)
				environmentValues := make([]*scalr.Environment, 0)
				for _, env := range environments {
					environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
				}
				configurationOptions.Environments = environmentValues
			}
		} else {
			configurationOptions.IsShared = scalr.Bool(false)
			configurationOptions.Environments = make([]*scalr.Environment, 0)
		}

		if _, ok := d.GetOk("aws"); ok {
			configurationOptions.AwsAccountType = scalr.String(d.Get("aws.0.account_type").(string))
			configurationOptions.AwsCredentialsType = scalr.String(d.Get("aws.0.credentials_type").(string))

			accessKeyIdI, accessKeyIdExists := d.GetOk("aws.0.access_key")
			accessKeyIdExists = accessKeyIdExists && len(accessKeyIdI.(string)) > 0
			accessSecretKeyI, accessSecretKeyExists := d.GetOk("aws.0.secret_key")
			accessSecretKeyExists = accessSecretKeyExists && len(accessSecretKeyI.(string)) > 0

			if accessKeyIdExists && accessSecretKeyExists {
				configurationOptions.AwsAccessKey = scalr.String(accessKeyIdI.(string))
				configurationOptions.AwsSecretKey = scalr.String(accessSecretKeyI.(string))
			} else if accessKeyIdExists || accessSecretKeyExists {
				return diag.Errorf("'access_key' and 'secret_key' fields can be used only together")
			}

			if *configurationOptions.AwsCredentialsType == "role_delegation" {
				configurationOptions.AwsTrustedEntityType = scalr.String(d.Get("aws.0.trusted_entity_type").(string))
				configurationOptions.AwsRoleArn = scalr.String(d.Get("aws.0.role_arn").(string))
				externalIdI, externalIdExists := d.GetOk("aws.0.external_id")
				if externalIdExists {
					configurationOptions.AwsExternalId = scalr.String(externalIdI.(string))
				}
				if len(*configurationOptions.AwsTrustedEntityType) == 0 {
					return diag.Errorf("'trusted_entity_type' field is required for 'role_delegation' credentials type of aws provider configuration")
				}
				if len(*configurationOptions.AwsRoleArn) == 0 {
					return diag.Errorf("'role_arn' field is required for 'role_delegation' credentials type of aws provider configuration")
				}
				if *configurationOptions.AwsTrustedEntityType == "aws_account" && (!externalIdExists || (len(externalIdI.(string)) == 0)) {
					return diag.Errorf("'external_id' field is required for 'role_delegation' credentials type with 'aws_account' entity type of aws provider configuration")
				}
			} else if *configurationOptions.AwsCredentialsType == "oidc" {
				configurationOptions.AwsRoleArn = scalr.String(d.Get("aws.0.role_arn").(string))
				configurationOptions.AwsAudience = scalr.String(d.Get("aws.0.audience").(string))
				if len(*configurationOptions.AwsRoleArn) == 0 {
					return diag.Errorf("'role_arn' field is required for 'oidc' credentials type of aws provider configuration")
				}
				if len(*configurationOptions.AwsAudience) == 0 {
					return diag.Errorf("'audience' field is required for 'oidc' credentials type of aws provider configuration")
				}
			} else if *configurationOptions.AwsCredentialsType != "access_keys" {
				return diag.Errorf("unknown aws provider configuration credentials type: %s, allowed: 'role_delegation', 'access_keys', 'oidc'", *configurationOptions.AwsCredentialsType)
			} else if !accessKeyIdExists || !accessSecretKeyExists {
				return diag.Errorf("'access_key' and 'secret_key' fields are required for 'access_keys' credentials type of aws provider configuration")
			}
		} else if _, ok := d.GetOk("google"); ok {
			configurationOptions.GoogleAuthType = scalr.String(d.Get("google.0.auth_type").(string))

			configurationOptions.GoogleUseDefaultProject = scalr.Bool(d.Get("google.0.use_default_project").(bool))
			googleCredentials, googleCredentialsExists := d.GetOk("google.0.credentials")
			googleCredentialsExists = googleCredentialsExists && len(googleCredentials.(string)) > 0
			serviceAccountEmail, serviceAccountEmailExists := d.GetOk("google.0.service_account_email")
			serviceAccountEmailExists = serviceAccountEmailExists && len(serviceAccountEmail.(string)) > 0
			workloadProviderName, workloadProviderNameExists := d.GetOk("google.0.workload_provider_name")
			workloadProviderNameExists = workloadProviderNameExists && len(workloadProviderName.(string)) > 0

			if *configurationOptions.GoogleAuthType == "service-account-key" {
				if !googleCredentialsExists {
					return diag.Errorf("'credentials' field is required for 'service-account-key' auth type of google provider configuration")
				}
				if serviceAccountEmailExists || workloadProviderNameExists {
					return diag.Errorf("'service_account_email' and 'workload_provider_name' fields of google provider configuration can be used only with 'oidc' auth type")
				}
				configurationOptions.GoogleCredentials = scalr.String(googleCredentials.(string))
			} else if *configurationOptions.GoogleAuthType == "oidc" {
				if !serviceAccountEmailExists || !workloadProviderNameExists {
					return diag.Errorf("'service_account_email' and 'workload_provider_name' fields are required for 'oidc' auth type of google provider configuration")
				}
				if googleCredentialsExists {
					return diag.Errorf("'credentials' field of google provider configuration can be used only with 'service-account-key' auth type")
				}
				configurationOptions.GoogleServiceAccountEmail = scalr.String(serviceAccountEmail.(string))
				configurationOptions.GoogleWorkloadProviderName = scalr.String(workloadProviderName.(string))
			} else {
				return diag.Errorf("unknown google provider configuration auth type: '%s', allowed: 'service-account-key', 'oidc'", *configurationOptions.GoogleAuthType)
			}

			if v, ok := d.GetOk("google.0.project"); ok {
				configurationOptions.GoogleProject = scalr.String(v.(string))
			}
		} else if _, ok := d.GetOk("scalr"); ok {
			configurationOptions.ScalrHostname = scalr.String(d.Get("scalr.0.hostname").(string))
			configurationOptions.ScalrToken = scalr.String(d.Get("scalr.0.token").(string))
		} else if _, ok := d.GetOk("azurerm"); ok {
			configurationOptions.AzurermClientId = scalr.String(d.Get("azurerm.0.client_id").(string))
			configurationOptions.AzurermSubscriptionId = scalr.String(d.Get("azurerm.0.subscription_id").(string))
			configurationOptions.AzurermTenantId = scalr.String(d.Get("azurerm.0.tenant_id").(string))

			authType := d.Get("azurerm.0.auth_type").(string)
			if authType == "oidc" {
				audience, audienceExists := d.GetOk("azurerm.0.audience")
				if !audienceExists {
					return diag.Errorf("'audience' field is required for 'oidc' auth type of azurerm provider configuration")
				}
				configurationOptions.AzurermAudience = scalr.String(audience.(string))
				configurationOptions.AzurermAuthType = scalr.String("oidc")
			} else if authType == "client-secrets" {
				client_secret, secretExists := d.GetOk("azurerm.0.client_secret")
				if !secretExists {
					return diag.Errorf("'client_secret' field is required for 'client-secrets' auth type of azurerm provider configuration")
				}
				configurationOptions.AzurermClientSecret = scalr.String(client_secret.(string))
				configurationOptions.AzurermAuthType = scalr.String("client-secrets")
			} else {
				return diag.Errorf("unknown azurerm provider configuration auth type: '%s', allowed: 'client-secrets', 'oidc'", authType)
			}

		}

		ownerResources := make([]*scalr.Team, 0)
		if owners, ok := d.GetOk("owners"); ok {
			for _, ownerId := range owners.(*schema.Set).List() {
				ownerResources = append(ownerResources, &scalr.Team{ID: ownerId.(string)})
			}
		}
		configurationOptions.Owners = ownerResources

		_, err := scalrClient.ProviderConfigurations.Update(ctx, id, configurationOptions)
		if err != nil {
			return diag.Errorf(
				"Error updating provider configuration %s: %v", id, err)
		}
	}

	if v, ok := d.GetOk("custom"); d.HasChange("custom") && ok {
		custom := v.([]interface{})[0].(map[string]interface{})

		err := syncArguments(ctx, id, custom, scalrClient)
		if err != nil {
			return diag.Errorf(
				"Error updating provider configuration %s arguments: %v", id, err)
		}
	}

	return resourceScalrProviderConfigurationRead(ctx, d, meta)
}

func syncArguments(ctx context.Context, providerConfigurationId string, custom map[string]interface{}, client *scalr.Client) error {
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
		if !exists || currentArgument.Sensitive && !(*configArgumentCreateOption.Sensitive) {
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
		configArgumentCreateOption, exists := configArgumentsCreateOptions[name]
		if !exists || currentArgument.Sensitive && !(*configArgumentCreateOption.Sensitive) {
			toDelete = append(toDelete, currentArgument.ID)
		}
	}
	_, _, _, err = changeParameters(
		ctx,
		client,
		providerConfigurationId,
		nil,
		nil,
		&toDelete,
	)
	if err != nil {
		return err
	}
	_, _, _, err = changeParameters(
		ctx,
		client,
		providerConfigurationId,
		&toCreate,
		&toUpdate,
		nil,
	)
	return err

}

func resourceScalrProviderConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	err := scalrClient.ProviderConfigurations.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf(
			"Error deleting provider configuration %s: %v", id, err)
	}

	return nil
}

// changeParameters is used to change parameters for provider configuration.
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
	wg.Add(numParallel)

	resultCh := make(chan result)

	for i := 0; i < numParallel; i++ {
		go func() {
			reqCtx, reqCancel := context.WithTimeout(context.Background(), time.Second*10)
			defer reqCancel()

			for t := range inputCh {
				if t.createOption != nil {
					parameter, err := client.ProviderConfigurationParameters.Create(reqCtx, configurationID, *t.createOption)
					resultCh <- result{created: parameter, err: err}
				} else if t.updateOption != nil {
					parameter, err := client.ProviderConfigurationParameters.Update(reqCtx, t.updateOption.ID, *t.updateOption)
					resultCh <- result{updated: parameter, err: err}
				} else {
					err := client.ProviderConfigurationParameters.Delete(reqCtx, *t.deleteId)
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
