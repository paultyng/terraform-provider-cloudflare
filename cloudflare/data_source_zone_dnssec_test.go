package cloudflare

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareZoneDNSSEC(t *testing.T) {
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_zone_dnssec.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareZoneDNSSECConfig(zoneID, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareZoneDNSSECDataSourceID(name),
					resource.TestCheckResourceAttrSet(name, "zone_id"),
					resource.TestMatchResourceAttr(name, "status", regexp.MustCompile("active|disabled|pending")),
					resource.TestCheckResourceAttrSet(name, "flags"),
					resource.TestCheckResourceAttrSet(name, "algorithm"),
					resource.TestCheckResourceAttrSet(name, "key_type"),
					resource.TestCheckResourceAttrSet(name, "digest_type"),
					resource.TestCheckResourceAttrSet(name, "digest_algorithm"),
					resource.TestCheckResourceAttrSet(name, "digest"),
					resource.TestCheckResourceAttrSet(name, "ds"),
					resource.TestCheckResourceAttrSet(name, "key_tag"),
					resource.TestCheckResourceAttrSet(name, "public_key"),
				),
			},
		},
	})
}

func testAccCheckCloudflareZoneDNSSECDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		all := s.RootModule().Resources
		rs, ok := all[n]
		if !ok {
			return fmt.Errorf("Can't find Zone DNSSEC data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot Zone DNSSEC ID not set")
		}
		return nil
	}
}

func testAccCloudflareZoneDNSSECConfig(zoneID string, name string) string {
	return fmt.Sprintf(`
data "cloudflare_zone_dnssec" "%s" {
	zone_id = cloudflare_zone_dnssec.%s.zone_id
}

%s
`, name, name, testAccCloudflareZoneDNSSECResourceConfig(zoneID, name))
}

func testAccCloudflareZoneDNSSECResourceConfig(zoneID string, name string) string {
	return fmt.Sprintf(`
resource "cloudflare_zone_dnssec" "%s" {
	zone_id = "%s"
}`, name, zoneID)
}
