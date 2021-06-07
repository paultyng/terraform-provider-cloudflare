package cloudflare

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("cloudflare_firewall_rule", &resource.Sweeper{
		Name: "cloudflare_firewall_rule",
		F:    testSweepCloudflareFirewallRuleSweeper,
	})
}

func testSweepCloudflareFirewallRuleSweeper(r string) error {
	client, clientErr := sharedClient()
	if clientErr != nil {
		log.Printf("[ERROR] Failed to create Cloudflare client: %s", clientErr)
	}

	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rules, rulesErr := client.FirewallRules(context.Background(), zoneID, cloudflare.PaginationOptions{})

	if rulesErr != nil {
		log.Printf("[ERROR] Failed to fetch Cloudflare firewall rules: %s", rulesErr)
	}

	for _, rule := range rules {
		err := client.DeleteFirewallRule(context.Background(), zoneID, rule.ID)

		if err != nil {
			log.Printf("[ERROR] Failed to delete Cloudflare firewall rule (%s) in zone ID: %s", rule.ID, zoneID)
		}
	}

	return nil
}

func TestAccFirewallRuleSimple(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_firewall_rule." + rnd
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	filterQuoted := `(http.request.uri.path ~ \".*wp-login-` + rnd + `.php\" or http.request.uri.path ~ \".*xmlrpc.php\") and ip.src ne 192.0.2.1`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testFirewallRuleConfig(rnd, zoneID, "true", "this is notes", filterQuoted, "allow", "1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "description", "this is notes"),
					resource.TestCheckResourceAttr(name, "paused", "true"),
					resource.TestCheckResourceAttr(name, "action", "allow"),
					resource.TestCheckResourceAttr(name, "priority", "1"),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
				),
			},
		},
	})
}

func testFirewallRuleConfig(resourceID, zoneID, paused, description, expression, action, priority string) string {
	return fmt.Sprintf(`
		resource "cloudflare_filter" "%[1]s" {
		  zone_id = "%[2]s"
		  paused = "%[3]s"
		  description = "%[4]s"
		  expression = "%[5]s"
		}
		resource "cloudflare_firewall_rule" "%[1]s" {
		  zone_id = "%[2]s"
		  paused = "%[3]s"
		  description = "%[4]s"
		  filter_id = "${cloudflare_filter.%[1]s.id}"
		  action = "%[6]s"
		  priority = %[7]s
		}
		`, resourceID, zoneID, paused, description, expression, action, priority)
}

func TestAccFirewallRuleBypass(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_firewall_rule." + rnd
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")

	filterQuoted := `(http.host eq \"` + domain + `\")`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testFirewallRuleBypassConfig(rnd, zoneID, "false", "this is notes", filterQuoted),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "description", "this is notes"),
					resource.TestCheckResourceAttr(name, "paused", "false"),
					resource.TestCheckResourceAttr(name, "action", "bypass"),
					resource.TestCheckResourceAttr(name, "priority", "2"),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "products.#", "2"),
				),
			},
		},
	})
}

func testFirewallRuleBypassConfig(resourceID, zoneID, paused, description, expression string) string {
	return fmt.Sprintf(`
		resource "cloudflare_filter" "%[1]s" {
		  zone_id = "%[2]s"
		  paused = "%[3]s"
		  description = "%[4]s"
		  expression = "%[5]s"
		}
		resource "cloudflare_firewall_rule" "%[1]s" {
		  zone_id = "%[2]s"
		  paused = "%[3]s"
		  description = "%[4]s"
		  filter_id = "${cloudflare_filter.%[1]s.id}"
		  action = "bypass"
		  priority = "2"
		  products = ["uaBlock", "waf"]
		}
		`, resourceID, zoneID, paused, description, expression)
}

func TestAccFirewallRuleWithUnicodeAndHTMLEntity(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_firewall_rule." + rnd
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")

	expression := `(http.host eq \"` + domain + `\")`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleWithUnicodeAndHTMLEntityConfig(rnd, zoneID, "true", "this is a 'test'", expression, "allow", "1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "description", "this is a \u0026#39;test\u0026#39;"),
					resource.TestCheckResourceAttr(name, "paused", "true"),
					resource.TestCheckResourceAttr(name, "action", "allow"),
					resource.TestCheckResourceAttr(name, "priority", "1"),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccFirewallRuleWithUnicodeAndHTMLEntityConfig(resourceID, zoneID, paused, description, expression, action, priority string) string {
	return fmt.Sprintf(`
		resource "cloudflare_filter" "%[1]s" {
		  zone_id = "%[2]s"
		  paused = "%[3]s"
		  description = "%[4]s"
		  expression = "%[5]s"
		}
		resource "cloudflare_firewall_rule" "%[1]s" {
		  zone_id = "%[2]s"
		  paused = "%[3]s"
		  description = "%[4]s"
		  filter_id = "${cloudflare_filter.%[1]s.id}"
		  action = "%[6]s"
		  priority = %[7]s
		}
		`, resourceID, zoneID, paused, description, expression, action, priority)
}
