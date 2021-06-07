package cloudflare

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudflareLogpushOwnershipChallenge() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareLogpushOwnershipChallengeCreate,
		Update: resourceCloudflareLogpushOwnershipChallengeCreate,
		Read:   resourceCloudflareLogpushOwnershipChallengeNoop,
		Delete: resourceCloudflareLogpushOwnershipChallengeNoop,

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"destination_conf": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"ownership_challenge_filename": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCloudflareLogpushOwnershipChallengeNoop(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCloudflareLogpushOwnershipChallengeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	destinationConf := d.Get("destination_conf").(string)

	challenge, err := client.GetLogpushOwnershipChallenge(context.Background(), zoneID, destinationConf)
	if err != nil {
		return fmt.Errorf("error requesting ownership challenge: %v", err)
	}

	// The ownership challenge doesn't have a unique identifier so we generate it
	// here from the filename which will be unique.
	d.SetId(stringChecksum(challenge.Filename))
	d.Set("ownership_challenge_filename", challenge.Filename)

	log.Printf("[INFO] Created Cloudflare Logpush Ownership Challenge: %s", d.Id())

	return nil
}
