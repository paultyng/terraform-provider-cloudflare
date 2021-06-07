package cloudflare

import (
	"context"
	"fmt"
	"html"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceCloudflareFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareFilterCreate,
		Read:   resourceCloudflareFilterRead,
		Update: resourceCloudflareFilterUpdate,
		Delete: resourceCloudflareFilterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareFilterImport,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"paused": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"expression": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(new) == old
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 500),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if html.UnescapeString(old) == html.UnescapeString(new) {
						return true
					}
					return false
				},
			},
			"ref": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
		},
	}
}

func resourceCloudflareFilterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	var err error

	var newFilter cloudflare.Filter

	if paused, ok := d.GetOk("paused"); ok {
		newFilter.Paused = paused.(bool)
	}

	if description, ok := d.GetOk("description"); ok {
		newFilter.Description = description.(string)
	}

	if expression, ok := d.GetOk("expression"); ok {
		newFilter.Expression = expression.(string)
	}

	if ref, ok := d.GetOk("ref"); ok {
		newFilter.Ref = ref.(string)
	}

	log.Printf("[DEBUG] Creating Cloudflare Filter from struct: %+v", newFilter)

	r, err := client.CreateFilters(context.Background(), zoneID, []cloudflare.Filter{newFilter})

	if err != nil {
		return fmt.Errorf("error creating Filter for zone %q: %s", zoneID, err)
	}

	if len(r) == 0 {
		return fmt.Errorf("failed to find id in Create response; resource was empty")
	}

	d.SetId(r[0].ID)

	log.Printf("[INFO] Cloudflare Filter ID: %s", d.Id())

	return resourceCloudflareFilterRead(d, meta)
}

func resourceCloudflareFilterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	log.Printf("[DEBUG] Getting a Filter record for zone %q, id %s", zoneID, d.Id())
	filter, err := client.Filter(context.Background(), zoneID, d.Id())

	log.Printf("[DEBUG] filter: %#v", filter)
	log.Printf("[DEBUG] filter error: %#v", err)

	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Filter %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error finding Filter %q: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Cloudflare Filter read configuration: %#v", filter)

	d.Set("paused", filter.Paused)
	d.Set("description", filter.Description)
	d.Set("expression", filter.Expression)
	d.Set("ref", filter.Ref)

	return nil
}

func resourceCloudflareFilterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	var newFilter cloudflare.Filter
	newFilter.ID = d.Id()

	if paused, ok := d.GetOk("paused"); ok {
		newFilter.Paused = paused.(bool)
	}

	if description, ok := d.GetOk("description"); ok {
		newFilter.Description = description.(string)
	}

	if expression, ok := d.GetOk("expression"); ok {
		newFilter.Expression = expression.(string)
	}

	if ref, ok := d.GetOk("ref"); ok {
		newFilter.Ref = ref.(string)
	}

	log.Printf("[DEBUG] Updating Cloudflare Filter from struct: %+v", newFilter)

	r, err := client.UpdateFilter(context.Background(), zoneID, newFilter)

	if err != nil {
		return fmt.Errorf("error updating Filter for zone %q: %s", zoneID, err)
	}

	if r.ID == "" {
		return fmt.Errorf("failed to find id in Update response; resource was empty")
	}

	return resourceCloudflareFilterRead(d, meta)
}

func resourceCloudflareFilterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	log.Printf("[INFO] Deleting Cloudflare Filter: id %s for zone %s", d.Id(), zoneID)

	err := client.DeleteFilter(context.Background(), zoneID, d.Id())

	if err != nil {
		return fmt.Errorf("Error deleting Cloudflare Filter: %s", err)
	}

	return nil
}

func resourceCloudflareFilterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// split the id so we can lookup
	idAttr := strings.SplitN(d.Id(), "/", 2)

	if len(idAttr) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/filterID\"", d.Id())
	}

	zoneID, filterID := idAttr[0], idAttr[1]

	log.Printf("[DEBUG] Importing Cloudflare Filter: id %s for zone %s", filterID, zoneID)

	d.Set("zone_id", zoneID)
	d.SetId(filterID)

	resourceCloudflareFilterRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
