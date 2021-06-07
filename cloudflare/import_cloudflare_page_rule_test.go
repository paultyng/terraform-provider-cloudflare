package cloudflare

import (
	"testing"

	"fmt"
	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudflarePageRule_Import(t *testing.T) {
	t.Parallel()
	var pageRule cloudflare.PageRule
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	name := "cloudflare_page_rule.test"
	target := fmt.Sprintf("test-import.%s", os.Getenv("CLOUDFLARE_DOMAIN"))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigFullySpecified(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists(name, &pageRule),
				),
			},
			{
				ResourceName:        name,
				ImportStateIdPrefix: fmt.Sprintf("%s/", zoneID),
				ImportState:         true,
				ImportStateVerify:   true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists(name, &pageRule),
				),
			},
		},
	})
}

func TestAccCloudflarePageRule_ImportWithBrowserCacheTTL30(t *testing.T) {
	var pageRule cloudflare.PageRule
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	name := "cloudflare_page_rule.test"
	testAccRunResourceTestSteps(t, []resource.TestStep{
		{
			Config: buildPageRuleConfig("test", "browser_cache_ttl = 30"),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckCloudflarePageRuleExists(name, &pageRule),
			),
		},
		{
			ResourceName:        name,
			ImportStateIdPrefix: fmt.Sprintf("%s/", zoneID),
			ImportState:         true,
			ImportStateVerify:   true,
			Check: resource.ComposeTestCheckFunc(
				testAccCheckCloudflarePageRuleExists(name, &pageRule),
			),
		},
	})
}

func TestAccCloudflarePageRule_ImportWithoutBrowserCacheTTL(t *testing.T) {
	var pageRule cloudflare.PageRule
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	name := "cloudflare_page_rule.test"
	testAccRunResourceTestSteps(t, []resource.TestStep{
		{
			Config: buildPageRuleConfig("test", `browser_check = "on"`),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckCloudflarePageRuleExists(name, &pageRule),
			),
		},
		{
			ResourceName:        name,
			ImportStateIdPrefix: fmt.Sprintf("%s/", zoneID),
			ImportState:         true,
			ImportStateVerify:   true,
			Check: resource.ComposeTestCheckFunc(
				testAccCheckCloudflarePageRuleExists(name, &pageRule),
			),
		},
	})
}
