package scalr

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	scalr "github.com/scalr/go-scalr"
)

func TestAccProviderConfiguration_custom(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrPorivderConfigurationCustomConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.kubernetes", &providerConfiguration),
					testAccCheckProviderConfigurationCustomValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "export_shell_variables", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.provider_name", "kubernetes"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.#", "3"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.4105667123.name", "host"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.4105667123.sensitive", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.4105667123.description", ""),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.4105667123.value", "my-host"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.2169404039.name", "client_certificate"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.2169404039.sensitive", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.2169404039.description", ""),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.2169404039.value", "-----BEGIN CERTIFICATE-----\nMIIB9TCCAWACAQAwgbgxG"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3103878395.name", "config_path"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3103878395.sensitive", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3103878395.description", "A path to a kube config file. some typo..."),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3103878395.value", "~/.kube/config"),
				),
			},
			{
				Config: testAccScalrPorivderConfigurationCustomConfigUpdated(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.kubernetes", &providerConfiguration),
					testAccCheckProviderConfigurationCustomUpdatedValues(&providerConfiguration, rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.provider_name", "kubernetes"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.#", "3"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.4105667123.name", "host"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.4105667123.sensitive", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.4105667123.description", ""),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.4105667123.value", "my-host"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.476034915.name", "config_path"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.476034915.description", "A path to a kube config file."),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.476034915.sensitive", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.476034915.value", "~/.kube/config"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3067103566.name", "username"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3067103566.sensitive", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3067103566.description", ""),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3067103566.value", "my-username"),
				),
			},
			{
				Config:      testAccScalrPorivderConfigurationCustomWithAwsAttrConfig(rName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Provider type can't be changed."),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
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
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.credentials_type", "role_delegation"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.account_type", "gov-cloud"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_google(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrPorivderConfigurationGoogleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.google", &providerConfiguration),
					testAccCheckProviderConfigurationGoogleValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.project", "my-project"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.credentials", "my-credentials"),
				),
			},
			{
				Config: testAccScalrPorivderConfigurationGoogleUpdatedConfig(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.google", &providerConfiguration),
					testAccCheckProviderConfigurationGoogleUpdatedValues(&providerConfiguration, rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.project", "my-new-project"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.credentials", "my-new-credentials"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_azurerm(t *testing.T) {
	t.Skip("TODO: add a valid credentials for azurerm testing.")
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrPorivderConfigurationAzurermConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.azurerm", &providerConfiguration),
					testAccCheckProviderConfigurationAzurermValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.client_id", "my-client-id"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.client_secret", "my-client-secret"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.subscription_id", "my-subscription-id"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.tenant_id", "my-tenant-id"),
				),
			},
			{
				Config: testAccScalrPorivderConfigurationAzurermUpdatedConfig(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.azurerm", &providerConfiguration),
					testAccCheckProviderConfigurationAzurermUpdatedValues(&providerConfiguration, rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.client_id", "my-new-client-id"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.client_secret", "my-new-client-secret"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.subscription_id", "my-new-subscription-id"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.tenant_id", "my-new-tenant-id"),
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
		if providerConfiguration.ExportShellVariables != true {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", true, providerConfiguration.ExportShellVariables)
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

func testAccCheckProviderConfigurationGoogleValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
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
		if providerConfiguration.GoogleProject != "my-project" {
			return fmt.Errorf("bad google project, expected \"%s\", got: %#v", "my-project", providerConfiguration.GoogleProject)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationGoogleUpdatedValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ExportShellVariables != false {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.GoogleProject != "my-new-project" {
			return fmt.Errorf("bad google project, expected \"%s\", got: %#v", "my-new-project", providerConfiguration.GoogleProject)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationAzurermValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
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
		if providerConfiguration.AzurermClientId != "my-client-id" {
			return fmt.Errorf("bad azurerm client id, expected \"%s\", got: %#v", "my-client-id", providerConfiguration.AzurermClientId)
		}
		if providerConfiguration.AzurermSubscriptionId != "my-subscription-id" {
			return fmt.Errorf("bad azurerm subscription id, expected \"%s\", got: %#v", "my-subscription-id", providerConfiguration.AzurermSubscriptionId)
		}
		if providerConfiguration.AzurermTenantId != "my-tenant-id" {
			return fmt.Errorf("bad azurerm tenant id, expected \"%s\", got: %#v", "my-tenant-id", providerConfiguration.AzurermTenantId)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationAzurermUpdatedValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ExportShellVariables != false {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.AzurermClientId != "my-new-client-id" {
			return fmt.Errorf("bad azurerm client id, expected \"%s\", got: %#v", "my-new-client-id", providerConfiguration.AzurermClientId)
		}
		if providerConfiguration.AzurermSubscriptionId != "my-new-subscription-id" {
			return fmt.Errorf("bad azurerm subscription id, expected \"%s\", got: %#v", "my-new-subscription-id", providerConfiguration.AzurermSubscriptionId)
		}
		if providerConfiguration.AzurermTenantId != "my-new-tenant-id" {
			return fmt.Errorf("bad azurerm tenant id, expected \"%s\", got: %#v", "my-new-tenant-id", providerConfiguration.AzurermTenantId)
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

		scalrClient := testAccProvider.Meta().(*scalr.Client)

		providerConfigurationResource, err := scalrClient.ProviderConfigurations.Read(ctx, rs.Primary.ID)

		if err != nil {
			return err
		}

		*providerConfiguration = *providerConfigurationResource

		return nil
	}
}

func testAccCheckProviderConfigurationResourceDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

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

func testAccScalrPorivderConfigurationCustomConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name                   = "%s"
  account_id             = "%s"
  export_shell_variables = true
  custom {
    provider_name = "kubernetes"
    argument {
      name      = "config_path"
      value     = "~/.kube/config"
      sensitive = false
	  description = "A path to a kube config file. some typo..."
    }
    argument {
      name      = "client_certificate"
      value     = "-----BEGIN CERTIFICATE-----\nMIIB9TCCAWACAQAwgbgxG"
      sensitive = true
    }
    argument {
      name      = "host"
      value     = "my-host"
    }
  }
}
`, name, defaultAccount)
}
func testAccScalrPorivderConfigurationCustomConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name                   = "%s"
  account_id             = "%s"
  custom {
    provider_name = "kubernetes"
    argument {
      name      = "config_path"
      value     = "~/.kube/config"
      sensitive = true
	  description = "A path to a kube config file."
    }
    argument {
      name      = "host"
      value     = "my-host"
    }
	argument {
		name      = "username"
		value     = "my-username"
	  }
  }
}
`, name, defaultAccount)
}

func testAccScalrPorivderConfigurationCustomWithAwsAttrConfig(name string) string {
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

func testAccScalrPorivderConfigurationGoogleConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "google" {
  name       = "%s"
  account_id = "%s"
  google {
    project     = "my-project"
    credentials = "my-credentials"
  }
}
`, name, defaultAccount)
}

func testAccScalrPorivderConfigurationGoogleUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "google" {
  name       = "%s"
  account_id = "%s"
  google {
    project     = "my-new-project"
    credentials = "my-new-credentials"
  }
}
`, name, defaultAccount)
}

func testAccScalrPorivderConfigurationAzurermConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "azurerm" {
  name       = "%s"
  account_id = "%s"
  azurerm {
    client_id       = "my-client-id"
    client_secret   = "my-client-secret"
    subscription_id = "my-subscription-id"
    tenant_id       = "my-tenant-id"
  }
}
`, name, defaultAccount)
}

func testAccScalrPorivderConfigurationAzurermUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "azurerm" {
  name       = "%s"
  account_id = "%s"
  azurerm {
    client_id       = "my-new-client-id"
    client_secret   = "my-new-client-secret"
    subscription_id = "my-new-subscription-id"
    tenant_id       = "my-new-tenant-id"
  }
}
`, name, defaultAccount)
}
