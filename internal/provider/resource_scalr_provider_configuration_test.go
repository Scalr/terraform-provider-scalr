package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/client"
)

func TestAccProviderConfiguration_import(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationCustomImportConfig(rName),
			},
			{
				ResourceName:      "scalr_provider_configuration.kubernetes",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccProviderConfiguration_custom(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationCustomConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.kubernetes", &providerConfiguration),
					testAccCheckProviderConfigurationCustomValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "scalr.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.provider_name", "kubernetes"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"scalr_provider_configuration.kubernetes",
						"custom.0.argument.*",
						map[string]string{
							"name":        "host",
							"sensitive":   "false",
							"description": "",
							"value":       "my-host",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"scalr_provider_configuration.kubernetes",
						"custom.0.argument.*",
						map[string]string{
							"name":        "client_certificate",
							"sensitive":   "true",
							"description": "",
							"value":       "-----BEGIN CERTIFICATE-----\nMIIB9TCCAWACAQAwgbgxG",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"scalr_provider_configuration.kubernetes",
						"custom.0.argument.*",
						map[string]string{
							"name":        "config_path",
							"sensitive":   "false",
							"description": "A path to a kube config file. some typo...",
							"value":       "~/.kube/config",
						},
					),
					resource.TestCheckResourceAttr(
						"scalr_provider_configuration.kubernetes", "owners.#", "0",
					),
				),
			},
			{
				Config: testAccScalrProviderConfigurationCustomConfigUpdated(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.kubernetes", &providerConfiguration),
					testAccCheckProviderConfigurationCustomUpdatedValues(&providerConfiguration, rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "scalr.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.provider_name", "kubernetes"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"scalr_provider_configuration.kubernetes",
						"custom.0.argument.*",
						map[string]string{
							"name":        "host",
							"sensitive":   "false",
							"description": "",
							"value":       "my-host",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"scalr_provider_configuration.kubernetes",
						"custom.0.argument.*",
						map[string]string{
							"name":        "config_path",
							"description": "A path to a kube config file.",
							"sensitive":   "true",
							"value":       "~/.kube/config",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"scalr_provider_configuration.kubernetes",
						"custom.0.argument.*",
						map[string]string{
							"name":        "username",
							"sensitive":   "false",
							"description": "",
							"value":       "my-username",
						},
					),
					resource.TestCheckResourceAttr(
						"scalr_provider_configuration.kubernetes", "owners.#", "1",
					),
				),
			},
			{
				// not shared to shared check
				Config: testAccScalrProviderConfigurationCustomConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.kubernetes", &providerConfiguration),
					testAccCheckProviderConfigurationCustomValues(&providerConfiguration, rName),
				),
			},
			{
				Config:      testAccScalrProviderConfigurationCustomWithAwsAttrConfig(rName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Provider type can't be changed."),
			},
		},
	})
}

func TestAccProviderConfiguration_aws_custom(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationCustomConfigAws(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.custom_aws", &providerConfiguration),
					testAccCheckProviderConfigurationCustomAwsValues(&providerConfiguration, rName),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_aws(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	accessKeyId, secretAccessKey, roleArn, externalId := getAwsTestingCreds(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationAwsConfig(rName, accessKeyId, secretAccessKey, roleArn, externalId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.aws", &providerConfiguration),
					testAccCheckProviderConfigurationAwsValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "scalr.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.credentials_type", "role_delegation"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.account_type", "regular"),
				),
			},
			{
				Config: testAccScalrProviderConfigurationAwsUpdatedConfig(rNewName, accessKeyId, secretAccessKey, roleArn, externalId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.aws", &providerConfiguration),
					testAccCheckProviderConfigurationAwsUpdatedValues(&providerConfiguration, rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "export_shell_variables", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "scalr.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.credentials_type", "role_delegation"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.account_type", "gov-cloud"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_scalr(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	scalrHostname := os.Getenv(client.HostnameEnvVar)
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationScalrConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.scalr", &providerConfiguration),
					testAccCheckProviderConfigurationScalrValues(&providerConfiguration, rName, scalrHostname),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "scalr.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "scalr.0.hostname", scalrHostname),
				),
			},
			{
				Config: testAccScalrProviderConfigurationScalrUpdatedConfig(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.scalr", &providerConfiguration),
					testAccCheckProviderConfigurationScalrUpdatedValues(&providerConfiguration, rNewName, scalrHostname),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "export_shell_variables", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "scalr.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.scalr", "scalr.0.hostname", scalrHostname+"/"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_google(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	credentials, project := getGoogleTestingCreds(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationGoogleConfig(rName, credentials, project),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.google", &providerConfiguration),
					testAccCheckProviderConfigurationGoogleValues(&providerConfiguration, rName, project),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.project", project),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.auth_type", "service-account-key"),
				),
			},
			{
				Config: testAccScalrProviderConfigurationGoogleUpdatedConfig(rNewName, credentials, project),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.google", &providerConfiguration),
					testAccCheckProviderConfigurationGoogleUpdatedValues(&providerConfiguration, rNewName, project),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "export_shell_variables", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.project", project),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.auth_type", "service-account-key"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "scalr.#", "0"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_google_oidc(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	_, project := getGoogleTestingCreds(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationGoogleOidcConfig(rName, project),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.google", &providerConfiguration),
					testAccCheckProviderConfigurationGoogleValues(&providerConfiguration, rName, project),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.project", project),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.auth_type", "oidc"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.service_account_email", "test-oidc@example.com"),
				),
			},
			{
				Config: testAccScalrProviderConfigurationGoogleOidcUpdatedConfig(rNewName, project),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.google", &providerConfiguration),
					testAccCheckProviderConfigurationGoogleUpdatedValues(&providerConfiguration, rNewName, project),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "export_shell_variables", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.project", project),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "scalr.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.auth_type", "oidc"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.service_account_email", "changed-test-oidc@example.com"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_aws_oidc(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationAWSOidcConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.aws", &providerConfiguration),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.role_arn", "arn:aws:iam::123456789012:role/scalr-oidc-role"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.credentials_type", "oidc"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.audience", "aws.scalr-run-workload"),
				),
			},
			{
				Config: testAccScalrProviderConfigurationAWSOidcUpdatedConfig(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.aws", &providerConfiguration),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.role_arn", "arn:aws:iam::123456789012:role/scalr-oidc-role2"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.credentials_type", "oidc"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.audience", "aws.scalr-run-workload2"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_azurerm(t *testing.T) {
	if true {
		t.Skip("TODO: add a valid credentials for azurerm testing.")
	}
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	armClientId, armClientSecret, armSubscription, armTenantId := getAzureTestingCreds(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationAzurermConfig(rName, armClientId, armClientSecret, armSubscription, armTenantId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.azurerm", &providerConfiguration),
					testAccCheckProviderConfigurationAzurermValues(&providerConfiguration, rName, armClientId, armSubscription, armTenantId),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "scalr.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.client_id", armClientId),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.subscription_id", armSubscription),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.tenant_id", armTenantId),
				),
			},
			{
				Config: testAccScalrProviderConfigurationAzurermUpdatedConfig(rNewName, armClientId, armClientSecret, armSubscription, armTenantId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.azurerm", &providerConfiguration),
					testAccCheckProviderConfigurationAzurermUpdatedValues(&providerConfiguration, rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "export_shell_variables", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "scalr.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.client_id", armClientId),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.subscription_id", armSubscription),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.tenant_id", armTenantId),
				),
			},
		},
	})
}

func testAccCheckProviderConfigurationCustomValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ProviderName != "kubernetes" {
			return fmt.Errorf("bad provider type, expected \"%s\", got: %#v", "kubernetes", providerConfiguration.ProviderName)
		}
		if providerConfiguration.IsShared != true {
			return fmt.Errorf("bad `is shared`, expected \"%t\", got: %#v", true, providerConfiguration.IsShared)
		}
		expectedArguments := []scalr.ProviderConfigurationParameter{
			{Key: "config_path", Sensitive: false, Value: "~/.kube/config", Description: "A path to a kube config file. some typo..."},
			{Key: "client_certificate", Sensitive: true, Value: ""},
			{Key: "host", Sensitive: false, Value: "my-host"},
		}
		receivedArguments := make(map[string]scalr.ProviderConfigurationParameter)
		for _, receivedArgument := range providerConfiguration.Parameters {
			receivedArguments[receivedArgument.Key] = *receivedArgument
		}
		for _, expectedArgument := range expectedArguments {
			receivedArgument, ok := receivedArguments[expectedArgument.Key]
			if !ok {
				return fmt.Errorf("argument \"%s\" not found", expectedArgument.Key)
			} else if expectedArgument.Sensitive != receivedArgument.Sensitive {
				return fmt.Errorf("argument \"%s\" bad Sensitive, expected \"%t\", got: \"%t\"", expectedArgument.Key, expectedArgument.Sensitive, receivedArgument.Sensitive)
			} else if !receivedArgument.Sensitive && expectedArgument.Value != receivedArgument.Value {
				return fmt.Errorf("argument \"%s\" bad Value, expected \"%s\", got: \"%s\"", expectedArgument.Key, expectedArgument.Value, receivedArgument.Value)
			} else if expectedArgument.Description != receivedArgument.Description {
				return fmt.Errorf("argument \"%s\" bad Description, expected \"%s\", got: \"%s\"", expectedArgument.Key, expectedArgument.Description, receivedArgument.Description)
			}
		}
		return nil
	}
}

func testAccCheckProviderConfigurationCustomUpdatedValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ExportShellVariables != false {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.IsShared != false {
			return fmt.Errorf("bad `is shared`, expected \"%t\", got: %#v", false, providerConfiguration.IsShared)
		}
		if len(providerConfiguration.Environments) != 1 {
			return fmt.Errorf("bad `environments`, expected len \"%d\", got: %#v", 1, len(providerConfiguration.Environments))
		}
		expectedArguments := []scalr.ProviderConfigurationParameter{
			{Key: "config_path", Sensitive: true, Value: "", Description: "A path to a kube config file."},
			{Key: "host", Sensitive: false, Value: "my-host"},
			{Key: "username", Sensitive: false, Value: "my-username"},
		}
		receivedArguments := make(map[string]scalr.ProviderConfigurationParameter)
		for _, receivedArgument := range providerConfiguration.Parameters {
			receivedArguments[receivedArgument.Key] = *receivedArgument
		}
		for _, expectedArgument := range expectedArguments {
			receivedArgument, ok := receivedArguments[expectedArgument.Key]
			if !ok {
				return fmt.Errorf("argument \"%s\" not found", expectedArgument.Key)
			} else if expectedArgument.Sensitive != receivedArgument.Sensitive {
				return fmt.Errorf("argument \"%s\" bad Sensitive, expected \"%t\", got: \"%t\"", expectedArgument.Key, expectedArgument.Sensitive, receivedArgument.Sensitive)
			} else if !receivedArgument.Sensitive && expectedArgument.Value != receivedArgument.Value {
				return fmt.Errorf("argument \"%s\" bad Value, expected \"%s\", got: \"%s\"", expectedArgument.Key, expectedArgument.Value, receivedArgument.Value)
			} else if expectedArgument.Description != receivedArgument.Description {
				return fmt.Errorf("argument \"%s\" bad Description, expected \"%s\", got: \"%s\"", expectedArgument.Key, expectedArgument.Description, receivedArgument.Description)
			}
		}
		return nil
	}
}

func testAccCheckProviderConfigurationAwsValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ProviderName != "aws" {
			return fmt.Errorf("bad provider type, expected \"%s\", got: %#v", "aws", providerConfiguration.ProviderName)
		}
		if providerConfiguration.ExportShellVariables != false {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.AwsCredentialsType != "role_delegation" {
			return fmt.Errorf("bad aws credentials type, expected \"%s\", got: %#v", "role_delegation", providerConfiguration.AwsCredentialsType)
		}
		if providerConfiguration.AwsAccountType != "regular" {
			return fmt.Errorf("bad aws account type, expected \"%s\", got: %#v", "regular", providerConfiguration.AwsAccountType)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationCustomAwsValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ProviderName != "aws" {
			return fmt.Errorf("bad provider type, expected \"%s\", got: %#v", "aws", providerConfiguration.ProviderName)
		}
		if !providerConfiguration.IsCustom {
			return fmt.Errorf("bad is-custom attr, expected \"%s\", got: %#v", "aws", providerConfiguration.IsCustom)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationAwsUpdatedValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ExportShellVariables != true {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", true, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.AwsCredentialsType != "role_delegation" {
			return fmt.Errorf("bad aws credentials type, expected \"%s\", got: %#v", "role_delegation", providerConfiguration.AwsCredentialsType)
		}
		if providerConfiguration.AwsAccountType != "gov-cloud" {
			return fmt.Errorf("bad aws account type, expected \"%s\", got: %#v", "gov-cloud", providerConfiguration.AwsAccountType)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationScalrValues(providerConfiguration *scalr.ProviderConfiguration, name string, scalrHostname string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ProviderName != "scalr" {
			return fmt.Errorf("bad provider type, expected \"%s\", got: %#v", "scalr", providerConfiguration.ProviderName)
		}
		if providerConfiguration.ExportShellVariables != false {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.ScalrHostname != scalrHostname {
			return fmt.Errorf("bad scalr hostname, expected \"%s\", got: %#v", scalrHostname, providerConfiguration.ScalrHostname)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationScalrUpdatedValues(providerConfiguration *scalr.ProviderConfiguration, name string, scalrHostname string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ExportShellVariables != true {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", true, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.ScalrHostname != scalrHostname+"/" {
			return fmt.Errorf("bad scalr hostname, expected \"%s\", got: %#v", "new.somehost.scalr.com", providerConfiguration.ScalrHostname)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationGoogleValues(providerConfiguration *scalr.ProviderConfiguration, name, project string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ProviderName != "google" {
			return fmt.Errorf("bad provider type, expected \"%s\", got: %#v", "google", providerConfiguration.ProviderName)
		}
		if providerConfiguration.ExportShellVariables != false {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.GoogleProject != project {
			return fmt.Errorf("bad google project, expected \"%s\", got: %#v", project, providerConfiguration.GoogleProject)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationGoogleUpdatedValues(providerConfiguration *scalr.ProviderConfiguration, name, project string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ExportShellVariables != true {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", true, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.GoogleProject != project {
			return fmt.Errorf("bad google project, expected \"%s\", got: %#v", project, providerConfiguration.GoogleProject)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationAzurermValues(providerConfiguration *scalr.ProviderConfiguration, name, armClientId, armSubscription, armTenantId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ProviderName != "azurerm" {
			return fmt.Errorf("bad provider type, expected \"%s\", got: %#v", "azurerm", providerConfiguration.ProviderName)
		}
		if providerConfiguration.ExportShellVariables != false {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.AzurermClientId != armClientId {
			return fmt.Errorf("bad azurerm client id, expected \"%s\", got: %#v", armClientId, providerConfiguration.AzurermClientId)
		}
		if providerConfiguration.AzurermSubscriptionId != armSubscription {
			return fmt.Errorf("bad azurerm subscription id, expected \"%s\", got: %#v", armSubscription, providerConfiguration.AzurermSubscriptionId)
		}
		if providerConfiguration.AzurermTenantId != armTenantId {
			return fmt.Errorf("bad azurerm tenant id, expected \"%s\", got: %#v", armTenantId, providerConfiguration.AzurermTenantId)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationAzurermUpdatedValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ExportShellVariables != true {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationExists(n string, providerConfiguration *scalr.ProviderConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		providerConfigurationResource, err := scalrClient.ProviderConfigurations.Read(ctx, rs.Primary.ID)

		if err != nil {
			return err
		}

		*providerConfiguration = *providerConfigurationResource

		return nil
	}
}

func testAccCheckProviderConfigurationResourceDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_provider_configuration" {
			continue
		}

		_, err := scalrClient.ProviderConfigurations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Provider configuraiton (%s) still exists.", rs.Primary.ID)
		}

		if !strings.Contains(err.Error(), fmt.Sprintf("ProviderConfiguration with ID '%s' not found or user unauthorized", rs.Primary.ID)) {
			return err
		}
	}

	return nil
}
func getAzureTestingCreds(t *testing.T) (armClientId string, armClientSecret string, armSubscription string, armTenantId string) {
	armClientId = os.Getenv("TEST_ARM_CLIENT_ID")
	armClientSecret = os.Getenv("TEST_ARM_CLIENT_SECRET")
	armSubscription = os.Getenv("TEST_ARM_SUBSCRIPTION_ID")
	armTenantId = os.Getenv("TEST_ARM_TENANT_ID")
	if len(armClientId) == 0 ||
		len(armClientSecret) == 0 ||
		len(armSubscription) == 0 ||
		len(armTenantId) == 0 {
		t.Skip("Please set TEST_ARM_CLIENT_ID, TEST_ARM_CLIENT_SECRET, TEST_ARM_SUBSCRIPTION_ID and TEST_ARM_TENANT_ID env variables to run this test.")
	}
	return
}

func getAwsTestingCreds(t *testing.T) (accessKeyId, secretAccessKey, roleArn, externalId string) {
	accessKeyId = os.Getenv("TEST_AWS_ACCESS_KEY")
	secretAccessKey = os.Getenv("TEST_AWS_SECRET_KEY")
	roleArn = os.Getenv("TEST_AWS_ROLE_ARN")
	externalId = os.Getenv("TEST_AWS_EXTERNAL_ID")
	if len(accessKeyId) == 0 ||
		len(secretAccessKey) == 0 ||
		len(roleArn) == 0 ||
		len(externalId) == 0 {
		t.Skip("Please set TEST_AWS_ACCESS_KEY, TEST_AWS_SECRET_KEY, TEST_AWS_ROLE_ARN and TEST_AWS_EXTERNAL_ID env variables to run this test.")
	}
	return
}

func getGoogleTestingCreds(t *testing.T) (credentials, project string) {
	credentials = os.Getenv("TEST_GOOGLE_CREDENTIALS")
	project = os.Getenv("TEST_GOOGLE_PROJECT")
	if len(credentials) == 0 ||
		len(project) == 0 {
		t.Skip("Please set TEST_GOOGLE_CREDENTIALS, TEST_GOOGLE_PROJECT env variables to run this test.")
	}
	return
}

func testAccScalrProviderConfigurationCustomConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name                   = "%s"
  account_id             = "%s"
  environments           = ["*"]
  owners                 = []
  custom {
    provider_name = "kubernetes"
    argument {
      name        = "config_path"
      value       = "~/.kube/config"
      sensitive   = false
      description = "A path to a kube config file. some typo..."
    }
    argument {
      name      = "client_certificate"
      value     = "-----BEGIN CERTIFICATE-----\nMIIB9TCCAWACAQAwgbgxG"
      sensitive = true
    }
    argument {
      name  = "host"
      value = "my-host"
    }
  }
}`, name, defaultAccount)
}

func testAccScalrProviderConfigurationCustomImportConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name                   = "%s"
  account_id             = "%s"
  environments           = ["*"]
  custom {
    provider_name = "kubernetes"
    argument {
      name        = "config_path"
      value       = "~/.kube/config"
      sensitive   = false
      description = "A path to a kube config file. some typo..."
    }
    argument {
      name      = "client_id"
      value     = "ID18021989"
      sensitive = false
    }
    argument {
      name  = "host"
      value = "my-host"
    }
  }
}
`, name, defaultAccount)
}

func testAccScalrProviderConfigurationCustomConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name                    = "test-provider-configuration-env"
  account_id              = "%s"
}
resource "scalr_iam_team" "test" {
	name        = "test-k8s-owner"
	description = "Test team"
	users       = []
}
resource "scalr_provider_configuration" "kubernetes" {
  name         = "%s"
  account_id   = "%s"
  environments = ["${scalr_environment.test.id}"]
  owners      = [scalr_iam_team.test.id]
  custom {
    provider_name = "kubernetes"
    argument {
      name        = "config_path"
      value       = "~/.kube/config"
      sensitive   = true
      description = "A path to a kube config file."
    }
    argument {
      name  = "host"
      value = "my-host"
    }
    argument {
      name  = "username"
      value = "my-username"
    }
  }
}
`, defaultAccount, name, defaultAccount)
}

func testAccScalrProviderConfigurationCustomConfigAws(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name                    = "test-provider-configuration-env"
  account_id              = "%s"
}

resource "scalr_provider_configuration" "custom_aws" {
  name         = "%s"
  account_id   = "%s"
  environments = ["${scalr_environment.test.id}"]
  custom {
    provider_name = "aws"
    argument {
      name        = "region"
      value       = "us-east-1"
      sensitive   = false
    }
  }
}
`, defaultAccount, name, defaultAccount)
}

func testAccScalrProviderConfigurationCustomWithAwsAttrConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name                   = "%s"
  account_id             = "%s"
  export_shell_variables = false
  aws {
    account_type        = "gov-cloud"
    credentials_type    = "access_keys"
    access_key          = "access_key"
    secret_key          = "secret_key"
    trusted_entity_type = "aws_account"
  }
}
`, name, defaultAccount)
}

func testAccScalrProviderConfigurationAwsConfig(name, accessKeyId, secretAccessKey, roleArn, externalId string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "aws" {
  name                   = "%s"
  account_id             = "%s"
  export_shell_variables = false
  aws {
    account_type        = "regular"
    credentials_type    = "role_delegation"
    access_key          = "%s"
    secret_key          = "%s"
    role_arn            = "%s"
    external_id         = "%s"
    trusted_entity_type = "aws_account"
  }
}
`, name, defaultAccount, accessKeyId, secretAccessKey, roleArn, externalId)
}

func testAccScalrProviderConfigurationAwsUpdatedConfig(name, accessKeyId, secretAccessKey, roleArn, externalId string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "aws" {
  name                   = "%s"
  account_id             = "%s"
  export_shell_variables = true
  aws {
    account_type        = "gov-cloud"
    credentials_type    = "role_delegation"
    access_key          = "%s"
    secret_key          = "%s"
    role_arn            = "%s"
    external_id         = "%s"
    trusted_entity_type = "aws_account"
  }
}
`, name, defaultAccount, accessKeyId, secretAccessKey, roleArn, externalId)
}

func testAccScalrProviderConfigurationGoogleOidcConfig(name, project string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "google" {
  name       = "%s"
  account_id = "%s"
  google {
    project     			= "%s"
    auth_type				= "oidc"
	service_account_email	= "test-oidc@example.com"
	workload_provider_name	= "projects/123/locations/global/workloadIdentityPools/testpool/providers/dev"
  }
}
`, name, defaultAccount, project)
}

func testAccScalrProviderConfigurationGoogleOidcUpdatedConfig(name, project string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "google" {
  name       = "%s"
  account_id = "%s"
  export_shell_variables = true
  google {
    project     			= "%s"
    auth_type				= "oidc"
	service_account_email	= "changed-test-oidc@example.com"
	workload_provider_name	= "projects/123/locations/global/workloadIdentityPools/testpool/providers/dev"
  }
}
`, name, defaultAccount, project)
}

func testAccScalrProviderConfigurationAWSOidcConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "aws" {
  name       = "%s"
  account_id = "%s"
  aws {
    credentials_type           = "oidc"
    role_arn                   = "arn:aws:iam::123456789012:role/scalr-oidc-role"
    audience = "aws.scalr-run-workload"
  }
}
`, name, defaultAccount)
}

func testAccScalrProviderConfigurationAWSOidcUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "aws" {
  name       = "%s"
  account_id = "%s"
  aws {
    credentials_type           = "oidc"
    role_arn                   = "arn:aws:iam::123456789012:role/scalr-oidc-role2"
    audience = "aws.scalr-run-workload2"
  }
}
`, name, defaultAccount)
}

func testAccScalrProviderConfigurationGoogleConfig(name, credentials, project string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "google" {
  name       = "%s"
  account_id = "%s"
  google {
    project     = "%s"
    credentials = <<-EOT
%s
EOT
  }
}
`, name, defaultAccount, project, credentials)
}

func testAccScalrProviderConfigurationGoogleUpdatedConfig(name, credentials, project string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "google" {
  name       = "%s"
  account_id = "%s"
  export_shell_variables = true
  google {
    project     = "%s"
    credentials = <<-EOT
%s
EOT
  }
}
`, name, defaultAccount, project, credentials)
}

func testAccScalrProviderConfigurationAzurermConfig(name, armClientId, armClientSecret, armSubscription, armTenantId string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "azurerm" {
 name       = "%s"
 account_id = "%s"
 export_shell_variables = false
 azurerm {
   client_id       = "%s"
   client_secret   = "%s"
   subscription_id = "%s"
   tenant_id       = "%s"
 }
}
`, name, defaultAccount, armClientId, armClientSecret, armSubscription, armTenantId)
}

func testAccScalrProviderConfigurationAzurermUpdatedConfig(name, armClientId, armClientSecret, armSubscription, armTenantId string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "azurerm" {
 name       = "%s"
 account_id = "%s"
 export_shell_variables = true
 azurerm {
   client_id       = "%s"
   client_secret   = "%s"
   subscription_id = "%s"
   tenant_id       = "%s"
 }
}
`, name, defaultAccount, armClientId, armClientSecret, armSubscription, armTenantId)
}

func testAccScalrProviderConfigurationScalrConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "scalr" {
  name       = "%s"
  account_id = "%s"
  scalr {
    hostname = "%s"
    token    = "%s"
  }
}
`, name, defaultAccount, os.Getenv(client.HostnameEnvVar), os.Getenv(client.TokenEnvVar))
}

func testAccScalrProviderConfigurationScalrUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "scalr" {
  name       = "%s"
  account_id = "%s"
  export_shell_variables = true
  scalr {
    hostname = "%s"
    token    = "%s"
  }
}
`, name, defaultAccount, os.Getenv(client.HostnameEnvVar)+"/", os.Getenv(client.TokenEnvVar))
}
