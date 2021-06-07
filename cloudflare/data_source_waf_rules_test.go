package cloudflare

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareWAFRules_NoFilter(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_rules.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFRulesConfig(zoneID, map[string]string{}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFRulesDataSourceID(name),
					resource.TestCheckResourceAttrSet(name, "rules.#"),
				),
			},
		},
	})
}

func TestAccCloudflareWAFRules_MatchDescription(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_rules.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFRulesConfig(zoneID, map[string]string{"description": "^SLR: .*"}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFRulesDataSourceID(name),
					resource.TestCheckResourceAttrSet(name, "rules.#"),
				),
			},
		},
	})
}

func TestAccCloudflareWAFRules_MatchMode(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_waf_rules.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareWAFRulesConfig(zoneID, map[string]string{"mode": "on"}, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWAFRulesDataSourceID(name),
				),
			},
		},
	})
}

func testAccCheckCloudflareWAFRulesDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		all := s.RootModule().Resources
		rs, ok := all[n]
		if !ok {
			return fmt.Errorf("Can't find WAF Rules data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot WAF Rules source ID not set")
		}
		return nil
	}
}

func testAccCloudflareWAFRulesConfig(zoneID string, filters map[string]string, name string) string {
	filters_str := make([]string, 0, len(filters))
	for k, v := range filters {
		filters_str = append(filters_str, fmt.Sprintf(`%[1]s = "%[2]s"`, k, v))
	}

	return fmt.Sprintf(`
data "cloudflare_waf_rules" "%[1]s" {
	zone_id = "%[2]s"

	filter {
		%[3]s
	}
}`, name, zoneID, strings.Join(filters_str, "\n\t\t"))
}
