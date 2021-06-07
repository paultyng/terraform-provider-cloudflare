package cloudflare

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareCustomSSL_Basic(t *testing.T) {
	var customSSL cloudflare.ZoneCustomSSL
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_ssl." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareCustomSSLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomSSLCertBasic(zoneID, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareCustomSSLExists(resourceName, &customSSL),
					resource.TestCheckResourceAttr(
						resourceName, "zone_id", zoneID),
					resource.TestMatchResourceAttr(
						resourceName, "priority", regexp.MustCompile("^[0-9]\\d*$")),
					resource.TestCheckResourceAttr(
						resourceName, "status", "active"),
					resource.TestMatchResourceAttr(
						resourceName, "zone_id", regexp.MustCompile("^[a-z0-9]{32}$")),
					resource.TestCheckResourceAttr(
						resourceName, "custom_ssl_options.bundle_method", "ubiquitous"),
					resource.TestCheckResourceAttr(
						resourceName, "custom_ssl_options.type", "legacy_custom"),
				),
			},
		},
	})
}

func testAccCheckCloudflareCustomSSLCertBasic(zoneID string, rName string) string {
	return fmt.Sprintf(`
resource "cloudflare_custom_ssl" "%[2]s" {
  zone_id = "%[1]s"
  custom_ssl_options = {
    certificate = "-----BEGIN CERTIFICATE-----\n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\nVYjN0hziUIEK9SutX1ZcpzGatcFUpw==\n-----END CERTIFICATE-----\n"
    private_key = "-----BEGIN RSA PRIVATE KEY-----\n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\n-----END RSA PRIVATE KEY-----\n"
    bundle_method = "ubiquitous",
    geo_restrictions = "us"
    type = "legacy_custom"
  }
}`, zoneID, rName)
}

func testAccCheckCloudflareCustomSSLDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_custom_ssl" {
			continue
		}

		err := client.DeleteSSL(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("cert still exists")
		}
	}

	return nil
}

func testAccCheckCloudflareCustomSSLExists(n string, customSSL *cloudflare.ZoneCustomSSL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No cert ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		foundCustomSSL, err := client.SSLDetails(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundCustomSSL.ID != rs.Primary.ID {
			return fmt.Errorf("cert not found")
		}

		*customSSL = foundCustomSSL

		return nil
	}
}

func TestAccCloudflareCustomSSLWithEmptyGeoRestrictions(t *testing.T) {
	var customSSL cloudflare.ZoneCustomSSL
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_ssl." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareCustomSSLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomSSLWithEmptyGeoRestrictions(zoneID, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareCustomSSLExists(resourceName, &customSSL),
					resource.TestCheckNoResourceAttr(resourceName, "custom_ssl_options.geo_restrictions"),
				),
			},
		},
	})
}

func testAccCheckCloudflareCustomSSLWithEmptyGeoRestrictions(zoneID string, rName string) string {
	return fmt.Sprintf(`
resource "cloudflare_custom_ssl" "%[2]s" {
  zone_id = "%[1]s"
  custom_ssl_options = {
    certificate = "-----BEGIN CERTIFICATE-----\n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\nVYjN0hziUIEK9SutX1ZcpzGatcFUpw==\n-----END CERTIFICATE-----\n"
    private_key = "-----BEGIN RSA PRIVATE KEY-----\n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\n-----END RSA PRIVATE KEY-----\n"
    bundle_method = "ubiquitous"
    type = "legacy_custom"
  }
}`, zoneID, rName)
}
