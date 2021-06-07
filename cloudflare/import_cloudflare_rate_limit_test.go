package cloudflare

import (
	"os"
	"testing"

	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudflareRateLimit_Import(t *testing.T) {
	t.Parallel()
	var rateLimit cloudflare.RateLimit
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_rate_limit." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRateLimitConfigMatchingUrl(zoneID, rnd, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRateLimitExists(name, &rateLimit),
					testAccCheckCloudflareRateLimitIDIsValid(name, zoneID),
				),
			},
			{
				ResourceName:        name,
				ImportStateIdPrefix: fmt.Sprintf("%s/", zoneID),
				ImportState:         true,
				ImportStateVerify:   true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRateLimitExists(name, &rateLimit),
					testAccCheckCloudflareRateLimitIDIsValid(name, zoneID),
				),
			},
		},
	})
}
