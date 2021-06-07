package cloudflare

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	zoneID = os.Getenv("CLOUDFLARE_ZONE_ID")
	domain = os.Getenv("CLOUDFLARE_DOMAIN")
)

func TestAccCloudflareAccessApplicationBasic(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_application.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessApplicationConfigBasic(rnd, domain, AccessIdentifier{Type: ZoneType, Value: zoneID}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "domain", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(name, "session_duration", "24h"),
					resource.TestCheckResourceAttr(name, "cors_headers.#", "0"),
					resource.TestCheckResourceAttr(name, "auto_redirect_to_identity", "false"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessApplicationConfigBasic(rnd, domain, AccessIdentifier{Type: AccountType, Value: accountID}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "domain", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(name, "session_duration", "24h"),
					resource.TestCheckResourceAttr(name, "cors_headers.#", "0"),
					resource.TestCheckResourceAttr(name, "auto_redirect_to_identity", "false"),
				),
			},
		},
	})
}

func TestAccCloudflareAccessApplicationWithCORS(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_application.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessApplicationConfigWithCORS(rnd, zoneID, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "domain", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(name, "session_duration", "24h"),
					resource.TestCheckResourceAttr(name, "cors_headers.#", "1"),
					resource.TestCheckResourceAttr(name, "cors_headers.0.allowed_methods.#", "3"),
					resource.TestCheckResourceAttr(name, "cors_headers.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(name, "cors_headers.0.max_age", "10"),
					resource.TestCheckResourceAttr(name, "auto_redirect_to_identity", "false"),
				),
			},
		},
	})
}

func TestAccCloudflareAccessApplicationWithAutoRedirectToIdentity(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_application.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessApplicationConfigWithAutoRedirectToIdentity(rnd, zoneID, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "domain", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(name, "session_duration", "24h"),
					resource.TestCheckResourceAttr(name, "auto_redirect_to_identity", "true"),
				),
			},
		},
	})
}

func TestAccCloudflareAccessApplicationWithEnableBindingCookie(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_application.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessApplicationConfigWithEnableBindingCookie(rnd, zoneID, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "domain", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(name, "session_duration", "24h"),
					resource.TestCheckResourceAttr(name, "enable_binding_cookie", "true"),
				),
			},
		},
	})
}

func TestAccCloudflareAccessApplicationWithCustomDenyFields(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_application.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessApplicationConfigWithCustomDenyFields(rnd, zoneID, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "domain", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(name, "session_duration", "24h"),
					resource.TestCheckResourceAttr(name, "custom_deny_message", "denied!"),
					resource.TestCheckResourceAttr(name, "custom_deny_url", "https://www.cloudflare.com"),
				),
			},
		},
	})
}

func TestAccCloudflareAccessApplicationWithADefinedIdps(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_application.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessApplicationConfigWithADefinedIdp(rnd, zoneID, domain, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "domain", fmt.Sprintf("%s.%s", rnd, domain)),
					resource.TestCheckResourceAttr(name, "session_duration", "24h"),
					resource.TestCheckResourceAttr(name, "auto_redirect_to_identity", "true"),
					resource.TestCheckResourceAttr(name, "allowed_idps.#", "1"),
				),
			},
		},
	})
}

func testAccCloudflareAccessApplicationConfigBasic(rnd string, domain string, identifier AccessIdentifier) string {
	return fmt.Sprintf(`
resource "cloudflare_access_application" "%[1]s" {
  %[3]s_id                  = "%[4]s"
  name                      = "%[1]s"
  domain                    = "%[1]s.%[2]s"
  session_duration          = "24h"
  auto_redirect_to_identity = false
}
`, rnd, domain, identifier.Type, identifier.Value)
}

func testAccCloudflareAccessApplicationConfigWithCORS(rnd, zoneID, domain string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_application" "%[1]s" {
  zone_id          = "%[2]s"
  name             = "%[1]s"
  domain           = "%[1]s.%[3]s"
  session_duration = "24h"
  cors_headers {
    allowed_methods = ["GET", "POST", "OPTIONS"]
    allowed_origins = ["https://example.com"]
    allow_credentials = true
    max_age = 10
  }
  auto_redirect_to_identity = false
}
`, rnd, zoneID, domain)
}

func testAccCloudflareAccessApplicationConfigWithAutoRedirectToIdentity(rnd, zoneID, domain string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_application" "%[1]s" {
  zone_id                   = "%[2]s"
  name                      = "%[1]s"
  domain                    = "%[1]s.%[3]s"
  session_duration          = "24h"
  auto_redirect_to_identity = true
}
`, rnd, zoneID, domain)
}

func testAccCloudflareAccessApplicationConfigWithEnableBindingCookie(rnd, zoneID, domain string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_application" "%[1]s" {
  zone_id                   = "%[2]s"
  name                      = "%[1]s"
  domain                    = "%[1]s.%[3]s"
  session_duration          = "24h"
  enable_binding_cookie     = true
}
`, rnd, zoneID, domain)
}

func testAccCloudflareAccessApplicationConfigWithCustomDenyFields(rnd, zoneID, domain string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_application" "%[1]s" {
  zone_id                   = "%[2]s"
  name                      = "%[1]s"
  domain                    = "%[1]s.%[3]s"
  session_duration          = "24h"
  custom_deny_message       = "denied!"
  custom_deny_url           = "https://www.cloudflare.com"
}
`, rnd, zoneID, domain)
}

func testAccCloudflareAccessApplicationConfigWithADefinedIdp(rnd, zoneID, domain string, accountID string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_identity_provider" "%[1]s" {
  account_id = "%[4]s"
  name = "%[1]s"
  type = "onetimepin"
}
resource "cloudflare_access_application" "%[1]s" {
  zone_id                   = "%[2]s"
  name                      = "%[1]s"
  domain                    = "%[1]s.%[3]s"
  session_duration          = "24h"
  auto_redirect_to_identity = true
  allowed_idps              = [cloudflare_access_identity_provider.%[1]s.id]
}
`, rnd, zoneID, domain, accountID)
}

func testAccCheckCloudflareAccessApplicationDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_access_application" {
			continue
		}

		_, err := client.AccessApplication(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("AccessApplication still exists")
		}

		_, err = client.AccessApplication(context.Background(), rs.Primary.Attributes["account_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("AccessApplication still exists")
		}
	}

	return nil
}

func TestAccCloudflareAccessApplicationWithZoneID(t *testing.T) {
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
	name := "cloudflare_access_application." + rnd
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	updatedName := fmt.Sprintf("%s-updated", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccessApplicationWithZoneID(rnd, zone, zoneID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
				),
			},
			{
				Config: testAccessApplicationWithZoneIDUpdated(rnd, zone, zoneID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", updatedName),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
				),
			},
		},
	})
}

func TestAccCloudflareAccessApplicationWithMissingCORSMethods(t *testing.T) {
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
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccessApplicationWithMissingCORSMethods(rnd, zone, zoneID),
				ExpectError: regexp.MustCompile("must set allowed_methods or allow_all_methods"),
			},
		},
	})
}

func TestAccCloudflareAccessApplicationWithMissingCORSOrigins(t *testing.T) {
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
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccessApplicationWithMissingCORSOrigins(rnd, zone, zoneID),
				ExpectError: regexp.MustCompile("must set allowed_origins or allow_all_origins"),
			},
		},
	})
}

func TestAccCloudflareAccessApplicationWithInvalidSessionDuration(t *testing.T) {
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
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccessApplicationWithInvalidSessionDuration(rnd, zone, zoneID),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`"session_duration" only supports "ns", "us" (or "µs"), "ms", "s", "m", or "h" as valid units.`)),
			},
		},
	})
}

func TestAccessApplicationMisconfiguredCORSCredentialsAllowingAllOrigins(t *testing.T) {
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
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccessApplicationMisconfiguredCORSAllowAllOriginsWithCredentials(rnd, zone, zoneID),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`CORS credentials are not permitted when all origins are allowed`)),
			},
		},
	})
}

func TestAccessApplicationMisconfiguredCORSCredentialsAllowingWildcardOrigins(t *testing.T) {
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
	zone := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccessApplicationMisconfiguredCORSAllowWildcardOriginWithCredentials(rnd, zone, zoneID),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`CORS credentials are not permitted when all origins are allowed`)),
			},
		},
	})
}

func testAccessApplicationWithZoneID(resourceID, zone, zoneID string) string {
	return fmt.Sprintf(`
    resource "cloudflare_access_application" "%[1]s" {
      name    = "%[1]s"
      zone_id = "%[3]s"
      domain  = "%[1]s.%[2]s"
    }
  `, resourceID, zone, zoneID)
}

func testAccessApplicationWithZoneIDUpdated(resourceID, zone, zoneID string) string {
	return fmt.Sprintf(`
    resource "cloudflare_access_application" "%[1]s" {
      name    = "%[1]s-updated"
      zone_id = "%[3]s"
      domain  = "%[1]s.%[2]s"
    }
  `, resourceID, zone, zoneID)
}

func testAccessApplicationWithMissingCORSMethods(resourceID, zone, zoneID string) string {
	return fmt.Sprintf(`
    resource "cloudflare_access_application" "%[1]s" {
      name    = "%[1]s-updated"
      zone_id = "%[3]s"
      domain  = "%[1]s.%[2]s"

    cors_headers {
      allow_all_origins = true
    }
  }
  `, resourceID, zone, zoneID)
}

func testAccessApplicationWithMissingCORSOrigins(resourceID, zone, zoneID string) string {
	return fmt.Sprintf(`
    resource "cloudflare_access_application" "%[1]s" {
      name    = "%[1]s-updated"
      zone_id = "%[3]s"
      domain  = "%[1]s.%[2]s"

    cors_headers {
      allow_all_methods = true
    }
  }
  `, resourceID, zone, zoneID)
}

func testAccessApplicationWithInvalidSessionDuration(resourceID, zone, zoneID string) string {
	return fmt.Sprintf(`
    resource "cloudflare_access_application" "%[1]s" {
      name             = "%[1]s-updated"
      zone_id          = "%[3]s"
      domain           = "%[1]s.%[2]s"
      session_duration = "24z"
  }
  `, resourceID, zone, zoneID)
}

func testAccessApplicationMisconfiguredCORSAllowAllOriginsWithCredentials(resourceID, zone, zoneID string) string {
	return fmt.Sprintf(`
    resource "cloudflare_access_application" "%[1]s" {
      name             = "%[1]s-updated"
      zone_id          = "%[3]s"
      domain           = "%[1]s.%[2]s"

      cors_headers {
        allowed_methods = ["GET"]
        allow_all_origins = true
        allow_credentials = true
      }
  }
  `, resourceID, zone, zoneID)
}

func testAccessApplicationMisconfiguredCORSAllowWildcardOriginWithCredentials(resourceID, zone, zoneID string) string {
	return fmt.Sprintf(`
    resource "cloudflare_access_application" "%[1]s" {
      name             = "%[1]s-updated"
      zone_id          = "%[3]s"
      domain           = "%[1]s.%[2]s"

      cors_headers {
        allowed_methods = ["GET"]
        allowed_origins = ["*"]
        allow_credentials = true
      }
  }
  `, resourceID, zone, zoneID)
}
