package cloudflare

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudflareZoneLockdown(t *testing.T) {
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_zone_lockdown." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testCloudflareZoneLockdownConfig(rnd, zoneID, "false", "1", "this is notes", rnd+"."+zoneName+"/*", "ip", "198.51.100.4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "paused", "false"),
					resource.TestCheckResourceAttr(name, "priority", "1"),
					resource.TestCheckResourceAttr(name, "description", "this is notes"),
					resource.TestCheckResourceAttr(name, "urls.#", "1"),
					resource.TestCheckResourceAttr(name, "configurations.#", "1"),
				),
			},
		},
	})
}

// test creating a config with only the required fields
func TestAccCloudflareZoneLockdown_OnlyRequired(t *testing.T) {
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_zone_lockdown." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testCloudflareZoneLockdownConfig(rnd, zoneID, "false", "1", "this is notes", rnd+"."+zoneName+"/*", "ip", "198.51.100.4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "urls.#", "1"),
					resource.TestCheckResourceAttr(name, "configurations.#", "1"),
				),
			},
		},
	})
}

func TestAccCloudflareZoneLockdown_Import(t *testing.T) {
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_zone_lockdown." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareWAFRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCloudflareZoneLockdownConfig(rnd, zoneID, "false", "1", "this is notes", rnd+"."+zoneName+"/*", "ip", "198.51.100.4"),
			},
			{
				ResourceName:        name,
				ImportStateIdPrefix: fmt.Sprintf("%s/", zoneID),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

func testCloudflareZoneLockdownConfig(resourceID, zoneID, paused, priority, description, url, target, value string) string {
	return fmt.Sprintf(`
				resource "cloudflare_zone_lockdown" "%[1]s" {
					zone_id = "%[2]s"
					paused = "%[3]s"
					priority = "%[4]s"
					description = "%[5]s"
					urls = ["%[6]s"]
					configurations {
						target = "%[7]s"
						value = "%[8]s"
					}
				}`, resourceID, zoneID, paused, priority, description, url, target, value)
}
