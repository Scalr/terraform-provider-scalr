package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccScalrIamTeamsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamTeamsDataSourceInitConfig, // depends_on works improperly with data sources
			},
			{
				Config: testAccScalrIamTeamsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIamTeamsDataSourceNameFilter(),
				),
			},
			{
				Config: testAccScalrIamTeamsDataSourceInitConfig, // depends_on works improperly with data sources
			},
		},
	})
}

func testAccCheckIamTeamsDataSourceNameFilter() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var expectedIds []string
		resourceNames := []string{"kubernetes2", "consul"}
		for _, name := range resourceNames {
			rsName := "scalr_iam_team." + name
			rs, ok := s.RootModule().Resources[rsName]
			if !ok {
				return fmt.Errorf("Not found: %s", rsName)
			}
			expectedIds = append(expectedIds, rs.Primary.ID)

		}
		dataSource, ok := s.RootModule().Resources["data.scalr_iam_teams.kubernetes2consul"]
		if !ok {
			return fmt.Errorf("Not found: data.scalr_iam_teams.kubernetes2consul")
		}
		if dataSource.Primary.Attributes["ids.#"] != "2" {
			return fmt.Errorf("Bad team ids, expected: %#v, got: %#v", expectedIds, dataSource.Primary.Attributes["ids"])
		}

		resultIds := []string{dataSource.Primary.Attributes["ids.0"], dataSource.Primary.Attributes["ids.1"]}

		for _, expectedId := range expectedIds {
			found := false
			for _, resultId := range resultIds {
				if resultId == expectedId {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("Bad team ids, expected: %#v, got: %#v", expectedIds, resultIds)
			}
		}
		return nil
	}
}

var testAccScalrIamTeamsDataSourceInitConfig = fmt.Sprintf(`
resource "scalr_iam_team" "kubernetes1" {
  name        = "kubernetes1"
  account_id  = "%[1]s"
}
resource "scalr_iam_team" "kubernetes2" {
  name        = "kubernetes2"
  account_id  = "%[1]s"
}
resource "scalr_iam_team" "consul" {
  name        = "consul"
  account_id  = "%[1]s"
  
}`, defaultAccount)

var testAccScalrIamTeamsDataSourceConfig = testAccScalrIamTeamsDataSourceInitConfig + `
data "scalr_iam_teams" "kubernetes2consul" {
  name = "in:kubernetes2,consul"
}
`
