package cloudflare

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("cloudflare_zones", &resource.Sweeper{
		Name: "cloudflare_zones",
		F:    testSweepCloudflareZones,
	})
}

func testSweepCloudflareZones(r string) error {
	client, clientErr := sharedClient()
	if clientErr != nil {
		log.Printf("[ERROR] Failed to create Cloudflare client: %s", clientErr)
	}

	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	if accountID == "" {
		return errors.New("CLOUDFLARE_ACCOUNT_ID must be set")
	}
	zoneFilter := cloudflare.WithZoneFilters("", accountID, "")
	zones, zoneErr := client.ListZonesContext(context.TODO(), zoneFilter)
	if zoneErr != nil {
		log.Printf("[ERROR] Failed to fetch Cloudflare zones: %s", zoneErr)
	}

	if len(zones.Result) == 0 {
		log.Print("[DEBUG] No Cloudflare Zones to sweep")
		return nil
	}

	for _, zone := range zones.Result {
		// Don't try and sweep the static domains.
		if zone.Name == "terraform.cfapi.net" || zone.Name == "terraform2.cfapi.net" {
			continue
		}

		log.Printf("[INFO] Deleting Cloudflare Zone ID: %s", zone.ID)
		_, err := client.DeleteZone(context.Background(), zone.ID)

		if err != nil {
			log.Printf("[ERROR] Failed to delete Cloudflare Zone (%s): %s", zone.Name, err)
		}
	}

	return nil
}

func TestAccCloudflareZonesMatchName(t *testing.T) {

	t.Parallel()
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_zones.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareZonesConfigMatchName(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareZonesDataSourceID(name),
					resource.TestCheckResourceAttr(name, "filter.0.name", "baa-com.cfapi.net"),
					resource.TestCheckResourceAttr(name, "filter.0.paused", "false"),
					resource.TestCheckResourceAttr(name, "zones.#", "1"),
				),
			},
		},
	})
}

func TestAccCloudflareZonesMatchPaused(t *testing.T) {
	t.Parallel()
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_zones.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareZonesConfigMatchPaused(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "filter.0.name", "baa-org.cfapi.net"),
					resource.TestCheckResourceAttr(name, "filter.0.paused", "true"),
					resource.TestCheckResourceAttr(name, "zones.#", "1"),
				),
			},
		},
	})
}

func TestAccCloudflareZonesMatchRegexFilter(t *testing.T) {
	t.Parallel()
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_zones.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareZonesConfigMatchRegex(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareZonesDataSourceID(name),
					resource.TestCheckResourceAttr(name, "filter.0.match", "baa-*"),
					resource.TestCheckResourceAttr(name, "zones.#", "1"),
				),
			},
		},
	})
}

func TestAccCloudflareZonesMatchFuzzyLookup(t *testing.T) {
	t.Parallel()
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("data.cloudflare_zones.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareZonesMatchFuzzyLookup(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareZonesDataSourceID(name),
					resource.TestCheckResourceAttr(name, "filter.0.name", "foo-net"),
					resource.TestCheckResourceAttr(name, "filter.0.lookup_type", "contains"),
					resource.TestCheckResourceAttr(name, "zones.#", "1"),
				),
			},
		},
	})
}

func testAccCheckCloudflareZonesDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		all := s.RootModule().Resources
		rs, ok := all[n]
		if !ok {
			return fmt.Errorf("Can't find zones data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot zones source ID not set")
		}
		return nil
	}
}

func testAccCheckCloudflareZonesReturned(n string, a string, check func(int) bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		all := s.RootModule().Resources
		rs, ok := all[n]
		if !ok {
			return fmt.Errorf("Can't find zones data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot zones source ID not set")
		}

		count, _ := strconv.Atoi(rs.Primary.Attributes[a])
		if !check(count) {
			return fmt.Errorf("Error evaluating %s.%s actual count: %d", n, a, count)
		}
		return nil
	}
}

func testAccCloudflareZonesConfigMatchName(rnd string) string {
	return fmt.Sprintf(`
data "cloudflare_zones" "%[2]s" {
  filter {
    name   = "baa-com.cfapi.net"
    // This is an ordering fix to ensure that the test suite doesn't assert
    // state before all the resources are available.
    paused = "${cloudflare_zone.foo_net.paused}"
  }
}

%[1]s
`, testZones, rnd)
}

func testAccCloudflareZonesConfigMatchPaused(rnd string) string {
	return fmt.Sprintf(`
data "cloudflare_zones" "%[2]s" {
  filter {
    name   = "baa-org.cfapi.net"
    paused = "${cloudflare_zone.baa_org.paused}"
  }
}

%[1]s
`, testZones, rnd)
}

func testAccCloudflareZonesConfigMatchRegex(rnd string) string {
	return fmt.Sprintf(`
data "cloudflare_zones" "%[2]s" {
  filter {
    match  = "baa-*"
    // This is an ordering fix to ensure that the test suite doesn't assert
    // state before all the resources are available.
    paused = "${cloudflare_zone.foo_net.paused}"
  }
}

%[1]s
`, testZones, rnd)
}

func testAccCloudflareZonesMatchFuzzyLookup(rnd string) string {
	return fmt.Sprintf(`
data "cloudflare_zones" "%[2]s" {
  filter {
    name = "foo-net"
    lookup_type = "contains"
    // This is an ordering fix to ensure that the test suite doesn't assert
    // state before all the resources are available.
    paused = "${cloudflare_zone.foo_net.paused}"
  }
}

%[1]s
`, testZones, rnd)
}

const testZones = `resource "cloudflare_zone" "baa_com" {
  zone       = "baa-com.cfapi.net"
  paused     = false
  jump_start = false
}

resource "cloudflare_zone" "baa_org" {
  zone       = "baa-org.cfapi.net"
  paused     = true
  jump_start = false
}

resource "cloudflare_zone" "baa_net" {
  zone       = "baa-net.cfapi.net"
  paused     = true
  jump_start = false
}

resource "cloudflare_zone" "foo_net" {
  zone       = "foo-net.cfapi.net"
  paused     = false
  jump_start = false
  depends_on = ["cloudflare_zone.baa_net", "cloudflare_zone.baa_org", "cloudflare_zone.baa_com"]
}`
