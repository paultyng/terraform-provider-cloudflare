package cloudflare

import (
	"context"
	"fmt"
	"testing"

	"time"

	"os"

	"regexp"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareLoadBalancer_Basic(t *testing.T) {
	// multiple instances of this config would conflict but we only use it once
	t.Parallel()
	testStartTime := time.Now().UTC()
	var loadBalancer cloudflare.LoadBalancer
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerConfigBasic(zoneID, zone, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
					// dont check that specified values are set, this will be evident by lack of plan diff
					// some values will get empty values
					resource.TestCheckResourceAttr(name, "pop_pools.#", "0"),
					resource.TestCheckResourceAttr(name, "region_pools.#", "0"),
					// also expect api to generate some values
					testAccCheckCloudflareLoadBalancerDates(name, &loadBalancer, testStartTime),
					resource.TestCheckResourceAttr(name, "proxied", "false"), // default value
					resource.TestCheckResourceAttr(name, "ttl", "30"),
					resource.TestCheckResourceAttr(name, "steering_policy", "off"),
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancer_SessionAffinity(t *testing.T) {
	t.Parallel()
	var loadBalancer cloudflare.LoadBalancer
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerConfigSessionAffinity(zoneID, zone, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
					// explicitly verify that our session_affinity has been set
					resource.TestCheckResourceAttr(name, "session_affinity", "cookie"),
					resource.TestCheckResourceAttr(name, "session_affinity_ttl", "1800"),
					resource.TestCheckResourceAttr(name, "session_affinity_attributes.samesite", "Auto"),
					resource.TestCheckResourceAttr(name, "session_affinity_attributes.secure", "Auto"),
					resource.TestCheckResourceAttr(name, "session_affinity_attributes.drain_duration", "60"),
					// dont check that other specified values are set, this will be evident by lack
					// of plan diff some values will get empty values
					resource.TestCheckResourceAttr(name, "pop_pools.#", "0"),
					resource.TestCheckResourceAttr(name, "region_pools.#", "0"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerConfigSessionAffinityIPCookie(zoneID, zone, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
					// explicitly verify that our session_affinity has been set
					resource.TestCheckResourceAttr(name, "session_affinity", "ip_cookie"),
					// session_affinity_ttl should not be present as it isn't set
					resource.TestCheckNoResourceAttr(name, "session_affinity_ttl"),
					resource.TestCheckNoResourceAttr(name, "session_affinity_attributes"),
					// dont check that other specified values are set, this will be evident by lack
					// of plan diff some values will get empty values
					resource.TestCheckResourceAttr(name, "pop_pools.#", "0"),
					resource.TestCheckResourceAttr(name, "region_pools.#", "0"),
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancer_GeoBalanced(t *testing.T) {
	t.Parallel()
	var loadBalancer cloudflare.LoadBalancer
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerConfigGeoBalanced(zoneID, zone, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
					// checking our overrides of default values worked
					resource.TestCheckResourceAttr(name, "description", "tf-acctest load balancer using geo-balancing"),
					resource.TestCheckResourceAttr(name, "proxied", "true"),
					resource.TestCheckResourceAttr(name, "ttl", "0"),
					resource.TestCheckResourceAttr(name, "steering_policy", "geo"),
					resource.TestCheckResourceAttr(name, "pop_pools.#", "1"),
					resource.TestCheckResourceAttr(name, "region_pools.#", "1"),
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancer_Rules(t *testing.T) {
	t.Parallel()
	var loadBalancer cloudflare.LoadBalancer
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerConfigRules(zoneID, zone, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
					// checking our overrides of default values worked
					resource.TestCheckResourceAttr(name, "description", "rules lb"),
					resource.TestCheckResourceAttr(name, "rules.0.name", "test rule 1"),
					resource.TestCheckResourceAttr(name, "rules.0.condition", "dns.qry.type == 28"),
					resource.TestCheckResourceAttr(name, "rules.0.overrides.#", "1"),
					resource.TestCheckResourceAttr(name, "rules.0.overrides.0.steering_policy", "geo"),
					resource.TestCheckResourceAttr(name, "rules.0.overrides.0.session_affinity_attributes.samesite", "Auto"),
					resource.TestCheckResourceAttr(name, "rules.0.overrides.0.session_affinity_attributes.secure", "Auto"),
					resource.TestCheckResourceAttr(name, "rules.#", "3"),
					resource.TestCheckResourceAttr(name, "rules.1.fixed_response.message_body", "hello"),
					resource.TestCheckResourceAttr(name, "rules.2.overrides.0.region_pools.#", "1"),
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancer_DuplicatePool(t *testing.T) {
	t.Parallel()
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudflareLoadBalancerConfigDuplicatePool(zoneID, zone, rnd),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("duplicate entry specified for pop pool in location \"LAX\". each location must only be specified once")),
			},
		},
	})
}

/**
Any change to a load balancer  results in a new resource
Although the API client contains a modify method, this always results in 405 status
*/
func TestAccCloudflareLoadBalancer_Update(t *testing.T) {
	t.Parallel()
	var loadBalancer cloudflare.LoadBalancer
	var initialId string
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerConfigBasic(zoneID, zone, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
				),
			},
			{
				PreConfig: func() {
					initialId = loadBalancer.ID
				},
				Config: testAccCheckCloudflareLoadBalancerConfigGeoBalanced(zoneID, zone, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
					func(state *terraform.State) error {
						if initialId != loadBalancer.ID {
							// want in place update
							return fmt.Errorf("load balancer id is different after second config applied ( %s -> %s )",
								initialId, loadBalancer.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancer_CreateAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var loadBalancer cloudflare.LoadBalancer
	var initialId string
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := "cloudflare_load_balancer." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerConfigBasic(zoneID, zone, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
					testAccManuallyDeleteLoadBalancer(name, &loadBalancer, &initialId),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCheckCloudflareLoadBalancerConfigBasic(zoneID, zone, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerExists(name, &loadBalancer),
					testAccCheckCloudflareLoadBalancerIDIsValid(name, zoneID),
					func(state *terraform.State) error {
						if initialId == loadBalancer.ID {
							return fmt.Errorf("load balancer id is unchanged even after we thought we deleted it ( %s )",
								loadBalancer.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckCloudflareLoadBalancerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_load_balancer" {
			continue
		}

		_, err := client.LoadBalancerDetails(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("load balancer still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCloudflareLoadBalancerExists(n string, loadBalancer *cloudflare.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Load Balancer ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		foundLoadBalancer, err := client.LoadBalancerDetails(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		*loadBalancer = foundLoadBalancer

		return nil
	}
}

func testAccCheckCloudflareLoadBalancerIDIsValid(n, expectedZoneID string) resource.TestCheckFunc {
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

		if rs.Primary.Attributes["zone_id"] != expectedZoneID {
			return fmt.Errorf("zoneID attribute %q doesn't match the expected value %q", rs.Primary.Attributes["zone_id"], expectedZoneID)
		}

		return nil
	}
}

func testAccCheckCloudflareLoadBalancerDates(n string, loadBalancer *cloudflare.LoadBalancer, testStartTime time.Time) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, _ := s.RootModule().Resources[n]

		for timeStampAttr, serverVal := range map[string]time.Time{"created_on": *loadBalancer.CreatedOn, "modified_on": *loadBalancer.ModifiedOn} {
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

func testAccManuallyDeleteLoadBalancer(name string, loadBalancer *cloudflare.LoadBalancer, initialId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, _ := s.RootModule().Resources[name]
		client := testAccProvider.Meta().(*cloudflare.API)
		*initialId = loadBalancer.ID
		err := client.DeleteLoadBalancer(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckCloudflareLoadBalancerConfigBasic(zoneID, zone, id string) string {
	return testAccCheckCloudflareLoadBalancerPoolConfigBasic(id) + fmt.Sprintf(`
resource "cloudflare_load_balancer" "%[3]s" {
  zone_id = "%[1]s"
  name = "tf-testacc-lb-%[3]s.%[2]s"
  steering_policy = ""
  fallback_pool_id = "${cloudflare_load_balancer_pool.%[3]s.id}"
  default_pool_ids = ["${cloudflare_load_balancer_pool.%[3]s.id}"]
}`, zoneID, zone, id)
}

func testAccCheckCloudflareLoadBalancerConfigSessionAffinity(zoneID, zone, id string) string {
	return testAccCheckCloudflareLoadBalancerPoolConfigBasic(id) + fmt.Sprintf(`
resource "cloudflare_load_balancer" "%[3]s" {
  zone_id = "%[1]s"
  name = "tf-testacc-lb-session-affinity-%[3]s.%[2]s"
  fallback_pool_id = "${cloudflare_load_balancer_pool.%[3]s.id}"
  default_pool_ids = ["${cloudflare_load_balancer_pool.%[3]s.id}"]
	session_affinity = "cookie"
	session_affinity_ttl = 1800
	session_affinity_attributes = {
    samesite = "Auto"
    secure = "Auto"
    drain_duration = 60
	}
}`, zoneID, zone, id)
}

func testAccCheckCloudflareLoadBalancerConfigSessionAffinityIPCookie(zoneID, zone, id string) string {
	return testAccCheckCloudflareLoadBalancerPoolConfigBasic(id) + fmt.Sprintf(`
resource "cloudflare_load_balancer" "%[3]s" {
  zone_id = "%[1]s"
  name = "tf-testacc-lb-session-affinity-%[3]s.%[2]s"
  fallback_pool_id = "${cloudflare_load_balancer_pool.%[3]s.id}"
  default_pool_ids = ["${cloudflare_load_balancer_pool.%[3]s.id}"]
	session_affinity = "ip_cookie"
}`, zoneID, zone, id)
}

func testAccCheckCloudflareLoadBalancerConfigGeoBalanced(zoneID, zone, id string) string {
	return testAccCheckCloudflareLoadBalancerPoolConfigBasic(id) + fmt.Sprintf(`
resource "cloudflare_load_balancer" "%[3]s" {
  zone_id = "%[1]s"
  name = "tf-testacc-lb-%[3]s.%[2]s"
  fallback_pool_id = "${cloudflare_load_balancer_pool.%[3]s.id}"
  default_pool_ids = ["${cloudflare_load_balancer_pool.%[3]s.id}"]
  description = "tf-acctest load balancer using geo-balancing"
  proxied = true // can't set ttl with proxied
  steering_policy = "geo"
  pop_pools {
    pop = "LAX"
    pool_ids = ["${cloudflare_load_balancer_pool.%[3]s.id}"]
  }
  region_pools {
    region = "WNAM"
    pool_ids = ["${cloudflare_load_balancer_pool.%[3]s.id}"]
  }
}`, zoneID, zone, id)
}

func testAccCheckCloudflareLoadBalancerConfigDuplicatePool(zoneID, zone, id string) string {
	return testAccCheckCloudflareLoadBalancerPoolConfigBasic(id) + fmt.Sprintf(`
resource "cloudflare_load_balancer" "%[3]s" {
  zone_id = "%[1]s"
  name = "tf-testacc-lb-%[3]s.%[2]s"
  fallback_pool_id = "${cloudflare_load_balancer_pool.%[3]s.id}"
  default_pool_ids = ["${cloudflare_load_balancer_pool.%[3]s.id}"]
  pop_pools {
    pop = "LAX"
    pool_ids = ["i_am_an_invalid_pool_id"]
  }
  pop_pools {
    pop = "LAX"
    pool_ids = ["${cloudflare_load_balancer_pool.%[3]s.id}"]
  }
}`, zoneID, zone, id)
}

func testAccCheckCloudflareLoadBalancerConfigRules(zoneID, zone, id string) string {
	return testAccCheckCloudflareLoadBalancerPoolConfigBasic(id) + fmt.Sprintf(`
resource "cloudflare_load_balancer" "%[3]s" {
  zone_id = "%[1]s"
  name = "tf-testacc-lb-%[3]s.%[2]s"
  steering_policy = ""
  description = "rules lb"
  fallback_pool_id = "${cloudflare_load_balancer_pool.%[3]s.id}"
  default_pool_ids = ["${cloudflare_load_balancer_pool.%[3]s.id}"]
  rules {
    name = "test rule 1"
    condition = "dns.qry.type == 28"
    overrides {
      steering_policy = "geo"
      session_affinity_attributes = {
        samesite = "Auto"
        secure = "Auto"
      }
    }
  }
  rules {
    name = "test rule 2"
    condition = "dns.qry.type == 28"
    fixed_response = {
      message_body = "hello"
      status_code = "200"
      content_type = "html"
      location = "www.example.com"
    }
  }
  rules {
    name = "test rule 3"
    condition = "dns.qry.type == 28"
    overrides {
      region_pools {
		    region = "ENAM"
		    pool_ids = ["${cloudflare_load_balancer_pool.%[3]s.id}"]
	    }
    }
  }
}`, zoneID, zone, id)
}
