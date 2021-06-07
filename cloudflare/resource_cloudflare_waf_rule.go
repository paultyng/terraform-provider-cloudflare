package cloudflare

import (
	"context"
	"fmt"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudflareWAFRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareWAFRuleCreate,
		Read:   resourceCloudflareWAFRuleRead,
		Update: resourceCloudflareWAFRuleUpdate,
		Delete: resourceCloudflareWAFRuleDelete,

		Importer: &schema.ResourceImporter{
			State: resourceCloudflareWAFRuleImport,
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"rule_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"package_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"mode": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCloudflareWAFRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	ruleID := d.Get("rule_id").(string)
	zoneID := d.Get("zone_id").(string)
	packageID := d.Get("package_id").(string)

	rule, err := client.WAFRule(context.Background(), zoneID, packageID, ruleID)
	if err != nil {
		if err.(*cloudflare.APIRequestError).InternalErrorCodeIs(1002) || err.(*cloudflare.APIRequestError).InternalErrorCodeIs(1004) {
			d.SetId("")
			return nil
		}

		return err
	}

	// Only need to set mode as that is the only attribute that could have changed
	d.Set("mode", rule.Mode)
	d.Set("group_id", rule.Group.ID)
	d.SetId(rule.ID)

	return nil
}

func resourceCloudflareWAFRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	ruleID := d.Get("rule_id").(string)
	zoneID := d.Get("zone_id").(string)
	packageID := d.Get("package_id").(string)
	mode := d.Get("mode").(string)

	// If no package ID is given try to resolve it
	var pkgList []cloudflare.WAFPackage
	if packageID == "" {
		var err error
		pkgList, err = client.ListWAFPackages(context.Background(), zoneID)
		if err != nil {
			return err
		}
	} else {
		pkgList = append(pkgList, cloudflare.WAFPackage{ID: packageID})
	}

	for _, pkg := range pkgList {
		var err error
		var rule cloudflare.WAFRule

		rule, err = client.WAFRule(context.Background(), zoneID, pkg.ID, ruleID)
		if err != nil {
			continue
		}

		d.Set("rule_id", rule.ID)
		d.Set("zone_id", zoneID)
		d.Set("group_id", rule.Group.ID)
		d.Set("package_id", pkg.ID)

		if rule.Mode != mode {
			err = resourceCloudflareWAFRuleUpdate(d, meta)
			if err != nil {
				d.SetId("")
				return err
			}
		}

		return resourceCloudflareWAFRuleRead(d, meta)
	}

	return fmt.Errorf("Unable to find WAF Rule %s", ruleID)
}

func resourceCloudflareWAFRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	ruleID := d.Get("rule_id").(string)
	zoneID := d.Get("zone_id").(string)
	packageID := d.Get("package_id").(string)

	rule, err := client.WAFRule(context.Background(), zoneID, packageID, ruleID)
	if err != nil {
		return err
	}

	// Find the default mode to be used
	defaultMode := "default"
	if !contains(rule.AllowedModes, defaultMode) {
		defaultMode = "on"
	}

	// Can't delete WAF Rule so instead reset it to default
	if rule.Mode != defaultMode {
		_, err = client.UpdateWAFRule(context.Background(), zoneID, packageID, ruleID, defaultMode)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceCloudflareWAFRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	ruleID := d.Get("rule_id").(string)
	zoneID := d.Get("zone_id").(string)
	mode := d.Get("mode").(string)
	packageID := d.Get("package_id").(string)

	// We can only update the mode of a WAF Rule
	_, err := client.UpdateWAFRule(context.Background(), zoneID, packageID, ruleID, mode)
	if err != nil {
		return err
	}

	return resourceCloudflareWAFRuleRead(d, meta)
}

func resourceCloudflareWAFRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*cloudflare.API)

	// split the id so we can lookup
	idAttr := strings.SplitN(d.Id(), "/", 2)
	var zoneID string
	var WAFID string
	if len(idAttr) == 2 {
		zoneID = idAttr[0]
		WAFID = idAttr[1]
	} else {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/WAFID\" for import", d.Id())
	}

	packs, err := client.ListWAFPackages(context.Background(), zoneID)
	if err != nil {
		return nil, fmt.Errorf("error listing WAF packages: %s", err)
	}

	for _, p := range packs {
		rule, err := client.WAFRule(context.Background(), zoneID, p.ID, WAFID)
		if err == nil {
			d.Set("rule_id", rule.ID)
			d.Set("zone_id", zoneID)
			d.Set("package_id", rule.PackageID)
			d.Set("group_id", rule.Group.ID)
			d.Set("mode", rule.Mode)

			// The ID is known by the user in advance
			d.SetId(WAFID)
		}
	}

	if d.Id() != WAFID {
		return nil, fmt.Errorf("Unable to find WAF Rule %s", WAFID)
	}

	resourceCloudflareWAFRuleRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
