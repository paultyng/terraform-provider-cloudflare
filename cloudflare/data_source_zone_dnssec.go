package cloudflare

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudflareZoneDNSSEC() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCloudflareZoneDNSSECRead,

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flags": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"digest_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"digest_algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"digest": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ds": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_tag": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCloudflareZoneDNSSECRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	zoneID := d.Get("zone_id").(string)

	log.Printf("[DEBUG] Reading Zone DNSSEC %s", zoneID)

	dnssec, err := client.ZoneDNSSECSetting(context.Background(), zoneID)
	if err != nil {
		return fmt.Errorf("Error finding Zone DNSSEC %q: %s", zoneID, err)
	}

	d.Set("zone_id", zoneID)
	d.Set("status", dnssec.Status)
	d.Set("flags", dnssec.Flags)
	d.Set("algorithm", dnssec.Algorithm)
	d.Set("key_type", dnssec.KeyType)
	d.Set("digest_type", dnssec.DigestType)
	d.Set("digest_algorithm", dnssec.DigestAlgorithm)
	d.Set("digest", dnssec.Digest)
	d.Set("ds", dnssec.DS)
	d.Set("key_tag", dnssec.KeyTag)
	d.Set("public_key", dnssec.PublicKey)

	d.SetId(stringChecksum(dnssec.ModifiedOn.String()))

	return nil
}
