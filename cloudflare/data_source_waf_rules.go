package cloudflare

import (
	"context"
	"fmt"
	"log"
	"regexp"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudflareWAFRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCloudflareWAFRulesRead,

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"package_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"group_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"priority": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"group_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"group_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"package_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"allowed_modes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudflareWAFRulesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	// Prepare the filters to be applied to the search
	filter, err := expandFilterWAFRules(d.Get("filter"))
	if err != nil {
		return err
	}

	// If no package ID is given, we will consider all for the zone
	packageID := d.Get("package_id").(string)
	var pkgList []cloudflare.WAFPackage
	if packageID == "" {
		var err error
		log.Printf("[DEBUG] Reading WAF Packages")
		pkgList, err = client.ListWAFPackages(context.Background(), zoneID)
		if err != nil {
			return err
		}
	} else {
		pkgList = append(pkgList, cloudflare.WAFPackage{ID: packageID})
	}

	log.Printf("[DEBUG] Reading WAF Rules")
	ruleIds := make([]string, 0)
	ruleDetails := make([]interface{}, 0)
	for _, pkg := range pkgList {
		ruleList, err := client.ListWAFRules(context.Background(), zoneID, pkg.ID)
		if err != nil {
			return err
		}

		foundGroup := false
		for _, rule := range ruleList {
			if filter.GroupID != "" {
				if filter.GroupID != rule.Group.ID {
					continue
				}

				// Allows to stop querying the API faster
				foundGroup = true
			}

			if filter.Description != nil && !filter.Description.Match([]byte(rule.Description)) {
				continue
			}

			if filter.Mode != "" && filter.Mode != rule.Mode {
				continue
			}

			ruleDetails = append(ruleDetails, map[string]interface{}{
				"id":            rule.ID,
				"description":   rule.Description,
				"priority":      rule.Priority,
				"mode":          rule.Mode,
				"group_id":      rule.Group.ID,
				"group_name":    rule.Group.Name,
				"package_id":    pkg.ID,
				"allowed_modes": rule.AllowedModes,
			})
			ruleIds = append(ruleIds, rule.ID)
		}

		if foundGroup {
			// We can stop looking further as a group is only part of a unique
			// package, meaning that if we found the group, no need to go look
			// at other packages
			break
		}
	}

	err = d.Set("rules", ruleDetails)
	if err != nil {
		return fmt.Errorf("Error setting WAF rules: %s", err)
	}

	d.SetId(stringListChecksum(ruleIds))
	return nil
}

func expandFilterWAFRules(d interface{}) (*searchFilterWAFRules, error) {
	cfg := d.([]interface{})
	filter := &searchFilterWAFRules{}
	if len(cfg) == 0 || cfg[0] == nil {
		return filter, nil
	}

	m := cfg[0].(map[string]interface{})
	description, ok := m["description"]
	if ok {
		match, err := regexp.Compile(description.(string))
		if err != nil {
			return nil, err
		}

		filter.Description = match
	}

	mode, ok := m["mode"]
	if ok {
		filter.Mode = mode.(string)
	}

	groupID, ok := m["group_id"]
	if ok {
		filter.GroupID = groupID.(string)
	}

	return filter, nil
}

type searchFilterWAFRules struct {
	Description *regexp.Regexp
	Mode        string
	GroupID     string
}
