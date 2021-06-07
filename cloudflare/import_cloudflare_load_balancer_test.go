package cloudflare

import (
	"fmt"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudflareLoadBalancer_Import(t *testing.T) {
	t.Parallel()
	var loadBalancer cloudflare.LoadBalancer
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerConfigBasic(zoneID, zone, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
				),
			},
			{
				ResourceName:        name,
				ImportStateIdPrefix: fmt.Sprintf("%s/", zoneID),
				ImportState:         true,
				ImportStateVerify:   true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
				),
			},
		},
	})
}
