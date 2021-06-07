package cloudflare

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAccessServiceTokenCreate(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the Access
	// Service Tokens endpoint does not yet support the API tokens and it
	// results in misleading state error messages.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_service_token.tf-acc-%s", rnd)
	resourceName := strings.Split(name, ".")[1]

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccessAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testCloudflareAccessServiceTokenBasicConfig(resourceName, resourceName, AccessIdentifier{Type: AccountType, Value: accountID}, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "name", resourceName),
					resource.TestCheckResourceAttrSet(name, "client_id"),
					resource.TestCheckResourceAttrSet(name, "client_secret"),
					resource.TestCheckResourceAttrSet(name, "expires_at"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccessAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testCloudflareAccessServiceTokenBasicConfig(resourceName, resourceName, AccessIdentifier{Type: ZoneType, Value: zoneID}, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "name", resourceName),
					resource.TestCheckResourceAttrSet(name, "client_id"),
					resource.TestCheckResourceAttrSet(name, "client_secret"),
					resource.TestCheckResourceAttrSet(name, "expires_at"),
				),
			},
		},
	})
}

func TestAccAccessServiceTokenUpdate(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the Access
	// Service Tokens endpoint does not yet support the API tokens and it
	// results in misleading state error messages.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_service_token.tf-acc-%s", rnd)
	resourceName := strings.Split(name, ".")[1]

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testCloudflareAccessServiceTokenBasicConfig(resourceName, resourceName, AccessIdentifier{Type: AccountType, Value: accountID}, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", resourceName),
				),
			},
			{
				Config: testCloudflareAccessServiceTokenBasicConfig(resourceName, resourceName+"-updated", AccessIdentifier{Type: AccountType, Value: accountID}, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", resourceName+"-updated"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccessAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testCloudflareAccessServiceTokenBasicConfig(resourceName, resourceName, AccessIdentifier{Type: ZoneType, Value: zoneID}, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", resourceName),
				),
			},
			{
				Config: testCloudflareAccessServiceTokenBasicConfig(resourceName, resourceName+"-updated", AccessIdentifier{Type: ZoneType, Value: zoneID}, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", resourceName+"-updated"),
				),
			},
		},
	})
}

func TestAccAccessServiceTokenUpdateWithExpiration(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the Access
	// Service Tokens endpoint does not yet support the API tokens and it
	// results in misleading state error messages.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	var initialState terraform.ResourceState

	name := fmt.Sprintf("cloudflare_access_service_token.tf-acc-%s", rnd)
	resourceName := strings.Split(name, ".")[1]
	expirationTime := 365

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccessAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testCloudflareAccessServiceTokenBasicConfig(resourceName, resourceName, AccessIdentifier{Type: ZoneType, Value: zoneID}, expirationTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessServiceTokenSaved(name, &initialState),
					resource.TestCheckResourceAttr(name, "min_days_for_renewal", strconv.Itoa(expirationTime)),
				),
				//Expiration of 365 will always force a new resource as long as the tokens expire in 365 days in cloudflare
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testCloudflareAccessServiceTokenBasicConfig(resourceName, resourceName, AccessIdentifier{Type: ZoneType, Value: zoneID}, expirationTime),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "min_days_for_renewal", strconv.Itoa(expirationTime)),
					testAccCheckCloudflareAccessServiceTokenRenewed(name, &initialState),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudflareAccessServiceTokenSaved(n string, resourceState *terraform.ResourceState) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Access Token ID is set")
		}

		*resourceState = *rs

		return nil
	}
}

func testAccCheckCloudflareAccessServiceTokenRenewed(n string, oldResourceState *terraform.ResourceState) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Access Token ID is set")
		}

		for _, attribute := range []string{"expires_at", "client_secret"} {
			if rs.Primary.Attributes[attribute] == oldResourceState.Primary.Attributes[attribute] {
				return fmt.Errorf("Resource attribute '%s' has not changed. Expected change between old state and new", attribute)
			}
		}

		return nil
	}
}

func TestAccAccessServiceTokenDelete(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the Access
	// Service Tokens endpoint does not yet support the API tokens and it
	// results in misleading state error messages.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_service_token.tf-acc-%s", rnd)
	resourceName := strings.Split(name, ".")[1]

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessServiceTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCloudflareAccessServiceTokenBasicConfig(resourceName, resourceName, AccessIdentifier{Type: AccountType, Value: accountID}, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "name", resourceName),
					resource.TestCheckResourceAttrSet(name, "client_id"),
					resource.TestCheckResourceAttrSet(name, "client_secret"),
					resource.TestCheckResourceAttrSet(name, "expires_at"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccessAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessServiceTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCloudflareAccessServiceTokenBasicConfig(resourceName, resourceName, AccessIdentifier{Type: ZoneType, Value: zoneID}, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "name", resourceName),
					resource.TestCheckResourceAttrSet(name, "client_id"),
					resource.TestCheckResourceAttrSet(name, "client_secret"),
					resource.TestCheckResourceAttrSet(name, "expires_at"),
				),
			},
		},
	})
}

func testCloudflareAccessServiceTokenBasicConfig(resourceName string, tokenName string, identifier AccessIdentifier, minDaysForRenewal int) string {
	return fmt.Sprintf(`
resource "cloudflare_access_service_token" "%[1]s" {
  %[3]s_id = "%[4]s"
  name     = "%[2]s"
  min_days_for_renewal ="%[5]d"
}`, resourceName, tokenName, identifier.Type, identifier.Value, minDaysForRenewal)
}

func testAccCheckCloudflareAccessServiceTokenDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_access_service_token" {
			continue
		}

		_, err := client.DeleteAccessServiceToken(context.Background(), rs.Primary.Attributes["account_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("access service token still exists")
		}

		_, err = client.DeleteZoneLevelAccessServiceToken(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("access service token still exists")
		}
	}

	return nil
}
