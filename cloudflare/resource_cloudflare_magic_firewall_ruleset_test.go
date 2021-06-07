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

func TestAccCloudflareMagicFirewallRulesetExists(t *testing.T) {
	skipMagicTransitTestForNonConfiguredDefaultZone(t)

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_magic_firewall_ruleset.%s", rnd)
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	var MagicFirewallRuleset cloudflare.MagicFirewallRuleset

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckAccount(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareMagicFirewallRulesetSimple(rnd, rnd, rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareMagicFirewallRulesetExists(name, &MagicFirewallRuleset),
					resource.TestCheckResourceAttr(
						name, "name", rnd),
				),
			},
		},
	})
}

func TestAccCloudflareMagicFirewallRulesetUpdateName(t *testing.T) {
	skipMagicTransitTestForNonConfiguredDefaultZone(t)

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_magic_firewall_ruleset.%s", rnd)
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	var MagicFirewallRuleset cloudflare.MagicFirewallRuleset
	var initialID string

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckAccount(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareMagicFirewallRulesetSimple(rnd, rnd, rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareMagicFirewallRulesetExists(name, &MagicFirewallRuleset),
					resource.TestCheckResourceAttr(
						name, "name", rnd),
				),
			},
			{
				PreConfig: func() {
					initialID = MagicFirewallRuleset.ID
				},
				Config: testAccCheckCloudflareMagicFirewallRulesetSimple(rnd, rnd+"-updated", rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareMagicFirewallRulesetExists(name, &MagicFirewallRuleset),
					func(state *terraform.State) error {
						if initialID == MagicFirewallRuleset.ID {
							return fmt.Errorf("forced recreation but Magic Firewall Ruleset got updated (id %q)",
								MagicFirewallRuleset.ID)
						}
						return nil
					},
					resource.TestCheckResourceAttr(
						name, "name", rnd+"-updated"),
				),
			},
		},
	})
}

func TestAccCloudflareMagicFirewallRulesetUpdateDescription(t *testing.T) {
	skipMagicTransitTestForNonConfiguredDefaultZone(t)

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_magic_firewall_ruleset.%s", rnd)
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	var MagicFirewallRuleset cloudflare.MagicFirewallRuleset
	var initialID string

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckAccount(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareMagicFirewallRulesetSimple(rnd, rnd, rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareMagicFirewallRulesetExists(name, &MagicFirewallRuleset),
					resource.TestCheckResourceAttr(
						name, "description", rnd),
				),
			},
			{
				PreConfig: func() {
					initialID = MagicFirewallRuleset.ID
				},
				Config: testAccCheckCloudflareMagicFirewallRulesetSimple(rnd, rnd, rnd+"-updated", accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareMagicFirewallRulesetExists(name, &MagicFirewallRuleset),
					func(state *terraform.State) error {
						if initialID != MagicFirewallRuleset.ID {
							return fmt.Errorf("wanted update but Magic Firewall Ruleset got recreated (id changed %q -> %q)",
								initialID, MagicFirewallRuleset.ID)
						}
						return nil
					},
					resource.TestCheckResourceAttr(
						name, "description", rnd+"-updated"),
				),
			},
		},
	})
}

func TestAccCloudflareMagicFirewallRulesetSingleRule(t *testing.T) {
	skipMagicTransitTestForNonConfiguredDefaultZone(t)

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_magic_firewall_ruleset.%s", rnd)
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	var MagicFirewallRuleset cloudflare.MagicFirewallRuleset

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckAccount(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareMagicFirewallRulesetSingleRule(rnd, rnd, rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareMagicFirewallRulesetExists(name, &MagicFirewallRuleset),
					resource.TestCheckResourceAttr(
						name, "name", rnd),
					resource.TestCheckResourceAttr(name, "rules.#", "1"),
					resource.TestCheckResourceAttr(name, "rules.0.action", "allow"),
					resource.TestCheckResourceAttr(name, "rules.0.description", "Allow TCP Ephemeral Ports"),
					resource.TestCheckResourceAttr(name, "rules.0.enabled", "true"),
					resource.TestCheckResourceAttr(name, "rules.0.expression", "tcp.dstport in { 32768..65535 }"),
				),
			},
		},
	})
}

func TestAccCloudflareMagicFirewallRulesetUpdateWithHigherPriority(t *testing.T) {
	skipMagicTransitTestForNonConfiguredDefaultZone(t)

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_magic_firewall_ruleset.%s", rnd)
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	var MagicFirewallRuleset cloudflare.MagicFirewallRuleset

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckAccount(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareMagicFirewallRulesetSingleRule(rnd, rnd, rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareMagicFirewallRulesetExists(name, &MagicFirewallRuleset),
					resource.TestCheckResourceAttr(
						name, "name", rnd),
					resource.TestCheckResourceAttr(name, "rules.#", "1"),
					resource.TestCheckResourceAttr(name, "rules.0.action", "allow"),
					resource.TestCheckResourceAttr(name, "rules.0.description", "Allow TCP Ephemeral Ports"),
					resource.TestCheckResourceAttr(name, "rules.0.enabled", "true"),
					resource.TestCheckResourceAttr(name, "rules.0.expression", "tcp.dstport in { 32768..65535 }"),
				),
			},
			{
				Config: testAccCheckCloudflareMagicFirewallRulesetUpdateWithHigherPriority(rnd, rnd, rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareMagicFirewallRulesetExists(name, &MagicFirewallRuleset),
					resource.TestCheckResourceAttr(
						name, "name", rnd),
					resource.TestCheckResourceAttr(name, "rules.#", "2"),
					resource.TestCheckResourceAttr(name, "rules.0.action", "block"),
					resource.TestCheckResourceAttr(name, "rules.0.description", "Block UDP Ephemeral Ports"),
					resource.TestCheckResourceAttr(name, "rules.0.enabled", "true"),
					resource.TestCheckResourceAttr(name, "rules.0.expression", "udp.dstport in { 32768..65535 }"),
					resource.TestCheckResourceAttr(name, "rules.1.action", "allow"),
					resource.TestCheckResourceAttr(name, "rules.1.description", "Allow TCP Ephemeral Ports"),
					resource.TestCheckResourceAttr(name, "rules.1.enabled", "true"),
					resource.TestCheckResourceAttr(name, "rules.1.expression", "tcp.dstport in { 32768..65535 }"),
				),
			},
		},
	})
}

func testAccCheckCloudflareMagicFirewallRulesetExists(n string, ruleset *cloudflare.MagicFirewallRuleset) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Magic Firewall Ruleset is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		foundRuleset, err := client.GetMagicFirewallRuleset(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*ruleset = foundRuleset

		return nil
	}
}

func testAccCheckCloudflareMagicFirewallRulesetSimple(ID, name, description, accountID string) string {
	return fmt.Sprintf(`
  resource "cloudflare_magic_firewall_ruleset" "%[1]s" {
	account_id = "%[4]s"
    name = "%[2]s"
	description = "%[3]s"
  }`, ID, name, description, accountID)
}

func testAccCheckCloudflareMagicFirewallRulesetSingleRule(ID, name, description, accountID string) string {
	return fmt.Sprintf(`
  resource "cloudflare_magic_firewall_ruleset" "%[1]s" {
	account_id = "%[4]s"
    name = "%[2]s"
	description = "%[3]s"
    rules = [
      {
        action = "allow"
        expression = "tcp.dstport in { 32768..65535 }"
        description = "Allow TCP Ephemeral Ports"
        enabled = "true"
      }
    ]
  }`, ID, name, description, accountID)
}

func testAccCheckCloudflareMagicFirewallRulesetUpdateWithHigherPriority(ID, name, description, accountID string) string {
	return fmt.Sprintf(`
  resource "cloudflare_magic_firewall_ruleset" "%[1]s" {
	account_id = "%[4]s"
    name = "%[2]s"
	description = "%[3]s"
    rules = [
      {
        action = "block"
        expression = "udp.dstport in { 32768..65535 }"
        description = "Block UDP Ephemeral Ports"
        enabled = "true"
      },
      {
        action = "allow"
        expression = "tcp.dstport in { 32768..65535 }"
        description = "Allow TCP Ephemeral Ports"
        enabled = "true"
      }
    ]
  }`, ID, name, description, accountID)
}
