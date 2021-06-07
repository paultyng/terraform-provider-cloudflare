package cloudflare

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareApiTokenPermissionGroups(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the API token
	// permission groups endpoint does not yet support the API tokens and it
	// results in misleading state error messages.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudflareApiTokenPermissionGroupsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCloudflareApiTokenPermissionGroups("data.cloudflare_api_token_permission_groups.some"),
				),
			},
		},
	})
}

func testAccCloudflareApiTokenPermissionGroups(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		apiTokenReadId, ok := a["permissions.API Tokens Read"]
		if !ok {
			return fmt.Errorf("couldn't get 'API Tokens Read' permission ID")
		}

		// PermissionGroupsIDs can be found at
		// https://developers.cloudflare.com/api/tokens/create/permissions
		apiTokenReadIdShouldBe := "0cc3a61731504c89b99ec1be78b77aa0"

		if apiTokenReadId != apiTokenReadIdShouldBe {
			return fmt.Errorf("ApiTokenPermissionGroups 'API Tokens Read' is '%s', but should be '%s'",
				apiTokenReadId,
				apiTokenReadIdShouldBe,
			)
		}

		return nil
	}
}

const testAccCloudflareApiTokenPermissionGroupsConfig = `
data "cloudflare_api_token_permission_groups" "some" {}
`
