package cloudflare

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareWAFOverrideCreateAndUpdate(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the WAF
	// overrides endpoint does not yet support the API tokens and it
	// results in misleading state error messages.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_waf_override.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareWAFOverrideDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWAFOverrideBasicConfig(zoneID, zoneName, rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "urls.#", "2"),
					resource.TestCheckResourceAttr(name, "urls.0", fmt.Sprintf("%s/basic-waf-override", zoneName)),
					resource.TestCheckResourceAttr(name, "urls.1", fmt.Sprintf("%s/another-basic-waf-override", zoneName)),
					resource.TestCheckResourceAttr(name, "rules.100015", "disable"),
					resource.TestCheckResourceAttr(name, "groups.ea8687e59929c1fd05ba97574ad43f77", "default"),
					resource.TestCheckResourceAttr(name, "rewrite_action.default", "block"),
					resource.TestCheckResourceAttr(name, "rewrite_action.challenge", "block"),
				),
			},
			{
				Config: testAccCheckCloudflareWAFOverrideBasicConfigUpdated(zoneID, zoneName, rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "urls.#", "1"),
					resource.TestCheckResourceAttr(name, "urls.0", fmt.Sprintf("%s/basic-waf-override", zoneName)),
					resource.TestCheckResourceAttr(name, "rules.100015", "disable"),
					resource.TestCheckResourceAttr(name, "groups.ea8687e59929c1fd05ba97574ad43f77", "default"),
					resource.TestCheckResourceAttr(name, "rewrite_action.default", "block"),
					resource.TestCheckResourceAttr(name, "rewrite_action.challenge", "block"),
				),
			},
		},
	})
}

func TestAccCloudflareWAFOverrideGroupOnly(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the WAF
	// overrides endpoint does not yet support the API tokens and it
	// results in misleading state error messages.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_waf_override.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareWAFOverrideDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWAFOverrideGroupsOnlyConfig(zoneID, zoneName, rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "urls.#", "1"),
					resource.TestCheckResourceAttr(name, "urls.0", fmt.Sprintf("%s/group-only-override", zoneName)),
					resource.TestCheckResourceAttr(name, "groups.ea8687e59929c1fd05ba97574ad43f77", "default"),
					resource.TestCheckResourceAttr(name, "rewrite_action.default", "block"),
					resource.TestCheckResourceAttr(name, "rewrite_action.challenge", "block"),
					resource.TestCheckNoResourceAttr(name, "rules"),
				),
			},
		},
	})
}

func testAccCheckCloudflareWAFOverrideBasicConfig(zoneID, zoneName, name string) string {
	return fmt.Sprintf(`
		resource "cloudflare_waf_override" "%[3]s" {
			zone_id = "%[1]s"
			urls = ["%[2]s/basic-waf-override", "%[2]s/another-basic-waf-override"]
			rules = {
				"100015": "disable"
			}
			groups = {
				"ea8687e59929c1fd05ba97574ad43f77": "default"
			}
			rewrite_action = {
				"default": "block",
  			"challenge": "block",
			}
		}`, zoneID, zoneName, name)
}

func testAccCheckCloudflareWAFOverrideGroupsOnlyConfig(zoneID, zoneName, name string) string {
	return fmt.Sprintf(`
		resource "cloudflare_waf_override" "%[3]s" {
			zone_id = "%[1]s"
			urls = ["%[2]s/group-only-override"]
			groups = {
				"ea8687e59929c1fd05ba97574ad43f77": "default"
			}
			rewrite_action = {
				"default": "block",
  			"challenge": "block",
			}
		}`, zoneID, zoneName, name)
}

func testAccCheckCloudflareWAFOverrideBasicConfigUpdated(zoneID, zoneName, name string) string {
	return fmt.Sprintf(`
		resource "cloudflare_waf_override" "%[3]s" {
			zone_id = "%[1]s"
			urls = ["%[2]s/basic-waf-override"]
			rules = {
				"100015": "disable"
			}
			groups = {
				"ea8687e59929c1fd05ba97574ad43f77": "default"
			}
			rewrite_action = {
				"default": "block",
  			"challenge": "block",
			}
		}`, zoneID, zoneName, name)
}

func testAccCheckCloudflareWAFOverrideDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_waf_override" {
			continue
		}

		_, err := client.WAFOverride(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("WAFOverride still exists")
		}
	}

	return nil
}
