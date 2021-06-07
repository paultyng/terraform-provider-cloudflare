package cloudflare

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareWAFGroups_NoFilter(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_groups.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFGroupsConfig(zoneID, map[string]string{}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFGroupsDataSourceID(name),
					resource.TestCheckResourceAttr(name, "groups.#", "30"),
				),
			},
		},
	})
}

func TestAccCloudflareWAFGroups_MatchName(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_groups.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFGroupsConfig(zoneID, map[string]string{"name": "OWASP.*"}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFGroupsDataSourceID(name),
					resource.TestCheckResourceAttr(name, "groups.#", "20"),
				),
			},
		},
	})
}

func TestAccCloudflareWAFGroups_MatchMode(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_groups.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFGroupsConfig(zoneID, map[string]string{"mode": "on"}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFGroupsDataSourceID(name),
				),
			},
		},
	})
}

func testAccCheckCloudflareWAFGroupsDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		all := s.RootModule().Resources
		rs, ok := all[n]
		if !ok {
			return fmt.Errorf("Can't find WAF Groups data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot WAF Groups source ID not set")
		}
		return nil
	}
}

func testAccCloudflareWAFGroupsConfig(zoneID string, filters map[string]string, name string) string {
	filters_str := make([]string, 0, len(filters))
	for k, v := range filters {
		filters_str = append(filters_str, fmt.Sprintf(`%[1]s = "%[2]s"`, k, v))
	}

	return fmt.Sprintf(`
				data "cloudflare_waf_groups" "%[1]s" {
					zone_id = "%[2]s"

					filter {
						%[3]s
					}
				}`, name, zoneID, strings.Join(filters_str, "\n\t\t\t\t"))
}
