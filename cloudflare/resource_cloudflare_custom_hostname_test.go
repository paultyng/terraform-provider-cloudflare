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

func TestAccCloudflareCustomHostnameBasic(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_hostname." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomHostnameBasic(zoneID, rnd, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.method", "txt"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.value"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.type"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.name"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification_http.http_url"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification_http.http_body"),
				),
			},
		},
	})
}

func testAccCheckCloudflareCustomHostnameBasic(zoneID, rnd, domain string) string {
	return fmt.Sprintf(`
resource "cloudflare_custom_hostname" "%[2]s" {
  zone_id = "%[1]s"
  hostname = "%[2]s.%[3]s"
  ssl {
    method = "txt"
  }
}
`, zoneID, rnd, domain)
}

func TestAccCloudflareCustomHostnameWithCustomOriginServer(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_hostname." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomHostnameWithCustomOriginServer(zoneID, rnd, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(resourceName, "custom_origin_server", fmt.Sprintf("origin.%s.terraform.cfapi.net", rnd)),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.method", "txt"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.value"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.type"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.name"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification_http.http_url"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification_http.http_body"),
				),
			},
		},
	})
}

func testAccCheckCloudflareCustomHostnameWithCustomOriginServer(zoneID, rnd, domain string) string {
	return fmt.Sprintf(`
resource "cloudflare_custom_hostname" "%[2]s" {
  zone_id = "%[1]s"
  hostname = "%[2]s.%[3]s"
  custom_origin_server = "origin.%[2]s.terraform.cfapi.net"
  ssl {
    method = "txt"
  }
}

resource "cloudflare_record" "%[2]s" {
  zone_id = "%[1]s"
  name    = "origin.%[2]s.terraform.cfapi.net"
  value   = "example.com"
  type    = "CNAME"
  ttl     = 3600
}`, zoneID, rnd, domain)
}

func TestAccCloudflareCustomHostnameWithHTTPValidation(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_hostname." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomHostnameWithHTTPValidation(zoneID, rnd, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.method", "http"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.value"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.type"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.name"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification_http.http_url"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification_http.http_body"),
				),
			},
		},
	})
}

func testAccCheckCloudflareCustomHostnameWithHTTPValidation(zoneID, rnd, domain string) string {
	return fmt.Sprintf(`
resource "cloudflare_custom_hostname" "%[2]s" {
  zone_id = "%[1]s"
  hostname = "%[2]s.%[3]s"
  ssl {
    method = "http"
  }
}
`, zoneID, rnd, domain)
}

func TestAccCloudflareCustomHostnameWithCustomSSLSettings(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_hostname." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomHostnameWithCustomSSLSettings(zoneID, rnd, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.http2", "off"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.min_tls_version", "1.2"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.ciphers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.ciphers.0", "ECDHE-RSA-AES128-GCM-SHA256"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.ciphers.1", "AES128-SHA"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.value"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.type"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification.name"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification_http.http_url"),
					resource.TestCheckResourceAttrSet(resourceName, "ownership_verification_http.http_body"),
				),
			},
		},
	})
}

func testAccCheckCloudflareCustomHostnameWithCustomSSLSettings(zoneID, rnd, domain string) string {
	return fmt.Sprintf(`
resource "cloudflare_custom_hostname" "%[2]s" {
  zone_id = "%[1]s"
  hostname = "%[2]s.%[3]s"
  ssl {
    method = "http"
    settings {
      http2 = "off"
      min_tls_version = "1.2"
      ciphers = [
        "ECDHE-RSA-AES128-GCM-SHA256",
        "AES128-SHA"
      ]
    }
  }
}
`, zoneID, rnd, domain)
}

func TestAccCloudflareCustomHostnameUpdate(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_hostname." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomHostnameWithCustomSSLSettings(zoneID, rnd, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.http2", "off"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.min_tls_version", "1.2"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.ciphers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.ciphers.0", "ECDHE-RSA-AES128-GCM-SHA256"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.ciphers.1", "AES128-SHA"),
				),
			},
			{
				Config: testAccCheckCloudflareCustomHostnameWithCustomSSLSettingsUpdated(zoneID, rnd, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.http2", "off"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.min_tls_version", "1.1"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.ciphers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.ciphers.0", "ECDHE-RSA-AES128-GCM-SHA256"),
					resource.TestCheckResourceAttr(resourceName, "ssl.0.settings.0.ciphers.1", "AES128-SHA"),
				),
			},
		},
	})
}

func testAccCheckCloudflareCustomHostnameWithCustomSSLSettingsUpdated(zoneID, rnd, domain string) string {
	return fmt.Sprintf(`
resource "cloudflare_custom_hostname" "%[2]s" {
  zone_id = "%[1]s"
  hostname = "%[2]s.%[3]s"
  ssl {
    method = "http"
    settings {
      http2 = "off"
      min_tls_version = "1.1"
      ciphers = [
        "ECDHE-RSA-AES128-GCM-SHA256",
        "AES128-SHA"
      ]
    }
  }
}
`, zoneID, rnd, domain)
}

func TestAccCloudflareCustomHostname_UpdatingZoneForcesNewResource(t *testing.T) {
	t.Parallel()

	var before, after cloudflare.CustomHostname
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	altZoneID := os.Getenv("CLOUDFLARE_ALT_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	altDomain := os.Getenv("CLOUDFLARE_ALT_DOMAIN")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_hostname." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAltZoneID(t)
			testAccPreCheckAltDomain(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomHostnameBasic(zoneID, rnd, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareCustomHostnameExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("%s.%s", rnd, domain)),
				),
			},
			{
				Config: testAccCheckCloudflareCustomHostnameBasic(altZoneID, rnd, altDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareCustomHostnameExists(resourceName, &after),
					testAccCheckCloudflareCustomHostnameRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "zone_id", altZoneID),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("%s.%s", rnd, altDomain)),
				),
			},
		},
	})
}

func TestAccCloudflareCustomHostnameImport(t *testing.T) {
	t.Parallel()

	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_custom_hostname." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareCustomHostnameBasic(zoneID, rnd, domain),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdPrefix:     fmt.Sprintf("%s/", zoneID),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ssl.#", "ssl.0.certificate_authority", "ssl.0.cname_name", "ssl.0.cname_target", "ssl.0.custom_certificate", "ssl.0.custom_key", "ssl.0.method", "ssl.0.status", "ssl.0.type", "ssl.0.wildcard"},
			},
		},
	})
}

func testAccCheckCloudflareCustomHostnameRecreated(before, after *cloudflare.CustomHostname) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID == after.ID {
			return fmt.Errorf("Expected change of CustomHostname Ids, but both were %v", before.ID)
		}
		return nil
	}
}

func testAccCheckCloudflareCustomHostnameExists(n string, customHostname *cloudflare.CustomHostname) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CustomHostname ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		foundCustomHostname, err := client.CustomHostname(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundCustomHostname.ID != rs.Primary.ID {
			return fmt.Errorf("CustomHostname not found")
		}

		*customHostname = foundCustomHostname

		return nil
	}
}
