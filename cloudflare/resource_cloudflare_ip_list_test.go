package cloudflare

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareIPListExists(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the IP List
	// endpoint does not yet support the API tokens.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_ip_list.%s", rnd)
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	var IPList cloudflare.IPList

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckAccount(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareIPList(rnd, rnd, rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareIPListExists(name, &IPList),
					resource.TestCheckResourceAttr(
						name, "name", rnd),
				),
			},
		},
	})
}

func TestAccCloudflareIPListUpdateDescription(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the IP List
	// endpoint does not yet support the API tokens.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_ip_list.%s", rnd)
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	var IPList cloudflare.IPList
	var initialID string

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckAccount(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareIPList(rnd, rnd, rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareIPListExists(name, &IPList),
					resource.TestCheckResourceAttr(
						name, "description", rnd),
				),
			},
			{
				PreConfig: func() {
					initialID = IPList.ID
				},
				Config: testAccCheckCloudflareIPList(rnd, rnd, rnd+"-updated", accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareIPListExists(name, &IPList),
					func(state *terraform.State) error {
						if initialID != IPList.ID {
							return fmt.Errorf("wanted update but IPList got recreated (id changed %q -> %q)",
								initialID, IPList.ID)
						}
						return nil
					},
					resource.TestCheckResourceAttr(name, "description", rnd+"-updated"),
				),
			},
		},
	})
}

func TestAccCloudflareIPListUpdate(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the IP List
	// endpoint does not yet support the API tokens.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_ip_list.%s", rnd)
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	var IPList cloudflare.IPList
	var initialID string

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckAccount(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareIPList(rnd, rnd, rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareIPListExists(name, &IPList),
					resource.TestCheckResourceAttr(
						name, "name", rnd),
				),
			},
			{
				PreConfig: func() {
					initialID = IPList.ID
				},
				Config: testAccCheckCloudflareIPListUpdate(rnd, rnd, rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareIPListExists(name, &IPList),
					func(state *terraform.State) error {
						if initialID != IPList.ID {
							return fmt.Errorf("wanted update but IPList got recreated (id changed %q -> %q)",
								initialID, IPList.ID)
						}
						return nil
					},
					resource.TestCheckResourceAttr(name, "item.#", "2"),
				),
			},
		},
	})
}

func testAccCheckCloudflareIPListExists(n string, list *cloudflare.IPList) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IP List ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		foundIPList, err := client.GetIPList(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*list = foundIPList

		return nil
	}
}

func testAccCheckCloudflareIPList(ID, name, description, accountID string) string {
	return fmt.Sprintf(`
  resource "cloudflare_ip_list" "%[1]s" {
	account_id = "%[4]s"
    name = "%[2]s"
	description = "%[3]s"
  	kind = "ip"
  }`, ID, name, description, accountID)
}

func testAccCheckCloudflareIPListUpdate(ID, name, description, accountID string) string {
	return fmt.Sprintf(`
  resource "cloudflare_ip_list" "%[1]s" {
	account_id = "%[4]s"
    name = "%[2]s"
	description = "%[3]s"
  	kind = "ip"

  	item {
      value = "192.0.2.0"
      comment = "one"
    }

    item {
      value = "192.0.2.1"
      comment = "two"
    }
  }`, ID, name, description, accountID)
}
