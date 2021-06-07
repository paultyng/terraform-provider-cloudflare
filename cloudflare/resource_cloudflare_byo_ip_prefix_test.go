package cloudflare

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudflareBYOIPPrefix(t *testing.T) {
	t.Parallel()
	prefixID := os.Getenv("CLOUDFLARE_BYO_IP_PREFIX_ID")

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_byo_ip_prefix.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAccount(t)
			testAccPreCheckBYOIPPrefix(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareBYOIPPrefixConfig(prefixID, "BYOIP Prefix Description old", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						name, "description", "BYOIP Prefix Description old"),
				),
			},
			{
				Config: testAccCheckCloudflareBYOIPPrefixConfig(prefixID, "BYOIP Prefix Description new", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						name, "description", "BYOIP Prefix Description new"),
				),
			},
		},
	})
}

func testAccCheckCloudflareBYOIPPrefixConfig(prefixID, description, name string) string {
	return fmt.Sprintf(`
  resource "cloudflare_byo_ip_prefix" "%[3]s" {
	  prefix_id = "%[1]s"
	  description = "%[2]s"
  }`, prefixID, description, name)
}
