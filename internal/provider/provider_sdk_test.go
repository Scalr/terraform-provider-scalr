package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/scalr/terraform-provider-scalr/internal/framework/defaults"
)

var testAccProviderSDK *schema.Provider
var errNoInstanceId = fmt.Errorf("No instance ID is set")
var githubToken = os.Getenv("githubToken")

// ctx is used as default context.Context when making API calls.
var ctx = context.Background()

func init() {
	testAccProviderSDK = Provider(testProviderVersion)
}

func TestProvider(t *testing.T) {
	if err := Provider(testProviderVersion).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = Provider(testProviderVersion)
}

func testAccPreCheck(t *testing.T) {
	// The credentials must be provided by the CLI config file for testing.
	if diags := Provider(testProviderVersion).Configure(context.Background(), &terraform.ResourceConfig{}); diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("err: %s", d.Summary)
			}
		}
	}
	// Set env variable to allow `account_id` compute the default value
	_ = os.Setenv(defaults.CurrentAccountIDEnvVar, defaultAccount)
}

func testVcsAccGithubTokenPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if githubToken == "" {
		t.Skip("Please set githubToken to run this test")
	}
}
