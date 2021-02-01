package cloudflare

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccCloudflareCustomSSL_Basic(t *testing.T) {
	acctest.UseBinaryDriver("acme", Provider)
	acctest.UseBinaryDriver("tls", Provider)

	t.Parallel()
	var customSSL cloudflare.ZoneCustomSSL
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	email := os.Getenv("CLOUDFLARE_EMAIL")

	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_ssl." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"acme": {},
			"tls":  {},
		},
		CheckDestroy: testAccCheckCloudflareCustomSSLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomSSLCertBasic(zoneID, rnd, domain, email),
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

func testAccCheckCloudflareCustomSSLCertBasic(zoneID, rName, domain, email string) string {
	return fmt.Sprintf(`
provider "acme" {
	server_url = "https://acme-v02.api.letsencrypt.org/directory"
}

resource "tls_private_key" "private_key" {
  algorithm = "RSA"
}

resource "acme_registration" "reg" {
  account_key_pem = "${tls_private_key.private_key.private_key_pem}"
  email_address   = "%[4]s"
}

resource "acme_certificate" "certificate" {
  account_key_pem           = "${acme_registration.reg.account_key_pem}"
  common_name               = "%[3]s"

  dns_challenge {
    provider = "cloudflare"
  }
}

resource "cloudflare_custom_ssl" "%[2]s" {
  zone_id = "%[1]s"
  custom_ssl_options = {
    certificate = "${acme_certificate.certificate.certificate_pem}"
    private_key = "${acme_certificate.certificate.private_key_pem}"
    bundle_method = "ubiquitous",
    geo_restrictions = "us"
    type = "legacy_custom"
  }
}`, zoneID, rName, domain, email)
}

func testAccCheckCloudflareCustomSSLDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_custom_ssl" {
			continue
		}

		err := client.DeleteSSL(rs.Primary.Attributes["zone_id"], rs.Primary.ID)
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
		foundCustomSSL, err := client.SSLDetails(rs.Primary.Attributes["zone_id"], rs.Primary.ID)
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
	t.Parallel()
	var customSSL cloudflare.ZoneCustomSSL
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	email := os.Getenv("CLOUDFLARE_EMAIL")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_ssl." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareCustomSSLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomSSLWithEmptyGeoRestrictions(zoneID, rnd, domain, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareCustomSSLExists(resourceName, &customSSL),
					resource.TestCheckNoResourceAttr(resourceName, "custom_ssl_options.geo_restrictions"),
				),
			},
		},
	})
}

func testAccCheckCloudflareCustomSSLWithEmptyGeoRestrictions(zoneID, rName, domain, email string) string {
	return fmt.Sprintf(`
provider "acme" {
	server_url = "https://acme-v02.api.letsencrypt.org/directory"
}

resource "tls_private_key" "private_key" {
  algorithm = "RSA"
}

resource "acme_registration" "reg" {
  account_key_pem = "${tls_private_key.private_key.private_key_pem}"
  email_address   = "%[4]s"
}

resource "acme_certificate" "certificate" {
  account_key_pem           = "${acme_registration.reg.account_key_pem}"
  common_name               = "%[3]s"

  dns_challenge {
    provider = "cloudflare"
  }
}

resource "cloudflare_custom_ssl" "%[2]s" {
  zone_id = "%[1]s"
  custom_ssl_options = {
    certificate = "${acme_certificate.certificate.certificate_pem}"
    private_key = "${acme_certificate.certificate.private_key_pem}"
    bundle_method = "ubiquitous",
    type = "legacy_custom"
  }
}`, zoneID, rName, domain, email)
}
