package cloudflare

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareLoadBalancerMonitor_Basic(t *testing.T) {
	// multiple instances of this config would conflict but we only use it once
	t.Parallel()
	testStartTime := time.Now().UTC()
	var loadBalancerMonitor cloudflare.LoadBalancerMonitor
	name := "cloudflare_load_balancer_monitor.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerMonitorConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerMonitorExists(name, &loadBalancerMonitor),
					// dont check that specified values are set, this will be evident by lack of plan diff
					// some values will get empty values
					resource.TestCheckResourceAttr(name, "description", ""),
					resource.TestCheckResourceAttr(name, "header.#", "0"),
					// also expect api to generate some values
					testAccCheckCloudflareLoadBalancerMonitorDates(name, &loadBalancerMonitor, testStartTime),
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancerMonitor_FullySpecified(t *testing.T) {
	t.Parallel()
	var loadBalancerMonitor cloudflare.LoadBalancerMonitor
	name := "cloudflare_load_balancer_monitor.test"
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerMonitorConfigFullySpecified(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerMonitorExists(name, &loadBalancerMonitor),
					// checking our overrides of default values worked
					resource.TestCheckResourceAttr(name, "path", "/custom"),
					resource.TestCheckResourceAttr(name, "header.#", "1"),
					resource.TestCheckResourceAttr(name, "retries", "5"),
					resource.TestCheckResourceAttr(name, "port", "8080"),
					resource.TestCheckResourceAttr(name, "expected_body", "dead"),
					resource.TestCheckResourceAttr(name, "probe_zone", zoneName),
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancerMonitor_EmptyExpectedBody(t *testing.T) {
	t.Parallel()
	var loadBalancerMonitor cloudflare.LoadBalancerMonitor
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_load_balancer_monitor.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerMonitorConfigEmptyExpectedBody(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerMonitorExists(name, &loadBalancerMonitor),
					// checking empty string value passes all validations and created
					resource.TestCheckResourceAttr(name, "expected_body", ""),
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancerMonitor_TcpFullySpecified(t *testing.T) {
	t.Parallel()
	var loadBalancerMonitor cloudflare.LoadBalancerMonitor
	name := "cloudflare_load_balancer_monitor.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerMonitorConfigTcpFullySpecified(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerMonitorExists(name, &loadBalancerMonitor),
					// checking our overrides of default values worked
					resource.TestCheckResourceAttr(name, "retries", "5"),
					resource.TestCheckResourceAttr(name, "port", "8080"),
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancerMonitor_NoRequired(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudflareLoadBalancerMonitorConfigMissingRequired(),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected_codes must be set")),
			},
		},
	})
}

func TestAccCloudflareLoadBalancerMonitor_Update(t *testing.T) {
	t.Parallel()
	var loadBalancerMonitor cloudflare.LoadBalancerMonitor
	var initialId string
	name := "cloudflare_load_balancer_monitor.test"
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerMonitorConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerMonitorExists(name, &loadBalancerMonitor),
				),
			},
			{
				PreConfig: func() {
					initialId = loadBalancerMonitor.ID
				},
				Config: testAccCheckCloudflareLoadBalancerMonitorConfigFullySpecified(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerMonitorExists(name, &loadBalancerMonitor),
					func(state *terraform.State) error {
						if initialId != loadBalancerMonitor.ID {
							return fmt.Errorf("wanted update but monitor got recreated (id changed %q -> %q)",
								initialId, loadBalancerMonitor.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccCloudflareLoadBalancerMonitor_CreateAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var loadBalancerMonitor cloudflare.LoadBalancerMonitor
	var initialId string
	name := "cloudflare_load_balancer_monitor.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLoadBalancerMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLoadBalancerMonitorConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerMonitorExists(name, &loadBalancerMonitor),
					testAccManuallyDeleteLoadBalancerMonitor(name, &loadBalancerMonitor, &initialId),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCheckCloudflareLoadBalancerMonitorConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareLoadBalancerMonitorExists(name, &loadBalancerMonitor),
					func(state *terraform.State) error {
						if initialId == loadBalancerMonitor.ID {
							return fmt.Errorf("load balancer monitor id is unchanged even after we thought we deleted it ( %s )",
								loadBalancerMonitor.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckCloudflareLoadBalancerMonitorDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_load_balancer_monitor" {
			continue
		}

		_, err := client.LoadBalancerMonitorDetails(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Load balancer monitor still exists")
		}
	}

	return nil
}

func testAccCheckCloudflareLoadBalancerMonitorExists(n string, load *cloudflare.LoadBalancerMonitor) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Load Balancer Monitor ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		foundLoadBalancerMonitor, err := client.LoadBalancerMonitorDetails(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*load = foundLoadBalancerMonitor

		return nil
	}
}

func testAccCheckCloudflareLoadBalancerMonitorDates(n string, loadBalancerMonitor *cloudflare.LoadBalancerMonitor, testStartTime time.Time) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, _ := s.RootModule().Resources[n]

		for timeStampAttr, serverVal := range map[string]time.Time{"created_on": *loadBalancerMonitor.CreatedOn, "modified_on": *loadBalancerMonitor.ModifiedOn} {
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

func testAccManuallyDeleteLoadBalancerMonitor(name string, loadBalancerMonitor *cloudflare.LoadBalancerMonitor, initialId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*cloudflare.API)
		*initialId = loadBalancerMonitor.ID
		err := client.DeleteLoadBalancerMonitor(context.Background(), loadBalancerMonitor.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckCloudflareLoadBalancerMonitorConfigBasic() string {
	return `
resource "cloudflare_load_balancer_monitor" "test" {
  expected_body = "alive"
  expected_codes = "2xx"

}`
}

func testAccCheckCloudflareLoadBalancerMonitorConfigFullySpecified(zoneName string) string {
	return fmt.Sprintf(`
resource "cloudflare_load_balancer_monitor" "test" {
  expected_body = "dead"
  expected_codes = "5xx"
  method = "HEAD"
  timeout = 9
  path = "/custom"
  interval = 60
  retries = 5
  port = 8080
  description = "this is a very weird load balancer"
  probe_zone = "%[1]s"
  header {
    header = "Host"
    values = ["%[1]s"]
  }
}`, zoneName)
}

func testAccCheckCloudflareLoadBalancerMonitorConfigEmptyExpectedBody(resourceName string) string {
	return fmt.Sprintf(`
resource "cloudflare_load_balancer_monitor" "%[1]s" {
  expected_body = ""
  expected_codes = "2xx"
  description = "we don't want to check for a given body"
}`, resourceName)
}

func testAccCheckCloudflareLoadBalancerMonitorConfigTcpFullySpecified() string {
	return `
resource "cloudflare_load_balancer_monitor" "test" {
  type = "tcp"
  method = "connection_established"
  timeout = 9
  interval = 60
  retries = 5
  port = 8080
  description = "this is a very weird tcp load balancer"
}`
}

func testAccCheckCloudflareLoadBalancerMonitorConfigMissingRequired() string {
	return `
resource "cloudflare_load_balancer_monitor" "test" {
  description = "this is a wrong config"
}`
}
