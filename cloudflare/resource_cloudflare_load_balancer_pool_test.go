package cloudflare

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"regexp"
	"testing"

	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareLoadBalancerPool_Basic(t *testing.T) {
	// multiple instances of this config would conflict but we only use it once
	t.Parallel()
	testStartTime := time.Now().UTC()
	var loadBalancerPool cloudflare.LoadBalancerPool
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer_pool." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerPoolConfigBasic(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerPoolExists(name, &loadBalancerPool),
					// dont check that specified values are set, this will be evident by lack of plan diff
					// some values will get empty values
					resource.TestCheckResourceAttr(name, "check_regions.#", "0"),
					resource.TestCheckResourceAttr(name, "header.#", "0"),
					// also expect api to generate some values
					testAccCheckCloudflareLoadBalancerPoolDates(name, &loadBalancerPool, testStartTime),
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancerPool_FullySpecified(t *testing.T) {
	t.Parallel()
	var loadBalancerPool cloudflare.LoadBalancerPool
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer_pool." + rnd
	headerValue := os.Getenv("CLOUDFLARE_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerPoolConfigFullySpecified(rnd, headerValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerPoolExists(name, &loadBalancerPool),
					// checking our overrides of default values worked
					resource.TestCheckResourceAttr(name, "enabled", "false"),
					resource.TestCheckResourceAttr(name, "description", "tfacc-fully-specified"),
					resource.TestCheckResourceAttr(name, "check_regions.#", "1"),
					resource.TestCheckResourceAttr(name, "minimum_origins", "2"),

					func(state *terraform.State) error {
						for _, rs := range state.RootModule().Resources {
							for k, v := range rs.Primary.Attributes {
								r, _ := regexp.Compile("origins.*\\.header.*\\.header")

								if r.MatchString(k) {
									if v == "Host" {
										return nil
									}
								}
							}
						}
						return errors.New("Not equal")
					},
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancerPool_CreateAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var loadBalancerPool cloudflare.LoadBalancerPool
	var initialId string
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer_pool." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerPoolConfigBasic(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerPoolExists(name, &loadBalancerPool),
					testAccManuallyDeleteLoadBalancerPool(name, &loadBalancerPool, &initialId),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCheckCloudflareLoadBalancerPoolConfigBasic(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerPoolExists(name, &loadBalancerPool),
					func(state *terraform.State) error {
						if initialId == loadBalancerPool.ID {
							return fmt.Errorf("load balancer pool id is unchanged even after we thought we deleted it ( %s )",
								loadBalancerPool.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckCloudflareLoadBalancerPoolDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_load_balancer_pool" {
			continue
		}

		_, err := client.LoadBalancerPoolDetails(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Load balancer pool still exists")
		}
	}

	return nil
}

func testAccCheckCloudflareLoadBalancerPoolExists(n string, loadBalancerPool *cloudflare.LoadBalancerPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Load Balancer ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		foundLoadBalancerPool, err := client.LoadBalancerPoolDetails(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*loadBalancerPool = foundLoadBalancerPool

		return nil
	}
}

func testAccCheckCloudflareLoadBalancerPoolDates(n string, loadBalancerPool *cloudflare.LoadBalancerPool, testStartTime time.Time) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, _ := s.RootModule().Resources[n]

		for timeStampAttr, serverVal := range map[string]time.Time{"created_on": *loadBalancerPool.CreatedOn, "modified_on": *loadBalancerPool.ModifiedOn} {
			timeStamp, err := time.Parse(time.RFC3339Nano, rs.Primary.Attributes[timeStampAttr])
			if err != nil {
				return err
			}

			if timeStamp != serverVal {
				return fmt.Errorf("state value of %s: %s is different than server created value: %s",
					timeStampAttr, rs.Primary.Attributes[timeStampAttr], serverVal.Format(time.RFC3339Nano))
			}

			// check retrieved values are reasonable
			// note this could fail if local time is out of sync with server time
			if timeStamp.Before(testStartTime) {
				return fmt.Errorf("State value of %s: %s should be greater than test start time: %s",
					timeStampAttr, timeStamp.Format(time.RFC3339Nano), testStartTime.Format(time.RFC3339Nano))
			}
		}

		return nil
	}
}

func testAccManuallyDeleteLoadBalancerPool(name string, loadBalancerPool *cloudflare.LoadBalancerPool, initialId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*cloudflare.API)
		*initialId = loadBalancerPool.ID
		err := client.DeleteLoadBalancerPool(context.Background(), loadBalancerPool.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

// using IPs from 192.0.2.0/24 as per RFC5737
func testAccCheckCloudflareLoadBalancerPoolConfigBasic(id string) string {
	return fmt.Sprintf(`
resource "cloudflare_load_balancer_pool" "%[1]s" {
  name = "my-tf-pool-basic-%[1]s"
  origins {
    name = "example-1"
    address = "192.0.2.1"
    enabled = true
  }
}`, id)
}

func testAccCheckCloudflareLoadBalancerPoolConfigFullySpecified(id string, headerValue string) string {
	return fmt.Sprintf(`
resource "cloudflare_load_balancer_pool" "%[1]s" {
  name = "my-tf-pool-basic-%[1]s"
  origins {
    name = "example-1"
    address = "192.0.2.1"
    enabled = false
    weight = 1.0
    header {
    	header = "Host"
    	values = ["test1.%[2]s"]
 	}
  }
  origins {
    name = "example-2"
    address = "192.0.2.2"
    weight = 0.5
    header {
    	header = "Host"
    	values = ["test2.%[2]s"]
 	}
  }
  check_regions = ["WEU"]
  description = "tfacc-fully-specified"
  enabled = false
  minimum_origins = 2
  // monitor = abcd TODO: monitor resource
  notification_email = "someone@example.com"
}`, id, headerValue)
	// TODO add field to config after creating monitor resource
}
