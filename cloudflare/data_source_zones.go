package cloudflare

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceCloudflareZones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCloudflareZonesRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"match": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"lookup_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"contains", "exact"}, false),
							Default:      "exact",
						},
						"status": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"paused": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"zones": {
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
					},
				},
			},
		},
	}
}

func dataSourceCloudflareZonesRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading Zones")
	client := meta.(*cloudflare.API)
	filter, err := expandFilter(d.Get("filter"))
	if err != nil {
		return err
	}

	zoneLookupValue := filter.name
	if filter.lookupType == "contains" {
		zoneLookupValue = "contains:" + zoneLookupValue
	}

	zoneFilter := cloudflare.WithZoneFilters(
		zoneLookupValue,
		"",
		filter.status,
	)

	zones, err := client.ListZonesContext(context.Background(), zoneFilter)
	if err != nil {
		return fmt.Errorf("error listing Zone: %s", err)
	}

	zoneIds := make([]string, 0)
	zoneDetails := make([]interface{}, 0)
	for _, v := range zones.Result {
		if filter.regexValue != nil {
			if !filter.regexValue.Match([]byte(v.Name)) {
				continue
			}
		}

		if filter.paused != v.Paused {
			continue
		}

		zoneDetails = append(zoneDetails, map[string]interface{}{
			"id":   v.ID,
			"name": v.Name,
		})
		zoneIds = append(zoneIds, v.ID)
	}

	err = d.Set("zones", zoneDetails)
	if err != nil {
		return fmt.Errorf("Error setting zones: %s", err)
	}

	d.SetId(stringListChecksum(zoneIds))
	return nil
}

func expandFilter(d interface{}) (*searchFilter, error) {
	cfg := d.([]interface{})
	filter := &searchFilter{}

	m := cfg[0].(map[string]interface{})
	name, ok := m["name"]
	if ok {
		filter.name = name.(string)
	}

	match, ok := m["match"]
	if ok {
		match, err := regexp.Compile(match.(string))
		if err != nil {
			return nil, err
		}

		filter.regexValue = match
	}

	lookupType, ok := m["lookup_type"]
	if ok {
		filter.lookupType = lookupType.(string)
	}

	paused, ok := m["paused"]
	if ok {
		filter.paused = paused.(bool)
	}

	status, ok := m["status"]
	if ok {
		filter.status = status.(string)
	}

	return filter, nil
}

type searchFilter struct {
	name       string
	regexValue *regexp.Regexp
	lookupType string
	status     string
	paused     bool
}
