package scalr

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/scalr/terraform-provider-scalr/version"
)

var testAccProvider *schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)
var noInstanceIdErr = fmt.Errorf("No instance ID is set")
var githubToken = os.Getenv("githubToken")

// ctx is used as default context.Context when making API calls.
var ctx = context.Background()

func init() {
	testAccProvider = Provider()
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"scalr": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = Provider()
}

func TestProvider_versionConstraints(t *testing.T) {
	cases := map[string]struct {
		constraints *disco.Constraints
		version     string
		result      string
	}{
		"compatible version": {
			constraints: &disco.Constraints{
				Service: "tfe.v2.1",
				Product: "scalr-provider",
				Minimum: "0.4.0",
				Maximum: "0.7.0",
			},
			version: "0.6.0",
		},
		"version too old": {
			constraints: &disco.Constraints{
				Service: "tfe.v2.1",
				Product: "scalr-provider",
				Minimum: "0.4.0",
				Maximum: "0.7.0",
			},
			version: "0.3.0",
			result:  "upgrade the Scalr provider to >= 0.4.0",
		},
		"version too new": {
			constraints: &disco.Constraints{
				Service: "tfe.v2.1",
				Product: "scalr-provider",
				Minimum: "0.4.0",
				Maximum: "0.7.0",
			},
			version: "0.8.0",
			result:  "downgrade the Scalr provider to <= 0.7.0",
		},
	}

	// Save and restore the actual version.
	v := version.ProviderVersion
	defer func() {
		version.ProviderVersion = v
	}()

	for name, tc := range cases {
		// Set the version for this test.
		version.ProviderVersion = tc.version

		err := checkConstraints(tc.constraints)
		if err == nil && tc.result != "" {
			t.Fatalf("%s: expected error to contain %q, but got no error", name, tc.result)
		}
		if err != nil && tc.result == "" {
			t.Fatalf("%s: unexpected error: %v", name, err)
		}
		if err != nil && !strings.Contains(err.Error(), tc.result) {
			t.Fatalf("%s: expected error to contain %q, got: %v", name, tc.result, err)
		}
	}
}

func testAccPreCheck(t *testing.T) {
	// The credentials must be provided by the CLI config file for testing.
	if diags := Provider().Configure(context.Background(), &terraform.ResourceConfig{}); diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("err: %s", d.Summary)
			}
		}
	}
}

func testVcsAccGithubTokenPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if githubToken == "" {
		t.Skip("Please set githubToken to run this test")
	}
}
