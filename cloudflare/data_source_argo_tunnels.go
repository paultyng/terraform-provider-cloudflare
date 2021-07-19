package cloudflare

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceCloudflareArgoTunnels() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCloudflareArgoTunnelsRead,

		Schema: map[string]*schema.Schema{
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
						"match": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"deleted": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"tunnels": {
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
						"deleted": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudflareArgoTunnelsRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading Argo Tunnels")
	client := meta.(*cloudflare.API)
	filter, err := expandArgoTunnelFilter(d.Get("filter"))
	if err != nil {
		return err
	}

	tunnels, err := client.ArgoTunnels(context.Background(), client.AccountID)
	if err != nil {
		return fmt.Errorf("error listing Tunnels: %s", err)
	}

	tunnelIds := make([]string, 0)
	tunnelDetails := make([]interface{}, 0)
	for _, v := range tunnels {
		if filter.regexValue != nil {
			if !filter.regexValue.Match([]byte(v.Name)) {
				continue
			}
		}

		if filter.name != "" && filter.name != v.Name {
			continue
		}

		if filter.is_deleted && v.DeletedAt == nil {
			continue
		}

		if !filter.is_deleted && v.DeletedAt != nil {
			continue
		}

		tunnelDetails = append(tunnelDetails, map[string]interface{}{
			"id":      v.ID,
			"name":    v.Name,
			"deleted": v.DeletedAt != nil,
		})
		tunnelIds = append(tunnelIds, v.ID)
	}

	err = d.Set("tunnels", tunnelDetails)
	if err != nil {
		return fmt.Errorf("Error setting tunnels: %s", err)
	}

	d.SetId(stringListChecksum(tunnelIds))
	return nil
}

func expandArgoTunnelFilter(d interface{}) (*argoTunnelsearchFilter, error) {
	cfg := d.([]interface{})
	filter := &argoTunnelsearchFilter{}

	if len(cfg) < 1 {
		return filter, nil
	}

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

	is_deleted, ok := m["deleted"]
	if ok {
		filter.is_deleted = is_deleted.(bool)
	}

	return filter, nil
}

type argoTunnelsearchFilter struct {
	name       string
	regexValue *regexp.Regexp
	is_deleted bool
}
