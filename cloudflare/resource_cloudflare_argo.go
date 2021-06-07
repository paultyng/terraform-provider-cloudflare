package cloudflare

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceCloudflareArgo() *schema.Resource {
	return &schema.Resource{
		// Pointing `Create` to the `Update `method is intentional. Argo
		// settings are always present, it's just whether or not the value
		// is "on" or "off".
		Create: resourceCloudflareArgoUpdate,
		Read:   resourceCloudflareArgoRead,
		Update: resourceCloudflareArgoUpdate,
		Delete: resourceCloudflareArgoDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareArgoImport,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tiered_caching": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
				Optional:     true,
			},
			"smart_routing": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
				Optional:     true,
			},
		},
	}
}

func resourceCloudflareArgoRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	log.Printf("[DEBUG] zone ID: %s", zoneID)

	checksum := stringChecksum(fmt.Sprintf("%s/argo", zoneID))
	d.SetId(checksum)
	d.Set("zone_id", zoneID)

	if _, ok := d.GetOk("tiered_caching"); ok {
		tieredCaching, err := client.ArgoTieredCaching(context.Background(), zoneID)
		if err != nil {
			return errors.Wrap(err, "failed to get tiered caching setting")
		}

		d.Set("tiered_caching", tieredCaching.Value)
	}

	if _, ok := d.GetOk("smart_routing"); ok {
		smartRouting, err := client.ArgoSmartRouting(context.Background(), zoneID)
		if err != nil {
			return errors.Wrap(err, "failed to get smart routing setting")
		}

		d.Set("smart_routing", smartRouting.Value)
	}

	return nil
}

func resourceCloudflareArgoUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	tieredCaching := d.Get("tiered_caching").(string)
	smartRouting := d.Get("smart_routing").(string)

	if smartRouting != "" {
		argoSmartRouting, err := client.UpdateArgoSmartRouting(context.Background(), zoneID, smartRouting)
		if err != nil {
			return errors.Wrap(err, "failed to update smart routing setting")
		}
		log.Printf("[DEBUG] Argo Smart Routing set to: %s", argoSmartRouting.Value)
	}

	if tieredCaching != "" {
		argoTieredCaching, err := client.UpdateArgoTieredCaching(context.Background(), zoneID, tieredCaching)
		if err != nil {
			return errors.Wrap(err, "failed to update tiered caching setting")
		}
		log.Printf("[DEBUG] Argo Tiered Caching set to: %s", argoTieredCaching.Value)
	}

	return resourceCloudflareArgoRead(d, meta)
}

func resourceCloudflareArgoDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	log.Printf("[DEBUG] Resetting Argo values to 'off'")

	_, smartRoutingErr := client.UpdateArgoSmartRouting(context.Background(), zoneID, "off")
	if smartRoutingErr != nil {
		return errors.Wrap(smartRoutingErr, "failed to update smart routing setting")
	}

	_, tieredCachingErr := client.UpdateArgoTieredCaching(context.Background(), zoneID, "off")
	if tieredCachingErr != nil {
		return errors.Wrap(tieredCachingErr, "failed to update tiered caching setting")
	}

	return nil
}

func resourceCloudflareArgoImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	zoneID := d.Id()

	id := stringChecksum(fmt.Sprintf("%s/argo", zoneID))
	d.SetId(id)
	d.Set("zone_id", zoneID)

	return []*schema.ResourceData{d}, nil
}
