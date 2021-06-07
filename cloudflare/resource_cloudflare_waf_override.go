package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceCloudflareWAFOverride() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareWAFOverrideCreate,
		Read:   resourceCloudflareWAFOverrideRead,
		Update: resourceCloudflareWAFOverrideUpdate,
		Delete: resourceCloudflareWAFOverrideDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareWAFOverrideImport,
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"urls": {
				Required: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"rules": {
				Optional: true,
				Type:     schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"paused": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(-1000000000, 1000000000),
			},
			"groups": {
				Optional: true,
				Type:     schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"rewrite_action": {
				Optional: true,
				Type:     schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"override_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCloudflareWAFOverrideRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	override, err := client.WAFOverride(context.Background(), zoneID, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "wafuriconfig.api.not_found") {
			log.Printf("[INFO] WAF override %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("failed to find WAF override %s: %s", d.Id(), err)
	}

	d.Set("zone_id", zoneID)
	d.Set("urls", override.URLs)
	d.Set("paused", override.Paused)
	d.Set("description", override.Description)
	d.Set("priority", override.Priority)

	if len(override.Rules) != 0 {
		d.Set("rules", override.Rules)
	}

	if len(override.Groups) != 0 {
		d.Set("groups", override.Groups)
	}

	if len(override.RewriteAction) != 0 {
		d.Set("rewrite_action", override.RewriteAction)
	}

	d.Set("override_id", override.ID)

	return nil
}

func resourceCloudflareWAFOverrideCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	newOverride, _ := buildWAFOverride(d)

	override, err := client.CreateWAFOverride(context.Background(), zoneID, newOverride)
	if err != nil {
		return fmt.Errorf("failed to create WAF override: %s", err)
	}

	d.SetId(override.ID)

	return resourceCloudflareWAFOverrideRead(d, meta)
}

func resourceCloudflareWAFOverrideUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	overrideID := d.Get("override_id").(string)
	updatedOverride, _ := buildWAFOverride(d)

	_, err := client.UpdateWAFOverride(context.Background(), zoneID, overrideID, updatedOverride)
	if err != nil {
		return fmt.Errorf("failed to update WAF override: %s", err)
	}

	return resourceCloudflareWAFOverrideRead(d, meta)
}

func resourceCloudflareWAFOverrideDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	overrideID := d.Get("override_id").(string)
	zoneID := d.Get("zone_id").(string)

	err := client.DeleteWAFOverride(context.Background(), zoneID, overrideID)
	if err != nil {
		return fmt.Errorf("failed to delete WAF override ID %s: %s", overrideID, err)
	}

	return resourceCloudflareWAFOverrideRead(d, meta)
}

func resourceCloudflareWAFOverrideImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 2)

	if len(idAttr) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/WAFOverrideID\"", d.Id())
	}

	zoneID, WAFOverrideID := idAttr[0], idAttr[1]

	log.Printf("[DEBUG] Importing WAF override: id %s for zone %s", WAFOverrideID, zoneID)

	d.Set("zone_id", zoneID)
	d.Set("override_id", WAFOverrideID)
	d.SetId(WAFOverrideID)

	resourceCloudflareWAFOverrideRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

// buildWAFOverride centralises the creation of a WAFOverride struct which can
// be reused between Create and Update methods to ensure consistent building of
// the values.
func buildWAFOverride(d *schema.ResourceData) (cloudflare.WAFOverride, error) {
	builtOverride := cloudflare.WAFOverride{}

	urls := d.Get("urls").([]interface{})
	for _, url := range urls {
		builtOverride.URLs = append(builtOverride.URLs, url.(string))
	}

	if rules, ok := d.GetOk("rules"); ok {
		rulesMap := make(map[string]string)
		for ruleID, state := range rules.(map[string]interface{}) {
			rulesMap[ruleID] = state.(string)
		}
		builtOverride.Rules = rulesMap
	}

	if pausedValue, ok := d.GetOk("paused"); ok {
		builtOverride.Paused = pausedValue.(bool)
	}

	if descriptionValue, ok := d.GetOk("description"); ok {
		builtOverride.Description = descriptionValue.(string)
	}

	if priorityValue, ok := d.GetOk("priority"); ok {
		builtOverride.Priority = priorityValue.(int)
	}

	if groupsValue, ok := d.GetOk("groups"); ok {
		groupsMap := make(map[string]string)
		for groupID, state := range groupsValue.(map[string]interface{}) {
			groupsMap[groupID] = state.(string)
		}
		builtOverride.Groups = groupsMap
	}

	if rewriteActionValue, ok := d.GetOk("rewrite_action"); ok {
		rewriteActions := make(map[string]string)
		for rewriteOriginal, rewriteWant := range rewriteActionValue.(map[string]interface{}) {
			rewriteActions[rewriteOriginal] = rewriteWant.(string)
		}
		builtOverride.RewriteAction = rewriteActions
	}

	return builtOverride, nil
}
