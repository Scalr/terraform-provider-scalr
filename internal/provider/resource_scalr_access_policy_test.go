package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrAccessPolicy_basic(t *testing.T) {
	ap := &scalr.AccessPolicy{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAccessPolicyBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAccessPolicyExists("scalr_access_policy.test", ap),
					resource.TestCheckResourceAttrSet("scalr_access_policy.test", "id"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "subject.0.type", "user"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "subject.0.id", testUser),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "scope.0.type", "environment"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "role_ids.0", readOnlyRole),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "role_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccScalrAccessPolicy_bad_scope(t *testing.T) {
	rg, _ := regexp.Compile(`scope.0.type must be one of \[workspace, environment, account], got: universe`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrAccessPolicyBadScope(),
				ExpectError: rg,
			},
		},
	})
}

func TestAccScalrAccessPolicy_bad_subject(t *testing.T) {
	rg, _ := regexp.Compile(`subject.0.type must be one of \[user, team, service_account], got: grandpa`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrAccessPolicyBadSubject(),
				ExpectError: rg,
			},
		},
	})
}

func TestAccScalrAccessPolicy_update(t *testing.T) {
	ap := &scalr.AccessPolicy{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAccessPolicyBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_access_policy.test", "id"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "subject.0.type", "user"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "subject.0.id", testUser),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "scope.0.type", "environment"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "role_ids.0", readOnlyRole),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "role_ids.#", "1"),
				),
			},

			{
				Config: testAccScalrAccessPolicyUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAccessPolicyExists("scalr_access_policy.test", ap),
					resource.TestCheckResourceAttrSet("scalr_access_policy.test", "id"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "subject.0.type", "user"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "subject.0.id", testUser),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "scope.0.type", "environment"),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "role_ids.0", readOnlyRole),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "role_ids.1", userRole),
					resource.TestCheckResourceAttr("scalr_access_policy.test", "role_ids.#", "2"),
				),
			},

			{
				Config:      testAccScalrAccessPolicyEmptyRoleId(rInt),
				ExpectError: regexp.MustCompile("Got error during parsing role ids: 0-th value is empty"),
			},
		},
	})
}

func TestAccScalrAccessPolicy_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAccessPolicyBasic(rInt),
			},

			{
				ResourceName:      "scalr_access_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrAccessPolicyExists(resId string, ap *scalr.AccessPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the ap
		r, err := scalrClient.AccessPolicies.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*ap = *r

		return nil
	}
}

func testAccCheckScalrAccessPolicyDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_access_policy" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.AccessPolicies.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("AccessPolicy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

var iamPolicyTemplate = `
resource "scalr_environment" "test" {
  name = "test-access-policies-provider-%d"
  account_id = "%s"
}


resource "scalr_access_policy" "test" {
  subject {
    type = "user"
    id = "%s"
  }
  scope {
    type = "environment"
    id = scalr_environment.test.id
  }
  role_ids = [
    "%s"
  ]
}`

func testAccScalrAccessPolicyBadScope() string {
	return fmt.Sprintf(`
resource "scalr_access_policy" "test" {
  subject {
    type = "user"
    id = "%s"
  }
  scope {
    type = "universe"
    id = "%s"
  }
  role_ids = [
    "%s"
  ]
}

`, testUser, defaultAccount, readOnlyRole)
}

func testAccScalrAccessPolicyBadSubject() string {
	return fmt.Sprintf(`
resource "scalr_access_policy" "test" {
  subject {
    type = "grandpa"
    id = "%s"
  }
  scope {
    type = "account"
    id = "%s"
  }
  role_ids = [
    "%s"
  ]
}

`, testUser, defaultAccount, readOnlyRole)
}

func testAccScalrAccessPolicyBasic(rInt int) string {
	return fmt.Sprintf(iamPolicyTemplate, rInt, defaultAccount, testUser, readOnlyRole)
}

func testAccScalrAccessPolicyEmptyRoleId(rInt int) string {
	return fmt.Sprintf(iamPolicyTemplate, rInt, defaultAccount, testUser, "")
}

func testAccScalrAccessPolicyUpdate(rInt int) string {
	return fmt.Sprintf(iamPolicyTemplate, rInt, defaultAccount, testUser, fmt.Sprintf("%s\", \"%s", readOnlyRole, userRole))
}
