package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	tfe "github.com/scalr/go-tfe"
)

func TestAccTFEPolicySet_basic(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_update(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateEmpty(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "1"),
				),
			},
			{
				Config: testAccTFEPolicySet_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updatePopulated(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_updatePopulated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "terraform-populated-updated"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateToGlobal(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_global,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetGlobal(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "terraform-global"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "true"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_updateToWorkspace(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_global,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetGlobal(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "terraform-global"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "true"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "1"),
				),
			},

			{
				Config: testAccTFEPolicySet_populated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetPopulated(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "terraform-populated"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "workspace_external_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFEPolicySet_vcs(t *testing.T) {
	policySet := &tfe.PolicySet{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if GITHUB_TOKEN == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
			if GITHUB_POLICY_SET_IDENTIFIER == "" {
				t.Skip("Please set GITHUB_POLICY_SET_IDENTIFIER to run this test")
			}
			if GITHUB_POLICY_SET_BRANCH == "" {
				t.Skip("Please set GITHUB_POLICY_SET_BRANCH to run this test")
			}
			if GITHUB_POLICY_SET_PATH == "" {
				t.Skip("Please set GITHUB_POLICY_SET_PATH to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_vcs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetExists("scalr_policy_set.foobar", policySet),
					testAccCheckTFEPolicySetAttributes(policySet),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "description", "Policy Set"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "global", "false"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "vcs_repo.0.identifier", GITHUB_POLICY_SET_IDENTIFIER),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "vcs_repo.0.branch", GITHUB_POLICY_SET_BRANCH),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "vcs_repo.0.ingress_submodules", "true"),
					resource.TestCheckResourceAttr(
						"scalr_policy_set.foobar", "policies_path", GITHUB_POLICY_SET_PATH),
				),
			},
		},
	})
}

func TestAccTFEPolicySetImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySet_populated,
			},

			{
				ResourceName:      "scalr_policy_set.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEPolicySetExists(n string, policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ps, err := tfeClient.PolicySets.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if ps.ID != rs.Primary.ID {
			return fmt.Errorf("PolicySet not found")
		}

		*policySet = *ps

		return nil
	}
}

func testAccCheckTFEPolicySetAttributes(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policySet.Name != "tst-terraform" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Description != "Policy Set" {
			return fmt.Errorf("Bad description: %s", policySet.Description)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetPopulated(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		if policySet.Name != "terraform-populated" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := tfeClient.Policies.Read(ctx, policyID)
		if policy.Name != "policy-foo" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := tfeClient.Workspaces.Read(ctx, "tst-terraform", "workspace-foo")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetGlobal(policySet *tfe.PolicySet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		if policySet.Name != "terraform-global" {
			return fmt.Errorf("Bad name: %s", policySet.Name)
		}

		if !policySet.Global {
			return fmt.Errorf("Bad value for global: %v", policySet.Global)
		}

		if len(policySet.Policies) != 1 {
			return fmt.Errorf("Wrong number of policies: %v", len(policySet.Policies))
		}

		policyID := policySet.Policies[0].ID
		policy, _ := tfeClient.Policies.Read(ctx, policyID)
		if policy.Name != "policy-foo" {
			return fmt.Errorf("Wrong member policy: %v", policy.Name)
		}

		// Even though the terraform config should have 0 workspaces, the API will return
		// workspaces for global policy sets. This list would be the same as listing the
		// workspaces for the organization itself.
		if len(policySet.Workspaces) != 1 {
			return fmt.Errorf("Wrong number of workspaces: %v", len(policySet.Workspaces))
		}

		workspaceID := policySet.Workspaces[0].ID
		workspace, _ := tfeClient.Workspaces.Read(ctx, "tst-terraform", "workspace-foo")
		if workspace.ID != workspaceID {
			return fmt.Errorf("Wrong member workspace: %v", workspace.Name)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_policy_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.PolicySets.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Sentinel policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFEPolicySet_basic = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "${scalr_organization.foobar.id}"
  policy_ids   = ["${scalr_sentinel_policy.foo.id}"]
}`

const testAccTFEPolicySet_empty = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}
 resource "scalr_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "${scalr_organization.foobar.id}"
}`

const testAccTFEPolicySet_populated = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_workspace" "foo" {
  name         = "workspace-foo"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_policy_set" "foobar" {
  name                   = "terraform-populated"
  organization           = "${scalr_organization.foobar.id}"
  policy_ids             = ["${scalr_sentinel_policy.foo.id}"]
  workspace_external_ids = ["${scalr_workspace.foo.external_id}"]
}`

const testAccTFEPolicySet_updatePopulated = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_sentinel_policy" "bar" {
  name         = "policy-bar"
  policy       = "main = rule { false }"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_workspace" "foo" {
  name         = "workspace-foo"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_workspace" "bar" {
  name         = "workspace-bar"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_policy_set" "foobar" {
  name                   = "terraform-populated-updated"
  organization           = "${scalr_organization.foobar.id}"
  policy_ids             = ["${scalr_sentinel_policy.bar.id}"]
  workspace_external_ids = ["${scalr_workspace.bar.external_id}"]
}`

const testAccTFEPolicySet_global = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_sentinel_policy" "foo" {
  name         = "policy-foo"
  policy       = "main = rule { true }"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_workspace" "foo" {
  name         = "workspace-foo"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_policy_set" "foobar" {
  name         = "terraform-global"
  organization = "${scalr_organization.foobar.id}"
  global       = true
  policy_ids   = ["${scalr_sentinel_policy.foo.id}"]
}`

var testAccTFEPolicySet_vcs = fmt.Sprintf(`
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_oauth_client" "test" {
  organization     = "${scalr_organization.foobar.id}"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "scalr_policy_set" "foobar" {
  name         = "tst-terraform"
  description  = "Policy Set"
  organization = "${scalr_organization.foobar.id}"
  vcs_repo {
    identifier         = "%s"
    branch             = "%s"
    ingress_submodules = true
    oauth_token_id     = "${scalr_oauth_client.test.oauth_token_id}"
  }

  policies_path = "%s"
}
`,
	GITHUB_TOKEN,
	GITHUB_POLICY_SET_IDENTIFIER,
	GITHUB_POLICY_SET_BRANCH,
	GITHUB_POLICY_SET_PATH,
)
