package cloudflare

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAPIToken(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the API token
	// endpoint does not yet support the API tokens without an explicit scope.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	resourceID := "cloudflare_api_token." + rnd
	permissionID := "82e64a83756745bbbb1c9c2701bf816b" // DNS read

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAPITokenWithoutCondition(rnd, rnd, permissionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "name", rnd),
				),
			},
			{
				Config: testAccCloudflareAPITokenWithoutCondition(rnd, rnd+"-updated", permissionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "name", rnd+"-updated"),
				),
			},
		},
	})
}

func TestAccAPITokenAllowDeny(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the API token
	// endpoint does not yet support the API tokens without an explicit scope.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	permissionID := "82e64a83756745bbbb1c9c2701bf816b" // DNS read

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAPITokenConfigAllowDeny(rnd, rnd, permissionID, zoneID, false),
			},
			{
				Config: testAPITokenConfigAllowDeny(rnd, rnd, permissionID, zoneID, true),
			},
			{
				Config: testAPITokenConfigAllowDeny(rnd, rnd, permissionID, zoneID, false),
			},
		},
	})
}

func TestAccAPITokenDoesNotSetConditions(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the API token
	// endpoint does not yet support the API tokens without an explicit scope.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	name := "cloudflare_api_token." + rnd
	permissionID := "82e64a83756745bbbb1c9c2701bf816b" // DNS read

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAPITokenWithoutCondition(rnd, rnd, permissionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckNoResourceAttr(name, "condition.0.request_ip.0.in"),
					resource.TestCheckNoResourceAttr(name, "condition.0.request_ip.0.not_in"),
				),
			},
		},
	})
}

func testAccCloudflareAPITokenWithoutCondition(resourceName, rnd, permissionID string) string {
	return fmt.Sprintf(`
	resource "cloudflare_api_token" "%[1]s" {
		name = "%[2]s"

		policy {
			effect = "allow"
			permission_groups = [ "%[3]s" ]
			resources = { "com.cloudflare.api.account.zone.*" = "*" }
		}
	}
`, resourceName, rnd, permissionID)
}

func TestAccAPITokenSetIndividualCondition(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the API token
	// endpoint does not yet support the API tokens without an explicit scope.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	name := "cloudflare_api_token." + rnd
	permissionID := "82e64a83756745bbbb1c9c2701bf816b" // DNS read

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAPITokenWithIndividualCondition(rnd, permissionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "condition.0.request_ip.0.in.0", "192.0.2.1/32"),
					resource.TestCheckNoResourceAttr(name, "condition.0.request_ip.0.not_in"),
				),
			},
		},
	})
}

func testAccCloudflareAPITokenWithIndividualCondition(rnd string, permissionID string) string {
	return fmt.Sprintf(`
	resource "cloudflare_api_token" "%[1]s" {
		name = "%[1]s"

		policy {
			effect = "allow"
			permission_groups = [ "%[2]s" ]
			resources = { "com.cloudflare.api.account.zone.*" = "*" }
		}

		condition {
			request_ip {
				in = ["192.0.2.1/32"]
			}
		}
	}
`, rnd, permissionID)
}

func TestAccAPITokenSetAllCondition(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the API token
	// endpoint does not yet support the API tokens without an explicit scope.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	name := "cloudflare_api_token." + rnd
	permissionID := "82e64a83756745bbbb1c9c2701bf816b" // DNS read

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAPITokenWithAllCondition(rnd, permissionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "condition.0.request_ip.0.in.0", "192.0.2.1/32"),
					resource.TestCheckResourceAttr(name, "condition.0.request_ip.0.not_in.0", "198.51.100.1/32"),
				),
			},
		},
	})
}

func testAccCloudflareAPITokenWithAllCondition(rnd string, permissionID string) string {
	return fmt.Sprintf(`
	resource "cloudflare_api_token" "%[1]s" {
		name = "%[1]s"

		policy {
			effect = "allow"
			permission_groups = [ "%[2]s" ]
			resources = { "com.cloudflare.api.account.zone.*" = "*" }
		}

		condition {
			request_ip {
				in     = ["192.0.2.1/32"]
				not_in = ["198.51.100.1/32"]
			}
		}
	}
`, rnd, permissionID)
}

func testAPITokenConfigAllowDeny(resourceID, name, permissionID, zoneID string, allowAllZonesExceptOne bool) string {
	var add string
	if allowAllZonesExceptOne {
		add = fmt.Sprintf(`
    		policy {
			  effect = "deny"
			  permission_groups = [
			    "%[1]s",
			  ]
			  resources = {
			    "com.cloudflare.api.account.zone.%[2]s" = "*"
			  }
	    	}
	  `, permissionID, zoneID)
	}

	return fmt.Sprintf(`
		resource "cloudflare_api_token" "%[1]s" {
		  name = "%[2]s"

		  policy {
			effect = "allow"
			permission_groups = [
			  "%[3]s",
			]
			resources = {
			  "com.cloudflare.api.account.zone.*" = "*"
			}
		  }
		  %[4]s
		}
		`, resourceID, name, permissionID, add)
}
