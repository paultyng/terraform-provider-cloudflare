package cloudflare

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareWAFPackages_NoFilter(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_packages.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFPackagesConfig(zoneID, map[string]string{}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFPackagesDataSourceID(name),
					resource.TestCheckResourceAttr(name, "packages.#", "3"),
				),
			},
		},
	})
}

func TestAccCloudflareWAFPackages_MatchName(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_packages.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFPackagesConfig(zoneID, map[string]string{"name": "USER"}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFPackagesDataSourceID(name),
					resource.TestCheckResourceAttr(name, "packages.#", "1"),
				),
			},
		},
	})
}

func TestAccCloudflareWAFPackages_MatchDetectionMode(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_packages.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFPackagesConfig(zoneID, map[string]string{"detection_mode": "traditional"}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFPackagesDataSourceID(name),
					resource.TestCheckResourceAttr(name, "packages.#", "2"),
				),
			},
		},
	})
}

func TestAccCloudflareWAFPackages_MatchSensitivity(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_packages.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFPackagesConfig(zoneID, map[string]string{"sensitivity": "high"}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFPackagesDataSourceID(name),
				),
			},
		},
	})
}

func TestAccCloudflareWAFPackages_MatchActionMode(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_packages.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFPackagesConfig(zoneID, map[string]string{"action_mode": "challenge"}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFPackagesDataSourceID(name),
				),
			},
		},
	})
}

func testAccCheckCloudflareWAFPackagesDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		all := s.RootModule().Resources
		rs, ok := all[n]
		if !ok {
			return fmt.Errorf("Can't find WAF Packages data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot WAF Packages source ID not set")
		}
		return nil
	}
}

func testAccCloudflareWAFPackagesConfig(zoneID string, filters map[string]string, name string) string {
	filters_str := make([]string, 0, len(filters))
	for k, v := range filters {
		filters_str = append(filters_str, fmt.Sprintf(`%[1]s = "%[2]s"`, k, v))
	}

	return fmt.Sprintf(`
				data "cloudflare_waf_packages" "%[1]s" {
					zone_id = "%[2]s"

					filter {
						%[3]s
					}
				}`, name, zoneID, strings.Join(filters_str, "\n\t\t\t\t"))
}
