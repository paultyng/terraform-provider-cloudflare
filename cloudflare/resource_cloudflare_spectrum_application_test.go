package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("cloudflare_spectrum_applications", &resource.Sweeper{
		Name: "cloudflare_spectrum_applications",
		F:    testSweepCloudflareSpectrumApplications,
	})
}

func testSweepCloudflareSpectrumApplications(r string) error {
	client, clientErr := sharedClient()
	if clientErr != nil {
		log.Printf("[ERROR] Failed to create Cloudflare client: %s", clientErr)
	}

	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zones, zoneErr := client.ListZones(context.Background(), zone)

	if zoneErr != nil {
		log.Printf("[ERROR] Failed to fetch Cloudflare zones: %s", zoneErr)
	}

	for _, zone := range zones {
		spectrumApps, spectrumErr := client.SpectrumApplications(context.Background(), zone.ID)
		if spectrumErr != nil {
			log.Printf("[ERROR] Failed to fetch Cloudflare spectrum applications: %s", zoneErr)
		}

		if len(spectrumApps) == 0 {
			log.Print("[DEBUG] No Cloudflare spectrum applications to sweep")
			return nil
		}

		for _, application := range spectrumApps {
			log.Printf("[INFO] Deleting Cloudflare spectrum application ID: %s", application.ID)
			err := client.DeleteSpectrumApplication(context.Background(), zone.ID, application.ID)

			if err != nil {
				log.Printf("[ERROR] Failed to delete Cloudflare spectrum application (%s) in zone ID: %s", application.ID, zone.ID)
			}
		}
	}

	return nil
}

func TestAccCloudflareSpectrumApplication_Basic(t *testing.T) {
	var spectrumApp cloudflare.SpectrumApplication
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_spectrum_application." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareSpectrumApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareSpectrumApplicationConfigBasic(zoneID, domain, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareSpectrumApplicationExists(name, &spectrumApp),
					testAccCheckCloudflareSpectrumApplicationIDIsValid(name),
					resource.TestCheckResourceAttr(name, "protocol", "tcp/22"),
					resource.TestCheckResourceAttr(name, "origin_direct.#", "1"),
					resource.TestCheckResourceAttr(name, "origin_direct.0", "tcp://1.2.3.4:23"),
					resource.TestCheckResourceAttr(name, "origin_port", "22"),
				),
			},
		},
	})
}

func TestAccCloudflareSpectrumApplication_OriginDNS(t *testing.T) {
	var spectrumApp cloudflare.SpectrumApplication
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_spectrum_application." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareSpectrumApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareSpectrumApplicationConfigOriginDNS(zoneID, domain, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareSpectrumApplicationExists(name, &spectrumApp),
					testAccCheckCloudflareSpectrumApplicationIDIsValid(name),
					resource.TestCheckResourceAttr(name, "protocol", "tcp/22"),
					resource.TestCheckResourceAttr(name, "origin_dns.#", "1"),
					resource.TestCheckResourceAttr(name, "origin_dns.0.name", fmt.Sprintf("%s.origin.%s", rnd, domain)),
					resource.TestCheckResourceAttr(name, "origin_port", "22"),
				),
			},
		},
	})
}

func TestAccCloudflareSpectrumApplication_OriginPortRange(t *testing.T) {
	var spectrumApp cloudflare.SpectrumApplication
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_spectrum_application." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareSpectrumApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareSpectrumApplicationConfigOriginPortRange(zoneID, domain, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareSpectrumApplicationExists(name, &spectrumApp),
					testAccCheckCloudflareSpectrumApplicationIDIsValid(name),
					resource.TestCheckResourceAttr(name, "protocol", "tcp/22-23"),
					resource.TestCheckResourceAttr(name, "origin_dns.#", "1"),
					resource.TestCheckResourceAttr(name, "origin_dns.0.name", fmt.Sprintf("%s.origin.%s", rnd, domain)),
					resource.TestCheckResourceAttr(name, "origin_port_range.#", "1"),
					resource.TestCheckResourceAttr(name, "origin_port_range.0.start", "2022"),
					resource.TestCheckResourceAttr(name, "origin_port_range.0.end", "2023"),
				),
			},
		},
	})
}

func TestAccCloudflareSpectrumApplication_Update(t *testing.T) {
	var spectrumApp cloudflare.SpectrumApplication
	var initialID string
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_spectrum_application." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareSpectrumApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareSpectrumApplicationConfigBasic(zoneID, domain, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareSpectrumApplicationExists(name, &spectrumApp),
					testAccCheckCloudflareSpectrumApplicationIDIsValid(name),
					resource.TestCheckResourceAttr(name, "origin_direct.0", "tcp://1.2.3.4:23"),
				),
			},
			{
				PreConfig: func() {
					initialID = spectrumApp.ID
				},
				Config: testAccCheckCloudflareSpectrumApplicationConfigBasicUpdated(zoneID, domain, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareSpectrumApplicationExists(name, &spectrumApp),
					testAccCheckCloudflareSpectrumApplicationIDIsValid(name),
					func(state *terraform.State) error {
						if initialID != spectrumApp.ID {
							// want in place update
							return fmt.Errorf("spectrum application id is different after second config applied ( %s -> %s )",
								initialID, spectrumApp.ID)
						}
						return nil
					},
					resource.TestCheckResourceAttr(name, "origin_direct.0", "tcp://1.2.3.4:23"),
				),
			},
		},
	})
}

func TestAccCloudflareSpectrumApplication_CreateAfterManualDestroy(t *testing.T) {
	var spectrumApp cloudflare.SpectrumApplication
	var initialID string
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_spectrum_application." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareSpectrumApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareSpectrumApplicationConfigBasic(zoneID, domain, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareSpectrumApplicationExists(name, &spectrumApp),
					testAccCheckCloudflareSpectrumApplicationIDIsValid(name),
					testAccManuallyDeleteSpectrumApplication(name, &spectrumApp, &initialID),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCheckCloudflareSpectrumApplicationConfigBasic(zoneID, domain, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareSpectrumApplicationExists(name, &spectrumApp),
					testAccCheckCloudflareSpectrumApplicationIDIsValid(name),
					func(state *terraform.State) error {
						if initialID == spectrumApp.ID {
							return fmt.Errorf("spectrum application id is unchanged even after we thought we deleted it ( %s )",
								spectrumApp.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckCloudflareSpectrumApplicationDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_spectrum_application" {
			continue
		}

		_, err := client.SpectrumApplication(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("spectrum application still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCloudflareSpectrumApplicationExists(n string, spectrumApp *cloudflare.SpectrumApplication) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Load Balancer ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		foundSpectrumApplication, err := client.SpectrumApplication(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		*spectrumApp = foundSpectrumApplication

		return nil
	}
}

func testAccCheckCloudflareSpectrumApplicationIDIsValid(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if len(rs.Primary.ID) != 32 {
			return fmt.Errorf("invalid id %q, should be a string of length 32", rs.Primary.ID)
		}

		if zoneID, ok := rs.Primary.Attributes["zone_id"]; !ok || len(zoneID) < 1 {
			return errors.New("zone_id is unset, should always be set with id")
		}
		return nil
	}
}

func testAccManuallyDeleteSpectrumApplication(name string, spectrumApp *cloudflare.SpectrumApplication, initialID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, _ := s.RootModule().Resources[name]
		client := testAccProvider.Meta().(*cloudflare.API)
		*initialID = spectrumApp.ID
		err := client.DeleteSpectrumApplication(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckCloudflareSpectrumApplicationConfigBasic(zoneID, zoneName, ID string) string {
	return fmt.Sprintf(`
resource "cloudflare_spectrum_application" "%[3]s" {
  zone_id  = "%[1]s"
  protocol = "tcp/22"

  dns {
    type = "CNAME"
    name = "%[3]s.%[2]s"
  }

  origin_direct = ["tcp://1.2.3.4:23"]
  origin_port   = 22
}
`, zoneID, zoneName, ID)
}

func testAccCheckCloudflareSpectrumApplicationConfigOriginDNS(zoneID, zoneName, ID string) string {
	return fmt.Sprintf(`
resource "cloudflare_spectrum_application" "%[3]s" {
  zone_id  = "%[1]s"
  protocol = "tcp/22"

  dns {
    type = "CNAME"
    name = "%[3]s.%[2]s"
  }

  origin_dns {
    name = "%[3]s.origin.%[2]s"
  }
  origin_port   = 22
}`, zoneID, zoneName, ID)
}

func testAccCheckCloudflareSpectrumApplicationConfigOriginPortRange(zoneID, zoneName, ID string) string {
	return fmt.Sprintf(`
resource "cloudflare_spectrum_application" "%[3]s" {
  zone_id  = "%[1]s"
  protocol = "tcp/22-23"

  dns {
    type = "CNAME"
    name = "%[3]s.%[2]s"
  }

  origin_dns {
    name = "%[3]s.origin.%[2]s"
  }
  origin_port_range {
    start = 2022
    end   = 2023
  }
}`, zoneID, zoneName, ID)
}

func testAccCheckCloudflareSpectrumApplicationConfigBasicUpdated(zoneID, zoneName, ID string) string {
	return fmt.Sprintf(`
resource "cloudflare_spectrum_application" "%[3]s" {
  zone_id  = "%[1]s"
  protocol = "tcp/22"

  dns {
		type = "CNAME"
		name = "%[3]s.%[2]s"
  }

  origin_direct = ["tcp://1.2.3.4:23"]
  origin_port   = 22
}`, zoneID, zoneName, ID)
}
