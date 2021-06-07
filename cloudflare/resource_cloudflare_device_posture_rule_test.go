package cloudflare

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareDevicePostureRuleBasic(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_device_posture_rule.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareDevicePostureRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareDevicePostureRuleConfigBasic(rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "type", "serial_number"),
					resource.TestCheckResourceAttr(name, "description", "My description"),
					resource.TestCheckResourceAttr(name, "match.0.platform", "windows"),
					resource.TestCheckResourceAttr(name, "input.0.id", "asdf-123"),
				),
			},
		},
	})
}

func testAccCloudflareDevicePostureRuleConfigBasic(rnd, accountID string) string {
	return fmt.Sprintf(`
resource "cloudflare_device_posture_rule" "%[1]s" {
	account_id                = "%[2]s"
	name                      = "%[1]s"
	type                      = "serial_number"
	description               = "My description"
	schedule                  = "24h"
	match {
		platform = "windows"
	}
	input {
		id = "asdf-123"
	}
}
`, rnd, accountID)
}

func testAccCheckCloudflareDevicePostureRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_device_posture_rule" {
			continue
		}

		_, err := client.DevicePostureRule(context.Background(), rs.Primary.Attributes["account_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Device Posture Rule still exists")
		}
	}

	return nil
}
