package cloudflare

import (
	"context"
	"fmt"
	"os"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareWAFRule_CreateThenUpdate(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	ruleID := "100000"
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_waf_rule.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareWAFRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWAFRuleConfig(zoneID, ruleID, "simulate", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "rule_id", ruleID),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttrSet(name, "package_id"),
					resource.TestCheckResourceAttrSet(name, "group_id"),
					resource.TestCheckResourceAttr(name, "mode", "simulate"),
				),
			},
			{
				Config: testAccCheckCloudflareWAFRuleConfig(zoneID, ruleID, "challenge", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "rule_id", ruleID),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttrSet(name, "package_id"),
					resource.TestCheckResourceAttrSet(name, "group_id"),
					resource.TestCheckResourceAttr(name, "mode", "challenge"),
				),
			},
		},
	})
}

func TestAccCloudflareWAFRule_CreateThenUpdate_SimpleModes(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	ruleID := "950000"
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_waf_rule.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareWAFRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWAFRuleConfig(zoneID, ruleID, "on", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "rule_id", ruleID),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttrSet(name, "package_id"),
					resource.TestCheckResourceAttrSet(name, "group_id"),
					resource.TestCheckResourceAttr(name, "mode", "on"),
				),
			},
			{
				Config: testAccCheckCloudflareWAFRuleConfig(zoneID, ruleID, "off", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "rule_id", ruleID),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttrSet(name, "package_id"),
					resource.TestCheckResourceAttrSet(name, "group_id"),
					resource.TestCheckResourceAttr(name, "mode", "off"),
				),
			},
		},
	})
}

func testAccCheckCloudflareWAFRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_waf_rule" {
			continue
		}

		rule, err := client.WAFRule(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.Attributes["package_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if rule.Mode != "default" && rule.Mode != "on" {
			return fmt.Errorf("Expected mode to be reset to default, got: %s", rule.Mode)
		}
	}

	return nil
}

func testAccCheckCloudflareWAFRuleConfig(zoneID, ruleID, mode, name string) string {
	return fmt.Sprintf(`
				resource "cloudflare_waf_rule" "%[4]s" {
					rule_id = %[2]s
					zone_id = "%[1]s"
					mode = "%[3]s"
				}`, zoneID, ruleID, mode, name)
}
