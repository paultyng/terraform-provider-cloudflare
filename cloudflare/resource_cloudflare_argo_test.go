package cloudflare

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudflareArgoOnlySetTieredCaching(t *testing.T) {
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_argo.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareArgoTieredCachingConfig(zoneID, rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "tiered_caching", "on"),
					resource.TestCheckNoResourceAttr(name, "smart_routing"),
				),
			},
		},
	})
}

func TestAccCloudflareArgoSetTieredAndSmartRouting(t *testing.T) {
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_argo.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareArgoFullConfig(zoneID, rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "tiered_caching", "on"),
					resource.TestCheckResourceAttr(name, "smart_routing", "on"),
				),
			},
		},
	})
}

func testAccCheckCloudflareArgoTieredCachingConfig(zoneID, name string) string {
	return fmt.Sprintf(`
  resource "cloudflare_argo" "%[2]s" {
	  zone_id        = "%[1]s"
    tiered_caching = "on"
  }`, zoneID, name)
}

func testAccCheckCloudflareArgoFullConfig(zoneID, name string) string {
	return fmt.Sprintf(`
  resource "cloudflare_argo" "%[2]s" {
	  zone_id        = "%[1]s"
    tiered_caching = "on"
    smart_routing  = "on"
  }`, zoneID, name)
}
