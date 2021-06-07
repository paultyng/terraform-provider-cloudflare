package cloudflare

import (
	"context"
	"fmt"
	"log"
	"regexp"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceCloudflareWAFGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCloudflareWAFGroupsRead,

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
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
						},
					},
				},
			},

			"groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"rules_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"modified_rules_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"package_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudflareWAFGroupsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	// Prepare the filters to be applied to the search
	filter, err := expandFilterWAFGroups(d.Get("filter"))
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

	log.Printf("[DEBUG] Reading WAF Groups")
	groupIds := make([]string, 0)
	groupDetails := make([]interface{}, 0)
	for _, pkg := range pkgList {
		groupList, err := client.ListWAFGroups(context.Background(), zoneID, pkg.ID)
		if err != nil {
			return err
		}

		for _, group := range groupList {
			if filter.Name != nil && !filter.Name.Match([]byte(group.Name)) {
				continue
			}

			if filter.Mode != "" && filter.Mode != group.Mode {
				continue
			}

			groupDetails = append(groupDetails, map[string]interface{}{
				"id":                   group.ID,
				"name":                 group.Name,
				"description":          group.Description,
				"mode":                 group.Mode,
				"rules_count":          group.RulesCount,
				"modified_rules_count": group.ModifiedRulesCount,
				"package_id":           pkg.ID,
			})
			groupIds = append(groupIds, group.ID)
		}
	}

	err = d.Set("groups", groupDetails)
	if err != nil {
		return fmt.Errorf("Error setting WAF groups: %s", err)
	}

	d.SetId(stringListChecksum(groupIds))
	return nil
}

func expandFilterWAFGroups(d interface{}) (*searchFilterWAFGroups, error) {
	cfg := d.([]interface{})
	filter := &searchFilterWAFGroups{}
	if len(cfg) == 0 || cfg[0] == nil {
		return filter, nil
	}

	m := cfg[0].(map[string]interface{})
	name, ok := m["name"]
	if ok {
		match, err := regexp.Compile(name.(string))
		if err != nil {
			return nil, err
		}

		filter.Name = match
	}

	mode, ok := m["mode"]
	if ok {
		filter.Mode = mode.(string)
	}

	return filter, nil
}

type searchFilterWAFGroups struct {
	Name *regexp.Regexp
	Mode string
}
