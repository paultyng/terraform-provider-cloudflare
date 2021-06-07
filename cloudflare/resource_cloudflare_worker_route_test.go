package cloudflare

import (
	"context"
	"fmt"
	"os"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	defaultScriptContent = "addEventListener('fetch', event => {event.respondWith(fetch(event.request))})"
)

func TestAccCloudflareWorkerRoute_MultiScriptEnt(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the Workers
	// service does not yet support the API tokens and it results in
	// misleading state error messages.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	var route cloudflare.WorkerRoute
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	routeRnd := generateRandomResourceName()
	routeName := "cloudflare_worker_route." + routeRnd
	pattern1 := fmt.Sprintf("%s/%s", zoneName, generateRandomResourceName())
	pattern2 := fmt.Sprintf("%s/%s", zoneName, generateRandomResourceName())

	// We also create a script in order to test routes since routes
	// need to point to a script
	scriptRnd := generateRandomResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareWorkerRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWorkerRouteConfigMultiScriptInitial(zoneID, routeRnd, scriptRnd, pattern1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkerRouteExists(routeName, &route),
					resource.TestCheckResourceAttr(routeName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(routeName, "pattern", pattern1),
					resource.TestCheckResourceAttr(routeName, "script_name", scriptRnd),
				),
			},
			{
				Config: testAccCheckCloudflareWorkerRouteConfigMultiScriptUpdate(zoneID, routeRnd, scriptRnd, pattern2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkerRouteExists(routeName, &route),
					resource.TestCheckResourceAttr(routeName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(routeName, "pattern", pattern2),
					resource.TestCheckResourceAttr(routeName, "script_name", ""),
				),
			},
		},
	})
}

func testAccCheckCloudflareWorkerRouteConfigMultiScriptInitial(zoneID, routeRnd, scriptRnd, pattern string) string {
	return fmt.Sprintf(`
resource "cloudflare_worker_route" "%[2]s" {
  zone_id = "%[1]s"
  pattern = "%[4]s"
  script_name = "${cloudflare_worker_script.%[3]s.name}"
}

resource "cloudflare_worker_script" "%[3]s" {
  name = "%[3]s"
  content = "%[5]s"
}`, zoneID, routeRnd, scriptRnd, pattern, defaultScriptContent)
}

func testAccCheckCloudflareWorkerRouteConfigMultiScriptUpdate(zoneID, routeRnd, scriptRnd, pattern string) string {
	return fmt.Sprintf(`
resource "cloudflare_worker_route" "%[2]s" {
  zone_id = "%[1]s"
  pattern = "%[4]s"
}

resource "cloudflare_worker_script" "%[3]s" {
  name = "%[3]s"
  content = "%[5]s"
}`, zoneID, routeRnd, scriptRnd, pattern, defaultScriptContent)
}

func TestAccCloudflareWorkerRoute_MultiScriptDisabledRoute(t *testing.T) {
	// Temporarily unset CLOUDFLARE_API_TOKEN if it is set as the Workers
	// service does not yet support the API tokens and it results in
	// misleading state error messages.
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		defer func(apiToken string) {
			os.Setenv("CLOUDFLARE_API_TOKEN", apiToken)
		}(os.Getenv("CLOUDFLARE_API_TOKEN"))
		os.Setenv("CLOUDFLARE_API_TOKEN", "")
	}

	var route cloudflare.WorkerRoute
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	routeRnd := generateRandomResourceName()
	routeName := "cloudflare_worker_route." + routeRnd
	pattern := fmt.Sprintf("%s/%s", zoneName, generateRandomResourceName())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAccount(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareWorkerRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWorkerRouteConfigMultiScriptDisabledRoute(zoneID, routeRnd, pattern),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkerRouteExists(routeName, &route),
					resource.TestCheckResourceAttr(routeName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(routeName, "pattern", pattern),
					resource.TestCheckNoResourceAttr(routeName, "script_name"),
				),
			},
		},
	})
}

func testAccCheckCloudflareWorkerRouteConfigMultiScriptDisabledRoute(zoneID, routeRnd, pattern string) string {
	return fmt.Sprintf(`
resource "cloudflare_worker_route" "%[2]s" {
  zone_id = "%[1]s"
  pattern = "%[3]s"
}`, zoneID, routeRnd, pattern)
}

func getRouteFromApi(zoneID, routeId string) (cloudflare.WorkerRoute, error) {
	if zoneID == "" {
		return cloudflare.WorkerRoute{}, fmt.Errorf("zoneID is required to get a route")
	}
	if routeId == "" {
		return cloudflare.WorkerRoute{}, fmt.Errorf("routeId is required to get a route")
	}

	client := testAccProvider.Meta().(*cloudflare.API)
	resp, err := client.ListWorkerRoutes(context.Background(), zoneID)
	if err != nil {
		return cloudflare.WorkerRoute{}, err
	}

	var route cloudflare.WorkerRoute
	for _, r := range resp.Routes {
		if r.ID == routeId {
			route = r
			break
		}
	}

	return route, nil
}

func testAccCheckCloudflareWorkerRouteExists(n string, route *cloudflare.WorkerRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Worker Route ID is set")
		}

		zoneID := rs.Primary.Attributes["zone_id"]
		routeId := rs.Primary.ID
		foundRoute, err := getRouteFromApi(zoneID, routeId)
		if err != nil {
			return err
		}

		if foundRoute.ID != routeId {
			return fmt.Errorf("Worker route with id %s not found", routeId)
		}

		*route = foundRoute
		return nil
	}
}

func testAccCheckCloudflareWorkerRouteDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_worker_route" {
			continue
		}

		zoneID := rs.Primary.Attributes["zone_id"]
		routeId := rs.Primary.ID
		route, err := getRouteFromApi(zoneID, routeId)

		if err != nil {
			return err
		}

		if route.ID != "" {
			return fmt.Errorf("Worker route with id %s still exists", route.ID)
		}

	}

	return nil
}
