package cloudflare

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("cloudflare_access_group", &resource.Sweeper{
		Name: "cloudflare_access_group",
		F:    testSweepCloudflareAccessGroups,
	})
}

func testSweepCloudflareAccessGroups(r string) error {
	client, clientErr := sharedClient()
	if clientErr != nil {
		log.Printf("[ERROR] Failed to create Cloudflare client: %s", clientErr)
	}

	// Zone level Access Groups
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	zoneAccessGroups, _, err := client.ZoneLevelAccessGroups(context.Background(), zoneID, cloudflare.PaginationOptions{})
	if err != nil {
		log.Printf("[ERROR] Failed to fetch zone level Access Groups: %s", err)
	}

	if len(zoneAccessGroups) == 0 {
		log.Print("[DEBUG] No Cloudflare zone level Access Groups to sweep")
		return nil
	}

	for _, accessGroup := range zoneAccessGroups {
		if err := client.DeleteZoneLevelAccessGroup(context.Background(), zoneID, accessGroup.ID); err != nil {
			log.Printf("[ERROR] Failed to delete zone level Access Group %s", accessGroup.ID)
		}
	}

	// Account level Access Groups
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	accountAccessGroups, _, err := client.AccessGroups(context.Background(), accountID, cloudflare.PaginationOptions{})
	if err != nil {
		log.Printf("[ERROR] Failed to fetch account level Access Groups: %s", err)
	}

	if len(accountAccessGroups) == 0 {
		log.Print("[DEBUG] No Cloudflare account level Access Groups to sweep")
		return nil
	}

	for _, accessGroup := range accountAccessGroups {
		if err := client.DeleteAccessGroup(context.Background(), accountID, accessGroup.ID); err != nil {
			log.Printf("[ERROR] Failed to delete account level Access Group %s", accessGroup.ID)
		}
	}

	return nil
}

var (
	accountID   = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	email       = "test@example.com"
	accessGroup cloudflare.AccessGroup
)

func TestAccCloudflareAccessGroupConfig_Basic(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_group.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessGroupConfigBasic(rnd, email, AccessIdentifier{Type: AccountType, Value: accountID}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessGroupExists(name, AccessIdentifier{Type: AccountType, Value: accountID}, &accessGroup),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "include.0.email.0", email),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessGroupConfigBasic(rnd, email, AccessIdentifier{Type: ZoneType, Value: zoneID}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessGroupExists(name, AccessIdentifier{Type: ZoneType, Value: zoneID}, &accessGroup),
					resource.TestCheckResourceAttr(name, "zone_id", zoneID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "include.0.email.0", email),
				),
			},
		},
	})
}

func TestAccCloudflareAccessGroup_Exclude(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_group.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccessGroupConfigExclude(rnd, accountID, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessGroupExists(name, AccessIdentifier{Type: AccountType, Value: accountID}, &accessGroup),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "include.0.email.0", email),
					resource.TestCheckResourceAttr(name, "exclude.0.email.0", email),
				),
			},
		},
	})
}

func TestAccCloudflareAccessGroup_Require(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_group.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccessGroupConfigRequire(rnd, accountID, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessGroupExists(name, AccessIdentifier{Type: AccountType, Value: accountID}, &accessGroup),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "include.0.email.0", email),
					resource.TestCheckResourceAttr(name, "require.0.email.0", email),
				),
			},
		},
	})
}

func TestAccCloudflareAccessGroup_FullConfig(t *testing.T) {
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_group.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccessGroupConfigFullConfig(rnd, accountID, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessGroupExists(name, AccessIdentifier{Type: AccountType, Value: accountID}, &accessGroup),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "include.0.email.0", email),
					resource.TestCheckResourceAttr(name, "exclude.0.email.0", email),
					resource.TestCheckResourceAttr(name, "require.0.email.0", email),
				),
			},
		},
	})
}

func TestAccCloudflareAccessGroupWithIDP(t *testing.T) {
	rnd := generateRandomResourceName()
	groupName := fmt.Sprintf("cloudflare_access_group.%s", rnd)
	githubOrg := "Terraform-Cloudflare-Provider-Test-Org"
	team := "test-team-1"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessGroupWithIDP(accountID, rnd, githubOrg, team),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessGroupExists(groupName, AccessIdentifier{Type: AccountType, Value: accountID}, &accessGroup),
					resource.TestCheckResourceAttr(groupName, "account_id", accountID),
					resource.TestCheckResourceAttr(groupName, "name", rnd),
					resource.TestCheckResourceAttrSet(groupName, "include.0.github.0.identity_provider_id"),
					resource.TestCheckResourceAttr(groupName, "include.0.github.0.name", githubOrg),
					resource.TestCheckResourceAttr(groupName, "include.0.github.0.teams.0", team),
				),
			},
		},
	})
}

func TestAccCloudflareAccessGroup_Updated(t *testing.T) {
	var before, after cloudflare.AccessGroup
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_group.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessGroupConfigBasic(rnd, email, AccessIdentifier{Type: AccountType, Value: accountID}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessGroupExists(name, AccessIdentifier{Type: AccountType, Value: accountID}, &before),
				),
			},
			{
				Config: testAccCloudflareAccessGroupConfigBasic(rnd, "test-changed@example.com", AccessIdentifier{Type: AccountType, Value: accountID}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessGroupExists(name, AccessIdentifier{Type: AccountType, Value: accountID}, &after),
					testAccCheckCloudflareAccessGroupIDUnchanged(&before, &after),
					resource.TestCheckResourceAttr(name, "include.0.email.0", "test-changed@example.com"),
				),
			},
		},
	})
}

func TestAccCloudflareAccessGroup_CreateAfterManualDestroy(t *testing.T) {
	var before, after cloudflare.AccessGroup
	var initialID string
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_access_group.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccessAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareAccessGroupConfigBasic(rnd, email, AccessIdentifier{Type: AccountType, Value: accountID}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessGroupExists(name, AccessIdentifier{Type: AccountType, Value: accountID}, &before),
					testAccManuallyDeleteAccessGroup(name, &initialID),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCloudflareAccessGroupConfigBasicWithUpdate(rnd, accountID, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareAccessGroupExists(name, AccessIdentifier{Type: AccountType, Value: accountID}, &after),
					testAccCheckCloudflareAccessGroupRecreated(&before, &after),
					resource.TestCheckResourceAttr(name, "account_id", accountID),
					resource.TestCheckResourceAttr(name, "name", fmt.Sprintf("%s-updated", rnd)),
					resource.TestCheckResourceAttr(name, "include.0.email.0", email),
				),
			},
		},
	})
}

func testAccCloudflareAccessGroupConfigBasic(resourceName string, email string, identifier AccessIdentifier) string {
	return fmt.Sprintf(`
resource "cloudflare_access_group" "%[1]s" {
  %[3]s_id = "%[4]s"
  name     = "%[1]s"

  include {
    email = ["%[2]s"]
  }
}`, resourceName, email, identifier.Type, identifier.Value)
}

func testAccCloudflareAccessGroupConfigBasicWithUpdate(resourceName, accountID, email string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_group" "%[1]s" {
  account_id = "%[2]s"
  name = "%[1]s-updated"

  include {
    email = ["%[3]s"]
  }
}`, resourceName, accountID, email)
}

func testAccessGroupConfigExclude(resourceName, accountID, email string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_group" "%[1]s" {
  account_id = "%[2]s"
  name = "%[1]s"

  include {
    email = ["%[3]s"]
  }

  exclude {
    email = ["%[3]s"]
  }
}`, resourceName, accountID, email)
}

func testAccessGroupConfigRequire(resourceName, accountID, email string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_group" "%[1]s" {
  account_id = "%[2]s"
  name = "%[1]s"

  include {
    email = ["%[3]s"]
  }

  require {
    email = ["%[3]s"]
  }
}`, resourceName, accountID, email)
}

func testAccessGroupConfigFullConfig(resourceName, accountID, email string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_group" "%[1]s" {
  account_id = "%[2]s"
  name = "%[1]s"

  include {
    email = ["%[3]s"]
  }

  require {
    email = ["%[3]s"]
  }

  exclude {
    email = ["%[3]s"]
  }
}`, resourceName, accountID, email)
}

func testAccCloudflareAccessGroupWithIDP(accountID, rnd, githubOrg, team string) string {
	return fmt.Sprintf(`
resource "cloudflare_access_identity_provider" "%[2]s" {
  account_id = "%[1]s"
  name = "%[2]s"
  type = "github"
  config {
    client_id = "test"
    client_secret = "secret"
  }
}

resource "cloudflare_access_group" "%[2]s" {
  account_id = "%[1]s"
  name = "%[2]s"

  include {
    github {
      name                 = "%[3]s"
      teams                = ["%[4]s"]
      identity_provider_id = cloudflare_access_identity_provider.%[2]s.id
    }
  }
}`, accountID, rnd, githubOrg, team)
}

func testAccCheckCloudflareAccessGroupExists(n string, accessIdentifier AccessIdentifier, accessGroup *cloudflare.AccessGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AccessGroup ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		var foundAccessGroup cloudflare.AccessGroup
		var err error

		if accessIdentifier.Type == AccountType {
			foundAccessGroup, err = client.AccessGroup(context.Background(), rs.Primary.Attributes["account_id"], rs.Primary.ID)
			if err != nil {
				return err
			}
		} else {
			foundAccessGroups, _, err := client.ZoneLevelAccessGroups(context.Background(), rs.Primary.Attributes["zone_id"], cloudflare.PaginationOptions{})
			if err != nil {
				return err
			}
			foundAccessGroup = foundAccessGroups[0]
		}

		if foundAccessGroup.ID != rs.Primary.ID {
			return fmt.Errorf("AccessGroup not found")
		}

		*accessGroup = foundAccessGroup

		return nil
	}
}

func testAccCheckCloudflareAccessGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_access_group" {
			continue
		}

		_, err := client.AccessGroup(context.Background(), rs.Primary.Attributes["account_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("AccessGroup still exists")
		}

		_, err = client.AccessGroup(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("AccessGroup still exists")
		}
	}

	return nil
}

func testAccCheckCloudflareAccessGroupIDUnchanged(before, after *cloudflare.AccessGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID != after.ID {
			return fmt.Errorf("ID should not change suring in place update, but got change %s -> %s", before.ID, after.ID)
		}
		return nil
	}
}

func testAccManuallyDeleteAccessGroup(name string, initialID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		*initialID = rs.Primary.ID
		err := client.DeleteAccessGroup(context.Background(), rs.Primary.Attributes["account_id"], rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckCloudflareAccessGroupRecreated(before, after *cloudflare.AccessGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID == after.ID {
			return fmt.Errorf("Expected change of AccessGroup Ids, but both were %v", before.ID)
		}
		return nil
	}
}
