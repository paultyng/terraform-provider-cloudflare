package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceCloudflareCustomPages() *schema.Resource {
	return &schema.Resource{
		// Pointing the `Create` at the `Update` method here is intentional.
		// Custom pages don't really get "created" as they are always
		// present in Cloudflare. We just update and toggle the settings to
		// be customised.
		Create: resourceCloudflareCustomPagesUpdate,
		Read:   resourceCloudflareCustomPagesRead,
		Update: resourceCloudflareCustomPagesUpdate,
		Delete: resourceCloudflareCustomPagesDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareCustomPagesImport,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"account_id"},
			},
			"account_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zone_id"},
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"basic_challenge",
					"waf_challenge",
					"waf_block",
					"ratelimit_block",
					"country_challenge",
					"ip_block",
					"under_attack",
					"500_errors",
					"1000_errors",
					"always_online",
				}, true),
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"default", "customized"}, true),
			},
		},
	}
}

func resourceCloudflareCustomPagesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	accountID := d.Get("account_id").(string)
	pageType := d.Get("type").(string)

	if accountID == "" && zoneID == "" {
		return fmt.Errorf("either `account_id` or `zone_id` must be set")
	}

	var (
		pageOptions cloudflare.CustomPageOptions
		identifier  string
	)

	if accountID != "" {
		pageOptions = cloudflare.CustomPageOptions{AccountID: accountID}
		identifier = accountID
	} else {
		pageOptions = cloudflare.CustomPageOptions{ZoneID: zoneID}
		identifier = zoneID
	}

	page, err := client.CustomPage(context.Background(), &pageOptions, pageType)
	if err != nil {
		return errors.New(err.Error())
	}

	// If the `page.State` comes back as "default", it's safe to assume we
	// don't need to keep the ID managed anymore as it will be relying on
	// Cloudflare's default pages.
	if page.State == "default" {
		log.Printf("[INFO] removing custom page configuration for '%s' as it is marked as being in the default state", pageType)
		d.SetId("")
		return nil
	}

	checksum := stringChecksum(fmt.Sprintf("%s/%s", identifier, page.ID))
	d.SetId(checksum)

	d.Set("state", page.State)
	d.Set("url", page.URL)
	d.Set("type", page.ID)

	return nil
}

func resourceCloudflareCustomPagesUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	accountID := d.Get("account_id").(string)
	zoneID := d.Get("zone_id").(string)

	var pageOptions cloudflare.CustomPageOptions
	if accountID != "" {
		pageOptions = cloudflare.CustomPageOptions{AccountID: accountID}
	} else {
		pageOptions = cloudflare.CustomPageOptions{ZoneID: zoneID}
	}

	pageType := d.Get("type").(string)
	customPageParameters := cloudflare.CustomPageParameters{
		URL:   d.Get("url").(string),
		State: "customized",
	}
	_, err := client.UpdateCustomPage(context.Background(), &pageOptions, pageType, customPageParameters)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to update '%s' custom page", pageType))
	}

	return resourceCloudflareCustomPagesRead(d, meta)
}

func resourceCloudflareCustomPagesDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	accountID := d.Get("account_id").(string)
	zoneID := d.Get("zone_id").(string)

	var pageOptions cloudflare.CustomPageOptions
	if accountID != "" {
		pageOptions = cloudflare.CustomPageOptions{AccountID: accountID}
	} else {
		pageOptions = cloudflare.CustomPageOptions{ZoneID: zoneID}
	}

	pageType := d.Get("type").(string)
	customPageParameters := cloudflare.CustomPageParameters{
		URL:   nil,
		State: "default",
	}
	_, err := client.UpdateCustomPage(context.Background(), &pageOptions, pageType, customPageParameters)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to update '%s' custom page", pageType))
	}

	return resourceCloudflareCustomPagesRead(d, meta)
}

func resourceCloudflareCustomPagesImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 3)
	if len(attributes) != 3 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"requestType/ID/pageType\"", d.Id())
	}
	requestType, identifier, pageType := attributes[0], attributes[1], attributes[2]

	d.Set("type", pageType)

	if requestType == "account" {
		d.Set("account_id", identifier)
	} else {
		d.Set("zone_id", identifier)
	}

	checksum := stringChecksum(fmt.Sprintf("%s/%s", identifier, pageType))
	d.SetId(checksum)

	resourceCloudflareCustomPagesRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
