package cloudflare

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAccessRuleASN(t *testing.T) {
	name := "cloudflare_access_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessRuleAccountConfig("challenge", "this is notes", "asn", "AS112"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "notes", "this is notes"),
					resource.TestCheckResourceAttr(name, "mode", "challenge"),
					resource.TestCheckResourceAttr(name, "configuration.target", "asn"),
					resource.TestCheckResourceAttr(name, "configuration.value", "AS112"),
				),
			},
		},
	})
}

func TestAccAccessRuleIPRange(t *testing.T) {
	name := "cloudflare_access_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccessRuleAccountConfig("challenge", "this is notes", "ip_range", "104.16.0.0/20"),
				ExpectError: regexp.MustCompile("ip_range with ipv4 address must be a /16 or /24, got a /20"),
			}, {
				Config: testAccessRuleAccountConfig("challenge", "this is notes", "ip_range", "104.16.0.0/24"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "notes", "this is notes"),
					resource.TestCheckResourceAttr(name, "mode", "challenge"),
					resource.TestCheckResourceAttr(name, "configuration.target", "ip_range"),
					resource.TestCheckResourceAttr(name, "configuration.value", "104.16.0.0/24"),
				),
			},
		},
	})
}

func testAccessRuleAccountConfig(mode, notes, target, value string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_rule" "test" {
  notes = "%[2]s"
  mode = "%[1]s"
  configuration =	 {
    target = "%[3]s"
    value = "%[4]s"
  }
}`, mode, notes, target, value)
}

func TestValidateAccessRuleConfigurationIPRange(t *testing.T) {
	ipRangeValid := map[string]bool{
		"192.168.0.1/32":           false,
		"192.168.0.1/24":           true,
		"192.168.0.1/64":           false,
		"192.168.0.1/31":           false,
		"192.168.0.1/16":           true,
		"fd82:0f75:cf0d:d7b3::/64": true,
		"fd82:0f75:cf0d:d7b3::/48": true,
		"fd82:0f75:cf0d:d7b3::/32": true,
		"fd82:0f75:cf0d:d7b3::/63": false,
		"fd82:0f75:cf0d:d7b3::/16": false,
	}

	for ipRange, valid := range ipRangeValid {
		warnings, errors := validateAccessRuleConfigurationIPRange(ipRange)
		isValid := len(errors) == 0
		if len(warnings) != 0 {
			t.Fatalf("ipRange is either invalid or valid, no room for warnings")
		}
		if isValid != valid {
			t.Fatalf("%s resulted in %v, expected %v", ipRange, isValid, valid)
		}
	}
}
