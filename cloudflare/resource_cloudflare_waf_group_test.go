package cloudflare

import (
	"fmt"
	"os"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccCloudflareWAFGroup_CreateThenUpdate(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	groupID, err := testAccGetWAFGroup(zoneID)
	if err != nil {
		t.Errorf(err.Error())
	}

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_waf_group.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareWAFGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWAFGroupConfig(zoneID, groupID, "on", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "group_id", groupID),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttrSet(name, "package_id"),
					resource.TestCheckResourceAttr(name, "mode", "on"),
				),
			},
			{
				Config: testAccCheckCloudflareWAFGroupConfig(zoneID, groupID, "off", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "group_id", groupID),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttrSet(name, "package_id"),
					resource.TestCheckResourceAttr(name, "mode", "off"),
				),
			},
		},
	})
}

func testAccGetWAFGroup(zoneID string) (string, error) {
	config := Config{}
	if apiToken, ok := os.LookupEnv("CLOUDFLARE_API_TOKEN"); ok {
		config.APIToken = apiToken
	} else if apiKey, ok := os.LookupEnv("CLOUDFLARE_API_KEY"); ok {
		config.APIKey = apiKey
		if email, ok := os.LookupEnv("CLOUDFLARE_EMAIL"); ok {
			config.Email = email
		} else {
			return "", fmt.Errorf("Cloudflare email is not set correctly")
		}
	} else {
		return "", fmt.Errorf("Cloudflare credentials are not set correctly")
	}

	client, err := config.Client()
	if err != nil {
		return "", err
	}

	pkgList, err := client.ListWAFPackages(zoneID)
	if err != nil {
		return "", fmt.Errorf("Error while listing WAF packages: %s", err)
	}

	for _, pkg := range pkgList {
		groupList, err := client.ListWAFGroups(zoneID, pkg.ID)
		if err != nil {
			return "", fmt.Errorf("Error while listing WAF groups for WAF package %s: %s", pkg.ID, err)
		}

		for _, group := range groupList {
			return group.ID, nil
		}
	}

	return "", fmt.Errorf("No group found")
}

func testAccCheckCloudflareWAFGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_waf_group" {
			continue
		}

		group, err := client.WAFGroup(rs.Primary.Attributes["zone_id"], rs.Primary.Attributes["package_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if group.Mode != "on" {
			return fmt.Errorf("Expected mode to be reset to on, got: %s", group.Mode)
		}
	}

	return nil
}

func testAccCheckCloudflareWAFGroupConfig(zoneID, groupID, mode, name string) string {
	return fmt.Sprintf(`
				resource "cloudflare_waf_group" "%[4]s" {
					zone_id = "%[1]s"
					group_id = "%[2]s"
					mode = "%[3]s"
				}`, zoneID, groupID, mode, name)
}
