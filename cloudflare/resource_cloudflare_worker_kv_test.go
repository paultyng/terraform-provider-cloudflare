package cloudflare

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareWorkersKV_Basic(t *testing.T) {
	t.Parallel()
	var kvPair cloudflare.WorkersKVPair
	name := generateRandomResourceName()
	key := generateRandomResourceName()
	value := generateRandomResourceName()
	resourceName := "cloudflare_workers_kv." + name

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckAccount(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCloudflareWorkersKVDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWorkersKV(name, key, value),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkersKVExists(key, &kvPair),
					resource.TestCheckResourceAttr(
						resourceName, "value", value,
					),
				),
			},
		},
	})
}

func testAccCloudflareWorkersKVDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_workers_kv" {
			continue
		}

		namespaceID := rs.Primary.Attributes["namespace_id"]
		key := rs.Primary.Attributes["key"]

		_, err := client.ReadWorkersKV(context.Background(), namespaceID, key)

		if err == nil {
			return fmt.Errorf("workers kv pair still exists")
		}
	}

	return nil
}

func testAccCheckCloudflareWorkersKV(rName string, key string, value string) string {
	return testAccCheckCloudflareWorkersKVNamespace(rName) + fmt.Sprintf(`
resource "cloudflare_workers_kv" "%[1]s" {
	namespace_id = cloudflare_workers_kv_namespace.%[1]s.id
	key = "%[2]s"
	value = "%[3]s"
}`, rName, key, value)
}

func testAccCheckCloudflareWorkersKVExists(key string, kv *cloudflare.WorkersKVPair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*cloudflare.API)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "cloudflare_workers_kv" {
				continue
			}

			namespaceID := rs.Primary.Attributes["namespace_id"]
			value, err := client.ReadWorkersKV(context.Background(), namespaceID, key)
			if err != nil {
				return err
			}

			if value == nil {
				return fmt.Errorf("workers kv key %s not found in namespace %s", key, namespaceID)
			}
		}

		return nil
	}
}
