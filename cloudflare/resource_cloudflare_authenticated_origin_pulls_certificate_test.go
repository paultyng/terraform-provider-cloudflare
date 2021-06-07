package cloudflare

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("cloudflare_authenticated_origin_pulls_certificate", &resource.Sweeper{
		Name: "cloudflare_authenticated_origin_pulls_certificate",
		F:    testSweepCloudflareAuthenticatdOriginPullsCertificates,
	})
}

func testSweepCloudflareAuthenticatdOriginPullsCertificates(r string) error {
	client, clientErr := sharedClient()
	if clientErr != nil {
		log.Printf("[ERROR] Failed to create Cloudflare client: %s", clientErr)
	}

	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	perZoneCertificates, certsErr := client.ListPerZoneAuthenticatedOriginPullsCertificates(context.Background(), zoneID)

	if certsErr != nil {
		log.Printf("[ERROR] Failed to fetch Cloudflare authenticated origin pull certificates: %s", certsErr)
	}

	if len(perZoneCertificates) == 0 {
		log.Print("[DEBUG] No Cloudflare authenticated origin pull certificates to sweep")
		return nil
	}

	for _, certificate := range perZoneCertificates {
		_, err := client.DeletePerZoneAuthenticatedOriginPullsCertificate(context.Background(), zoneID, certificate.ID)

		if err != nil {
			log.Printf("[ERROR] Failed to delete Cloudflare authenticated origin pull certificate (%s) in zone ID: %s", certificate.ID, zoneID)
		}
	}

	return nil
}

func TestAccCloudflareAuthenticatedOriginPullsCertificatePerZone(t *testing.T) {
	var perZoneAOP cloudflare.PerZoneAuthenticatedOriginPullsCertificateDetails
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_authenticated_origin_pulls_certificate.%s", rnd)
	aopType := "per-zone"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAuthenticatedOriginPullsCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareAuthenticatedOriginPullsCertificateConfig(zoneID, rnd, aopType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAuthenticatedOriginPullsCertificatePerZoneExists(name, &perZoneAOP),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "type", aopType),
				),
			},
		},
	})
}

func TestAccCloudflareAuthenticatedOriginPullsCertificatePerHostname(t *testing.T) {
	var perZoneAOP cloudflare.PerHostnameAuthenticatedOriginPullsCertificateDetails
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_authenticated_origin_pulls_certificate.%s", rnd)
	aopType := "per-hostname"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAuthenticatedOriginPullsCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareAuthenticatedOriginPullsCertificateConfig(zoneID, rnd, aopType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAuthenticatedOriginPullsCertificatePerHostnameExists(name, &perZoneAOP),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "type", aopType),
				),
			},
		},
	})
}

func testAccCheckCloudflareAuthenticatedOriginPullsCertificatePerZoneExists(n string, perZoneAOPCert *cloudflare.PerZoneAuthenticatedOriginPullsCertificateDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No cert ID is set")
		}
		client := testAccProvider.Meta().(*cloudflare.API)
		foundPerZoneAOPCert, err := client.GetPerZoneAuthenticatedOriginPullsCertificateDetails(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}
		if foundPerZoneAOPCert.ID != rs.Primary.ID {
			return fmt.Errorf("cert not found")
		}
		*perZoneAOPCert = foundPerZoneAOPCert
		return nil
	}
}

func testAccCheckCloudflareAuthenticatedOriginPullsCertificatePerHostnameExists(n string, perHostnameAOPCert *cloudflare.PerHostnameAuthenticatedOriginPullsCertificateDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No cert ID is set")
		}
		client := testAccProvider.Meta().(*cloudflare.API)
		foundPerHostnameAOPCert, err := client.GetPerHostnameAuthenticatedOriginPullsCertificate(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}
		if foundPerHostnameAOPCert.ID != rs.Primary.ID {
			return fmt.Errorf("cert not found")
		}
		*perHostnameAOPCert = foundPerHostnameAOPCert
		return nil
	}
}

func testAccCheckCloudflareAuthenticatedOriginPullsCertificateConfig(zoneID, name, aopType string) string {
	return fmt.Sprintf(`
  resource "cloudflare_authenticated_origin_pulls_certificate" "%[2]s" {
	  zone_id = "%[1]s"
	  certificate = "-----BEGIN CERTIFICATE-----\nMIIEsTCCA5mgAwIBAgISA53fvg2BvlK2QXSkdZewcNo4MA0GCSqGSIb3DQEBCwUA\nMEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD\nExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0yMDA2MjUyMTAzNDdaFw0y\nMDA5MjMyMTAzNDdaMB4xHDAaBgNVBAMTE3RlcnJhZm9ybS5jZmFwaS5uZXQwdjAQ\nBgcqhkjOPQIBBgUrgQQAIgNiAASBYi00+H4E7uUeogweuutTWvuAz8TC6ClQYemH\nCGA6xKrvSgWwjhvVM9joPhGlbUDbINKhVMdZd7q3DgBinVu9GjjKf1Ajxnr6nEsK\naq37tZmtUFawbqnJHAI+O3uTan+jggJpMIICZTAOBgNVHQ8BAf8EBAMCB4AwHQYD\nVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0O\nBBYEFACS0TnEhBjGvOG127Yn2O1/UCOoMB8GA1UdIwQYMBaAFKhKamMEfd265tE5\nt6ZFZe/zqOyhMG8GCCsGAQUFBwEBBGMwYTAuBggrBgEFBQcwAYYiaHR0cDovL29j\nc3AuaW50LXgzLmxldHNlbmNyeXB0Lm9yZzAvBggrBgEFBQcwAoYjaHR0cDovL2Nl\ncnQuaW50LXgzLmxldHNlbmNyeXB0Lm9yZy8wHgYDVR0RBBcwFYITdGVycmFmb3Jt\nLmNmYXBpLm5ldDBMBgNVHSAERTBDMAgGBmeBDAECATA3BgsrBgEEAYLfEwEBATAo\nMCYGCCsGAQUFBwIBFhpodHRwOi8vY3BzLmxldHNlbmNyeXB0Lm9yZzCCAQUGCisG\nAQQB1nkCBAIEgfYEgfMA8QB3AF6nc/nfVsDntTZIfdBJ4DJ6kZoMhKESEoQYdZaB\ncUVYAAABcu2CH2EAAAQDAEgwRgIhAK4dA41POH3dCyi/5CN98MbBRAl8a6LyeQls\nJyZ+y1sIAiEAoMtsQKVgf8APT7/DGj/b4OzMO6EBKWcrGkZpTi7nyyQAdgCyHgXM\ni6LNiiBOh2b5K7mKJSBna9r6cOeySVMt74uQXgAAAXLtgh9PAAAEAwBHMEUCIQC1\nnxSRx2fcqG8gw5z0QK5PGktggqIulg2Jrwr20ZfXKwIgGxNlOEucj1t71h4PaLuy\nnBigJo57ztE5t56o0dlUOzEwDQYJKoZIhvcNAQELBQADggEBACy8MS07SVQLMeGK\na3E7jn7mQciQkt063tnIYbvnUTeYQZVe1Rzk6Tm9GyQoL7MIFAvTHbsB9bNzIRrl\nubefCn4s6PHnVyDGiPY/yQgGjymXyxcsfwVnc3XO3i6N8AN1MQuKMx+Kx69sHVpa\nKq9Qlu1HlStlX/eUWMcoDk1WaCJ7xm17npvdWDweDg71Qlgnl6ukggN+cQwKepw5\n4tMnqmhrzMH+xnH2dTIQ10lgB31AlwBSbOUymhg8XN+BIeXW54mBjdxkBd++7+0q\nv7oFDmljpwQSAC2BMU8ah7lwRhQxgTrG0z10Qdje1CJ8ylRHArIeISlx+jBAwKQh\nulkb7Ck=\n-----END CERTIFICATE-----\n"
	  private_key = "-----BEGIN EC PRIVATE KEY-----\nMIGkAgEBBDD+Um5v/lCBTCvHEcZlLnSz6XX1fEOk5FxfUdiQvcY5x6WXuu3dDgDf\nvKIS0J6AsxygBwYFK4EEACKhZANiAASBYi00+H4E7uUeogweuutTWvuAz8TC6ClQ\nYemHCGA6xKrvSgWwjhvVM9joPhGlbUDbINKhVMdZd7q3DgBinVu9GjjKf1Ajxnr6\nnEsKaq37tZmtUFawbqnJHAI+O3uTan8=\n-----END EC PRIVATE KEY-----\n"
      type = "%[3]s"
  }`, zoneID, name, aopType)
}

func testAccCheckCloudflareAuthenticatedOriginPullsCertificateDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)
	for _, rs := range s.RootModule().Resources {
		if rs.Primary.Attributes["type"] == "per-zone" {
			_, err := client.DeletePerZoneAuthenticatedOriginPullsCertificate(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Error deleting Per-Zone AOP certificate on zone %q: %s", zoneID, err)
			}
		} else if rs.Primary.Attributes["type"] == "per-hostname" {
			_, err := client.DeletePerZoneAuthenticatedOriginPullsCertificate(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Error deleting Per-Zone AOP certificate on zone %q: %s", zoneID, err)
			}
		}
	}
	return nil
}
