package cloudflare

import (
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudflareLoadBalancerPool_Import(t *testing.T) {
	t.Parallel()
	var loadBalancerPool cloudflare.LoadBalancerPool
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer_pool." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerPoolConfigBasic(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerPoolExists(name, &loadBalancerPool),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerPoolExists(name, &loadBalancerPool),
				),
			},
		},
	})
}
