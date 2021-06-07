package cloudflare

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareTeamsListBasic(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_teams_list.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareTeamsListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareTeamsListConfigBasic(rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "type", "SERIAL"),
					resource.TestCheckResourceAttr(name, "description", "My description"),
					resource.TestCheckResourceAttr(name, "items.#", "2"),
					resource.TestCheckResourceAttr(name, "items.0", "asdf-1234"),
				),
			},
		},
	})
}

func testAccCloudflareTeamsListConfigBasic(rnd, accountID string) string {
	return fmt.Sprintf(`
resource "cloudflare_teams_list" "%[1]s" {
	account_id  = "%[2]s"
	name        = "%[1]s"
	description = "My description"
	type        = "SERIAL"
	items       = ["asdf-1234", "asdf-5678"]
}
`, rnd, accountID)
}

func testAccCheckCloudflareTeamsListDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_teams_list" {
			continue
		}

		_, err := client.TeamsList(context.Background(), rs.Primary.Attributes["account_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Teams List still exists")
		}
	}

	return nil
}
