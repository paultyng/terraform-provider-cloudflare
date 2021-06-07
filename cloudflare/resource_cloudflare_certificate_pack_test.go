package cloudflare

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("cloudflare_certificate_pack", &resource.Sweeper{
		Name: "cloudflare_certificate_pack",
		F:    testSweepCloudflareCertificatePack,
	})
}

func testSweepCloudflareCertificatePack(r string) error {
	client, clientErr := sharedClient()
	if clientErr != nil {
		log.Printf("[ERROR] Failed to create Cloudflare client: %s", clientErr)
	}

	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	certificates, certErr := client.ListCertificatePacks(context.Background(), zoneID)
	if certErr != nil {
		log.Printf("[ERROR] Failed to fetch certificate packs: %s", clientErr)
	}

	if len(certificates) == 0 {
		log.Print("[DEBUG] No Cloudflare certificate packs to sweep")
		return nil
	}

	for _, certificate := range certificates {
		if err := client.DeleteCertificatePack(context.Background(), zoneID, certificate.ID); err != nil {
			log.Printf("[ERROR] Failed to delete certificate pack %s", certificate.ID)
		}
	}

	return nil
}

func TestAccCertificatePackAdvancedDigicert(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_certificate_pack." + rnd
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificatePackAdvancedDigicertConfig(zoneID, domain, "advanced", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "type", "advanced"),
					resource.TestCheckResourceAttr(name, "hosts.#", "2"),
					resource.TestCheckResourceAttr(name, "validation_method", "http"),
					resource.TestCheckResourceAttr(name, "validity_days", "365"),
					resource.TestCheckResourceAttr(name, "certificate_authority", "digicert"),
					resource.TestCheckResourceAttr(name, "cloudflare_branding", "false"),
				),
			},
		},
	})
}

func testAccCertificatePackAdvancedDigicertConfig(zoneID, domain, certType, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_certificate_pack" "%[3]s" {
  zone_id = "%[1]s"
  type = "%[4]s"
  hosts = [
    "%[3]s.%[2]s",
    "%[2]s"
  ]
  validation_method = "http"
  validity_days = 365
  certificate_authority = "digicert"
  cloudflare_branding = false
}`, zoneID, domain, rnd, certType)
}

func TestAccCertificatePackAdvancedLetsEncrypt(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_certificate_pack." + rnd
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificatePackAdvancedLetsEncryptConfig(zoneID, domain, "advanced", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "type", "advanced"),
					resource.TestCheckResourceAttr(name, "hosts.#", "2"),
					resource.TestCheckResourceAttr(name, "validation_method", "txt"),
					resource.TestCheckResourceAttr(name, "validity_days", "90"),
					resource.TestCheckResourceAttr(name, "certificate_authority", "lets_encrypt"),
					resource.TestCheckResourceAttr(name, "cloudflare_branding", "false"),
				),
			},
		},
	})
}

func testAccCertificatePackAdvancedLetsEncryptConfig(zoneID, domain, certType, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_certificate_pack" "%[3]s" {
  zone_id = "%[1]s"
  type = "%[4]s"
  hosts = [
    "*.%[2]s",
    "%[2]s"
  ]
  validation_method = "txt"
  validity_days = 90
  certificate_authority = "lets_encrypt"
  cloudflare_branding = false
}`, zoneID, domain, rnd, certType)
}

func TestAccCertificatePackDedicatedCustom(t *testing.T) {
	t.Skip("Pending investigation into ACM entitlements")

	rnd := generateRandomResourceName()
	name := "cloudflare_certificate_pack." + rnd
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domain := os.Getenv("CLOUDFLARE_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificatePackDedicatedCustomConfig(zoneID, domain, "dedicated_custom", rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "type", "dedicated_custom"),
					resource.TestCheckResourceAttr(name, "hosts.#", "2"),
				),
			},
		},
	})
}

func testAccCertificatePackDedicatedCustomConfig(zoneID, domain, certType, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_certificate_pack" "%[3]s" {
  zone_id = "%[1]s"
  type = "%[4]s"
  hosts = [
    "%[3]s.%[2]s",
    "%[2]s"
  ]
}`, zoneID, domain, rnd, certType)
}
