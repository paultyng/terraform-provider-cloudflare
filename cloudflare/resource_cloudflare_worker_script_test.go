package cloudflare

import (
	"context"
	"fmt"
	"strings"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	scriptContent1 = `addEventListener('fetch', event => {event.respondWith(new Response('test 1'))});`
	scriptContent2 = `addEventListener('fetch', event => {event.respondWith(new Response('test 2'))});`
	encodedWasm    = "AGFzbQEAAAAGgYCAgAAA" // wat source: `(module)`, so literally just an empty wasm module
)

func TestAccCloudflareWorkerScript_MultiScriptEnt(t *testing.T) {
	t.Parallel()

	var script cloudflare.WorkerScript
	rnd := generateRandomResourceName()
	name := "cloudflare_worker_script." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareWorkerScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWorkerScriptConfigMultiScriptInitial(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkerScriptExists(name, &script, nil),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "content", scriptContent1),
				),
			},
			{
				Config: testAccCheckCloudflareWorkerScriptConfigMultiScriptUpdate(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkerScriptExists(name, &script, nil),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "content", scriptContent2),
				),
			},
			{
				Config: testAccCheckCloudflareWorkerScriptConfigMultiScriptUpdateBinding(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkerScriptExists(name, &script, []string{"MY_KV_NAMESPACE", "MY_PLAIN_TEXT", "MY_SECRET_TEXT", "MY_WASM"}),
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "content", scriptContent2),
				),
			},
		},
	})
}

func testAccCheckCloudflareWorkerScriptConfigMultiScriptInitial(rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_worker_script" "%[1]s" {
  name = "%[1]s"
  content = "%[2]s"
}`, rnd, scriptContent1)
}

func testAccCheckCloudflareWorkerScriptConfigMultiScriptUpdate(rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_worker_script" "%[1]s" {
  name = "%[1]s"
  content = "%[2]s"
}`, rnd, scriptContent2)
}

func testAccCheckCloudflareWorkerScriptConfigMultiScriptUpdateBinding(rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_workers_kv_namespace" "%[1]s" {
  title = "%[1]s"
}

resource "cloudflare_worker_script" "%[1]s" {
  name    = "%[1]s"
  content = "%[2]s"

  kv_namespace_binding {
    name         = "MY_KV_NAMESPACE"
    namespace_id = cloudflare_workers_kv_namespace.%[1]s.id
  }

  plain_text_binding {
    name = "MY_PLAIN_TEXT"
    text = "%[1]s"
  }

  secret_text_binding {
    name = "MY_SECRET_TEXT"
    text = "%[1]s"
  }

  webassembly_binding {
    name = "MY_WASM"
    module = "%[3]s"
  }
}`, rnd, scriptContent2, encodedWasm)
}

func getRequestParamsFromResource(rs *terraform.ResourceState) cloudflare.WorkerRequestParams {
	params := cloudflare.WorkerRequestParams{
		ScriptName: rs.Primary.Attributes["name"],
	}

	return params
}

func testAccCheckCloudflareWorkerScriptExists(n string, script *cloudflare.WorkerScript, bindings []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Worker Script ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		params := getRequestParamsFromResource(rs)
		r, err := client.DownloadWorker(context.Background(), &params)
		if err != nil {
			return err
		}

		if r.Script == "" {
			return fmt.Errorf("Worker Script not found")
		}

		name := strings.Replace(n, "cloudflare_worker_script.", "", -1)
		foundBindings, err := getWorkerScriptBindings(name, client)
		if err != nil {
			return fmt.Errorf("cannot list script bindings: %v", err)
		}

		for _, binding := range bindings {
			if _, ok := foundBindings[binding]; !ok {
				return fmt.Errorf("cannot find binding with name %s", binding)
			}
		}

		*script = r.WorkerScript
		return nil
	}
}

func testAccCheckCloudflareWorkerScriptDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_worker_script" {
			continue
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		params := getRequestParamsFromResource(rs)
		r, _ := client.DownloadWorker(context.Background(), &params)

		if r.Script != "" {
			return fmt.Errorf("Worker script with id %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
