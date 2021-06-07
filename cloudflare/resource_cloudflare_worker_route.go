package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceCloudflareWorkerRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareWorkerRouteCreate,
		Read:   resourceCloudflareWorkerRouteRead,
		Update: resourceCloudflareWorkerRouteUpdate,
		Delete: resourceCloudflareWorkerRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareWorkerRouteImport,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"pattern": {
				Type:     schema.TypeString,
				Required: true,
			},

			"script_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getRouteFromResource(d *schema.ResourceData) cloudflare.WorkerRoute {
	route := cloudflare.WorkerRoute{
		ID:      d.Id(),
		Pattern: d.Get("pattern").(string),
		Script:  d.Get("script_name").(string),
	}
	return route
}

func resourceCloudflareWorkerRouteCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	route := getRouteFromResource(d)
	zoneID := d.Get("zone_id").(string)

	log.Printf("[INFO] Creating Cloudflare Worker Route from struct: %+v", route)

	r, err := client.CreateWorkerRoute(context.Background(), zoneID, route)
	if err != nil {
		return errors.Wrap(err, "error creating worker route")
	}

	if r.ID == "" {
		return fmt.Errorf("failed to find id in Create response; resource was empty")
	}

	d.SetId(r.ID)

	log.Printf("[INFO] Cloudflare Worker Route ID: %s", d.Id())

	return nil
}

func resourceCloudflareWorkerRouteRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	routeID := d.Id()

	// There isn't a dedicated endpoint for retrieving a specific route, so we
	// list all routes and find the target route by comparing IDs
	resp, err := client.ListWorkerRoutes(context.Background(), zoneID)

	if err != nil {
		return errors.Wrap(err, "error reading worker routes")
	}

	var route cloudflare.WorkerRoute
	for _, r := range resp.Routes {
		if r.ID == routeID {
			route = r
			break
		}
	}

	// If the resource is deleted, we should set the ID to "" and not
	// return an error according to the terraform spec
	if route.ID == "" {
		d.SetId("")
		return nil
	}

	d.Set("pattern", route.Pattern)
	d.Set("script_name", route.Script)

	return nil
}

func resourceCloudflareWorkerRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	route := getRouteFromResource(d)

	log.Printf("[INFO] Updating Cloudflare Worker Route from struct: %+v", route)

	_, err := client.UpdateWorkerRoute(context.Background(), zoneID, route.ID, route)
	if err != nil {
		return errors.Wrap(err, "error updating worker route")
	}

	return nil
}

func resourceCloudflareWorkerRouteDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	route := getRouteFromResource(d)

	log.Printf("[INFO] Deleting Cloudflare Worker Route from zone %+v with id: %+v", zoneID, route.ID)

	_, err := client.DeleteWorkerRoute(context.Background(), zoneID, route.ID)
	if err != nil {
		return errors.Wrap(err, "error deleting worker route")
	}

	return nil
}

func resourceCloudflareWorkerRouteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// split the id so we can lookup
	idAttr := strings.SplitN(d.Id(), "/", 2)
	var zoneID string
	var routeID string
	if len(idAttr) == 2 {
		zoneID = idAttr[0]
		routeID = idAttr[1]
	} else {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/routeID\"", d.Id())
	}

	d.Set("zone_id", zoneID)
	d.SetId(routeID)

	resourceCloudflareWorkerRouteRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
