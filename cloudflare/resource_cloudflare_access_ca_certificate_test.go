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

func TestAccCloudflareAccessCACertificate_AccountLevel(t *testing.T) {
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_ca_certificate.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessCACertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessCACertificateBasic(rnd, domain, AccessIdentifier{Type: AccountType, Value: accountID}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttrSet(name, "application_id"),
					resource.TestCheckResourceAttrSet(name, "aud"),
					resource.TestCheckResourceAttrSet(name, "public_key"),
				),
			},
		},
	})
}

func TestAccCloudflareAccessCACertificate_ZoneLevel(t *testing.T) {
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_ca_certificate.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessCACertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessCACertificateBasic(rnd, domain, AccessIdentifier{Type: ZoneType, Value: zoneID}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttrSet(name, "application_id"),
					resource.TestCheckResourceAttrSet(name, "aud"),
					resource.TestCheckResourceAttrSet(name, "public_key"),
				),
			},
		},
	})
}

func testAccCloudflareAccessCACertificateBasic(resourceName, domain string, identifier AccessIdentifier) string {
	return fmt.Sprintf(`
resource "cloudflare_access_application" "%[1]s" {
	name     = "%[1]s"
	%[3]s_id = "%[4]s"
	domain   = "%[1]s.%[2]s"
}

resource "cloudflare_access_ca_certificate" "%[1]s" {
  %[3]s_id       = "%[4]s"
  application_id = cloudflare_access_application.%[1]s.id
}`, resourceName, domain, identifier.Type, identifier.Value)
}

func testAccCheckCloudflareAccessCACertificateDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_access_ca_certificate" {
			continue
		}

		_, err := client.AccessCACertificate(context.Background(), rs.Primary.Attributes["account_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Access CA certificate still exists")
		}

		_, err = client.ZoneLevelAccessCACertificate(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Access CA certificate still exists")
		}
	}

	return nil
}
