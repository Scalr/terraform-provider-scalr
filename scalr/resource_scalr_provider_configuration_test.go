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
				Config: testAccScalrProviderConfigurationCustomConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.kubernetes", &providerConfiguration),
					testAccCheckProviderConfigurationCustomValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "export_shell_variables", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "scalr.#", "0"),
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
				Config: testAccScalrProviderConfigurationCustomConfigUpdated(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.kubernetes", &providerConfiguration),
					testAccCheckProviderConfigurationCustomUpdatedValues(&providerConfiguration, rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "scalr.#", "0"),
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
				Config:      testAccScalrProviderConfigurationCustomWithAwsAttrConfig(rName),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationAwsConfig(rName),
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
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.access_key", "my-access-key"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.secret_key", "my-secret-key"),
				),
			},
			{
				Config: testAccScalrProviderConfigurationAwsUpdatedConfig(rNewName),
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
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.access_key", ""),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.secret_key", "my-new-secret-key"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_scalr(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	scalrHostname := os.Getenv("SCALR_HOSTNAME")
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationGoogleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.google", &providerConfiguration),
					testAccCheckProviderConfigurationGoogleValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "scalr.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.project", "my-project"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.credentials", "my-credentials"),
				),
			},
			{
				Config: testAccScalrProviderConfigurationGoogleUpdatedConfig(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.google", &providerConfiguration),
					testAccCheckProviderConfigurationGoogleUpdatedValues(&providerConfiguration, rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "scalr.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.project", "my-new-project"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.credentials", "my-new-credentials"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_azurerm(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationAzurermConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.azurerm", &providerConfiguration),
					testAccCheckProviderConfigurationAzurermValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "scalr.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.client_id", "my-client-id"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.client_secret", "my-client-secret"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.subscription_id", "my-subscription-id"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.tenant_id", "my-tenant-id"),
				),
			},
			{
				Config: testAccScalrProviderConfigurationAzurermUpdatedConfig(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.azurerm", &providerConfiguration),
					testAccCheckProviderConfigurationAzurermUpdatedValues(&providerConfiguration, rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "name", rNewName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "scalr.#", "0"),
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
		if providerConfiguration.AwsAccessKey != "my-access-key" {
			return fmt.Errorf("bad aws access key, expected \"%s\", got: %#v", "my-access-key", providerConfiguration.AwsAccessKey)
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
		if providerConfiguration.AwsAccessKey != "" {
			return fmt.Errorf("bad aws access key, expected \"%s\", got: %#v", "my-new-access-key", providerConfiguration.AwsAccessKey)
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

func testAccScalrProviderConfigurationCustomConfig(name string) string {
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
func testAccScalrProviderConfigurationCustomConfigUpdated(name string) string {
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

func testAccScalrProviderConfigurationCustomWithAwsAttrConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name                   = "%s"
  account_id             = "%s"
  export_shell_variables = false
  aws {
    secret_key = "my-secret-key"
    access_key = "my-access-key"
  }
}
`, name, defaultAccount)
}

func testAccScalrProviderConfigurationAwsConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "aws" {
  name                   = "%s"
  account_id             = "%s"
  export_shell_variables = false
  aws {
    secret_key = "my-secret-key"
    access_key = "my-access-key"
  }
}
`, name, defaultAccount)
}

func testAccScalrProviderConfigurationAwsUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "aws" {
  name                   = "%s"
  account_id             = "%s"
  export_shell_variables = true
  aws {
    secret_key = "my-new-secret-key"
  }
}
`, name, defaultAccount)
}

func testAccScalrProviderConfigurationGoogleConfig(name string) string {
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

func testAccScalrProviderConfigurationGoogleUpdatedConfig(name string) string {
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

func testAccScalrProviderConfigurationAzurermConfig(name string) string {
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

func testAccScalrProviderConfigurationAzurermUpdatedConfig(name string) string {
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
`, name, defaultAccount, os.Getenv("SCALR_HOSTNAME"), os.Getenv("SCALR_TOKEN"))
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
`, name, defaultAccount, os.Getenv("SCALR_HOSTNAME")+"/", os.Getenv("SCALR_TOKEN"))
}
