package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceCloudflareCustomHostnameFallbackOrigin() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareCustomHostnameFallbackOriginCreate,
		Read:   resourceCloudflareCustomHostnameFallbackOriginRead,
		Update: resourceCloudflareCustomHostnameFallbackOriginUpdate,
		Delete: resourceCloudflareCustomHostnameFallbackOriginDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareCustomHostnameFallbackOriginImport,
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"origin": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCloudflareCustomHostnameFallbackOriginRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	customHostnameFallbackOrigin, err := client.CustomHostnameFallbackOrigin(context.Background(), zoneID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error reading custom hostname fallback origin %q", zoneID))
	}

	d.Set("origin", customHostnameFallbackOrigin.Origin)
	d.Set("status", customHostnameFallbackOrigin.Status)

	return nil
}

func resourceCloudflareCustomHostnameFallbackOriginDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	err := client.DeleteCustomHostnameFallbackOrigin(context.Background(), zoneID)
	if err != nil {
		return errors.Wrap(err, "failed to delete custom hostname fallback origin")
	}

	return nil
}

func resourceCloudflareCustomHostnameFallbackOriginCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	origin := d.Get("origin").(string)

	fallbackOrigin := cloudflare.CustomHostnameFallbackOrigin{
		Origin: origin,
	}

	return resource.Retry(d.Timeout(schema.TimeoutDefault), func() *resource.RetryError {
		_, err := client.UpdateCustomHostnameFallbackOrigin(context.Background(), zoneID, fallbackOrigin)
		if err != nil {
			if err.(*cloudflare.APIRequestError).InternalErrorCodeIs(1414) {
				return resource.RetryableError(fmt.Errorf("expected custom hostname resource to be ready for modification but is still pending"))
			}
			return resource.NonRetryableError(errors.Wrap(err, "failed to create custom hostname fallback origin"))
		}

		fallbackHostname, err := client.CustomHostnameFallbackOrigin(context.Background(), zoneID)

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("failed to fetch custom hostname: %s", err))
		}

		// Address an eventual consistency issue where deleting a fallback hostname
		// and then adding it _may_ cause some issues. It is possible that the status does
		// move into the active state during the retry period.
		if fallbackHostname.Status != "pending_deployment" && fallbackHostname.Status != "active" {
			return resource.RetryableError(fmt.Errorf("expected custom hostname fallback to be created but was %s", fallbackHostname.Status))
		}

		id := stringChecksum(fmt.Sprintf("%s/custom_hostnames_fallback_origin", zoneID))
		d.SetId(id)

		return resource.NonRetryableError(resourceCloudflareCustomHostnameFallbackOriginRead(d, meta))
	})

}

func resourceCloudflareCustomHostnameFallbackOriginUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	origin := d.Get("origin").(string)

	fallbackOrigin := cloudflare.CustomHostnameFallbackOrigin{
		Origin: origin,
	}

	return resource.Retry(d.Timeout(schema.TimeoutDefault), func() *resource.RetryError {
		_, err := client.UpdateCustomHostnameFallbackOrigin(context.Background(), zoneID, fallbackOrigin)
		if err != nil {
			if err.(*cloudflare.APIRequestError).InternalErrorCodeIs(1414) {
				return resource.RetryableError(fmt.Errorf("expected custom hostname resource to be ready for modification but is still pending"))
			}
			return resource.NonRetryableError(errors.Wrap(err, "failed to update custom hostname fallback origin"))
		}

		return resource.NonRetryableError(resourceCloudflareCustomHostnameFallbackOriginRead(d, meta))
	})
}

func resourceCloudflareCustomHostnameFallbackOriginImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 2)

	if len(idAttr) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/origin\"", d.Id())
	}

	zoneID, origin := idAttr[0], idAttr[1]

	log.Printf("[DEBUG] Importing Cloudflare Custom Hostname Fallback Origin: origin %s for zone %s", origin, zoneID)

	d.Set("zone_id", zoneID)
	d.Set("origin", origin)

	id := stringChecksum(fmt.Sprintf("%s/custom_hostnames_fallback_origin", zoneID))
	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}
