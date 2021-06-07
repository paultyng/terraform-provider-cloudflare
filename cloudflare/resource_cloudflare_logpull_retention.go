package cloudflare

import (
	"context"
	"fmt"
	"log"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudflareLogpullRetention() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareLogpullRetentionSet,
		Read:   resourceCloudflareLogpullRetentionRead,
		Update: resourceCloudflareLogpullRetentionSet,
		Delete: resourceCloudflareLogpullRetentionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareLogpullRetentionImport,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceCloudflareLogpullRetentionSet(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	status := d.Get("enabled").(bool)

	_, err := client.SetLogpullRetentionFlag(context.Background(), zoneID, status)
	if err != nil {
		return fmt.Errorf("error setting Logpull Retention for zone ID %q: %s", zoneID, err)
	}

	d.SetId(stringChecksum("logpull-retention/" + zoneID))

	return resourceCloudflareLogpullRetentionRead(d, meta)
}

func resourceCloudflareLogpullRetentionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	logpullConf, err := client.GetLogpullRetentionFlag(context.Background(), zoneID)
	if err != nil {
		return fmt.Errorf("error getting Logpull Retention for zone ID %q: %s", zoneID, err)
	}

	d.Set("enabled", logpullConf.Flag)

	return nil
}

func resourceCloudflareLogpullRetentionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	_, err := client.SetLogpullRetentionFlag(context.Background(), zoneID, false)
	if err != nil {
		return fmt.Errorf("error setting Logpull Retention for zone ID %q: %s", zoneID, err)
	}

	d.SetId("")

	return nil
}

func resourceCloudflareLogpullRetentionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	zoneID := d.Id()

	log.Printf("[DEBUG] Importing Cloudflare Logpull Retention option for zone ID: %s", zoneID)

	d.Set("zone_id", zoneID)
	d.SetId(stringChecksum("logpull-retention/" + zoneID))

	resourceCloudflareLogpullRetentionRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
