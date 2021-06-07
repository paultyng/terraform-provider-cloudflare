package cloudflare

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAccessPolicyServiceToken(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the Access
	// service does not yet support the API tokens and it results in
	// misleading state error messages.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyServiceTokenConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.service_token.#", "1"),
				),
			},
		},
	})
}

func TestAccAccessPolicyAnyServiceToken(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyAnyServiceTokenConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.any_valid_service_token", "true"),
				),
			},
		},
	})
}

func TestAccAccessPolicyWithZoneID(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	updatedName := fmt.Sprintf("%s-updated", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyWithZoneID(rnd, zone, zoneID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "include.0.any_valid_service_token", "true"),
				),
			},
			{
				Config: testAccessPolicyWithZoneIDUpdated(rnd, zone, zoneID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", updatedName),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "include.0.any_valid_service_token", "true"),
				),
			},
		},
	})
}

func testAccessPolicyServiceTokenConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_service_token" "%[1]s" {
			account_id = "%[3]s"
			name       = "%[1]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "non_identity"
			precedence     = "10"

			include {
				service_token = ["${cloudflare_access_service_token.%[1]s.id}"]
			}
		}

	`, resourceID, zone, accountID)
}

func testAccessPolicyAnyServiceTokenConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "non_identity"
			precedence     = "10"

			include {
				any_valid_service_token = true
			}
		}

	`, resourceID, zone, accountID)
}

func testAccessPolicyWithZoneID(resourceID, zone, zoneID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name    = "%[1]s"
			zone_id = "%[3]s"
			domain  = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			zone_id        = "%[3]s"
			decision       = "non_identity"
			precedence     = "10"

			include {
				any_valid_service_token = true
			}
		}

	`, resourceID, zone, zoneID)
}

func testAccessPolicyWithZoneIDUpdated(resourceID, zone, zoneID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name    = "%[1]s"
			zone_id = "%[3]s"
			domain  = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s-updated"
			zone_id        = "%[3]s"
			decision       = "non_identity"
			precedence     = "10"

			include {
				any_valid_service_token = true
			}
		}

	`, resourceID, zone, zoneID)
}

func TestAccessPolicyGroup(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyGroupConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.group.#", "1"),
				),
			},
		},
	})
}

func testAccessPolicyGroupConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_group" "%[1]s" {
  			account_id     = "%[3]s"
  			name           = "%[1]s"

  			include {
				  ip = ["127.0.0.1/32"]
  			}
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "non_identity"
			precedence     = "10"

			include {
				group = [cloudflare_access_group.%[1]s.id]
			}
		}

	`, resourceID, zone, accountID)
}

func TestAccessPolicyMTLS(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyMTLSConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.certificate", "true"),
				),
			},
		},
	})
}

func testAccessPolicyMTLSConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "non_identity"
			precedence     = "10"

			include {
				certificate = true
			}
		}

	`, resourceID, zone, accountID)
}

func TestAccessPolicyCommonName(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyCommonNameConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.common_name", "example@example.com"),
				),
			},
		},
	})
}

func testAccessPolicyCommonNameConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "allow"
			precedence     = "1"

			include {
				common_name = "example@example.com"
			}
		}

	`, resourceID, zone, accountID)
}

func TestAccessPolicyEmailDomain(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyEmailDomainConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.email_domain.0", "example.com"),
				),
			},
		},
	})
}

func testAccessPolicyEmailDomainConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "allow"
			precedence     = "1"

			include {
				email_domain = ["example.com"]
			}
		}

	`, resourceID, zone, accountID)
}

func TestAccessPolicyEmails(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyEmailsConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.email.#", "2"),
					resource.TestCheckResourceAttr(name, "include.0.email.0", "a@example.com"),
					resource.TestCheckResourceAttr(name, "include.0.email.1", "b@example.com"),
				),
			},
		},
	})
}

func testAccessPolicyEmailsConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "allow"
			precedence     = "1"

			include {
				email = ["a@example.com", "b@example.com"]
			}
		}

	`, resourceID, zone, accountID)
}

func TestAccessPolicyEveryone(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyEveryoneConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.everyone", "true"),
				),
			},
		},
	})
}

func testAccessPolicyEveryoneConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "allow"
			precedence     = "1"

			include {
				everyone = true
			}
		}

	`, resourceID, zone, accountID)
}

func TestAccessPolicyIPs(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyIPsConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.ip.#", "2"),
					resource.TestCheckResourceAttr(name, "include.0.ip.0", "10.0.0.1/32"),
					resource.TestCheckResourceAttr(name, "include.0.ip.1", "10.0.0.2/32"),
				),
			},
		},
	})
}

func testAccessPolicyIPsConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "allow"
			precedence     = "1"

			include {
				ip = ["10.0.0.1/32", "10.0.0.2/32"]
			}
		}

	`, resourceID, zone, accountID)
}

func TestAccessPolicyAuthMethod(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyAuthMethodConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.auth_method", "hwk"),
				),
			},
		},
	})
}

func testAccessPolicyAuthMethodConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "allow"
			precedence     = "1"

			include {
				auth_method = "hwk"
			}
		}

	`, resourceID, zone, accountID)
}

func TestAccessPolicyGeo(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyGeoConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.geo.#", "2"),
					resource.TestCheckResourceAttr(name, "include.0.geo.0", "US"),
					resource.TestCheckResourceAttr(name, "include.0.geo.1", "AU"),
				),
			},
		},
	})
}

func testAccessPolicyGeoConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "allow"
			precedence     = "1"

			include {
				geo = ["US", "AU"]
			}
		}

	`, resourceID, zone, accountID)
}

func TestAccessPolicyOkta(t *testing.T) {
	rnd := generateRandomResourceName()
	name := "cloudflare_access_policy." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessPolicyOktaConfig(rnd, zone, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "include.0.okta.0.name.#", "2"),
					resource.TestCheckResourceAttr(name, "include.0.okta.0.name.0", "jacob-group"),
					resource.TestCheckResourceAttr(name, "include.0.okta.0.name.1", "jacob-group1"),
					resource.TestCheckResourceAttr(name, "include.0.okta.0.identity_provider_id", "225934dc-14e4-4f55-87be-f5d798d23f91"),
				),
			},
		},
	})
}

func testAccessPolicyOktaConfig(resourceID, zone, accountID string) string {
	return fmt.Sprintf(`
		resource "cloudflare_access_application" "%[1]s" {
			name       = "%[1]s"
			account_id = "%[3]s"
			domain     = "%[1]s.%[2]s"
		}

		resource "cloudflare_access_policy" "%[1]s" {
			application_id = "${cloudflare_access_application.%[1]s.id}"
			name           = "%[1]s"
			account_id     = "%[3]s"
			decision       = "allow"
			precedence     = "1"

			include {
				okta {
					name = ["jacob-group", "jacob-group1"]
					identity_provider_id = "225934dc-14e4-4f55-87be-f5d798d23f91"
				}
			}
		}

	`, resourceID, zone, accountID)
}
