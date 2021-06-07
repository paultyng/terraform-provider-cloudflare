package cloudflare

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareWorkersKVNamespace_Basic(t *testing.T) {
	t.Parallel()
	var namespace cloudflare.WorkersKVNamespace
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_workers_kv_namespace." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCloudflareWorkersKVNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWorkersKVNamespace(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkersKVNamespaceExists(rnd, &namespace),
					resource.TestCheckResourceAttr(
						resourceName, "title", rnd,
					),
				),
			},
		},
	})
}

func testAccCloudflareWorkersKVNamespaceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_workers_kv_namespace" {
			continue
		}

		resp, err := client.ListWorkersKVNamespaces(context.Background())

		if err == nil {
			return err
		}

		for _, n := range resp {
			if n.ID == rs.Primary.ID {
				return fmt.Errorf("Namespace still exists")
			}
		}
	}

	return nil
}

func testAccCheckCloudflareWorkersKVNamespace(rName string) string {
	return fmt.Sprintf(`
resource "cloudflare_workers_kv_namespace" "%[1]s" {
	title = "%[1]s"
}`, rName)
}

func testAccCheckCloudflareWorkersKVNamespaceExists(title string, namespace *cloudflare.WorkersKVNamespace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*cloudflare.API)
		resp, err := client.ListWorkersKVNamespaces(context.Background())
		if err != nil {
			return err
		}

		for _, n := range resp {
			if n.Title == title {
				*namespace = n
				return nil
			}
		}

		return fmt.Errorf("namespace not found")
	}
}
