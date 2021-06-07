package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceCloudflareLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareLoadBalancerCreate,
		Read:   resourceCloudflareLoadBalancerRead,
		Update: resourceCloudflareLoadBalancerUpdate,
		Delete: resourceCloudflareLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareLoadBalancerImport,
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"fallback_pool_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},

			"default_pool_ids": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 32),
				},
			},

			"session_affinity": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validation.StringInSlice([]string{"", "none", "cookie", "ip_cookie"}, false),
			},

			"proxied": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"ttl"},
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"ttl": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"proxied"}, // this is set to zero regardless of config when proxied=true
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},

			"steering_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"off", "geo", "dynamic_latency", "random", ""}, false),
				Computed:     true,
			},

			"session_affinity_ttl": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      nil,
				ValidateFunc: validation.IntBetween(1800, 604800),
			},

			"session_affinity_attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     rulesElem,
			},

			// nb enterprise only
			"pop_pools": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     popPoolElem,
			},

			"region_pools": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     regionPoolElem,
			},

			"created_on": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"modified_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

var rulesElem = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 200),
		},

		"priority": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},

		"disabled": {
			Type:     schema.TypeBool,
			Optional: true,
		},

		"condition": {
			Type:     schema.TypeString,
			Optional: true,
		},

		"terminates": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},

		"overrides": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{

					"session_affinity": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"", "none", "cookie", "ip_cookie"}, false),
					},

					"session_affinity_ttl": {
						Type:         schema.TypeInt,
						Optional:     true,
						ValidateFunc: validation.IntBetween(1800, 604800),
					},

					"session_affinity_attributes": {
						Type:     schema.TypeMap,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},

					"ttl": {
						Type:     schema.TypeInt,
						Optional: true,
					},

					"steering_policy": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"off", "geo", "dynamic_latency", "random", "proximity", ""}, false),
					},

					"fallback_pool": {
						Type:     schema.TypeString,
						Optional: true,
					},

					"default_pools": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},

					"pop_pools": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem:     popPoolElem,
					},

					"region_pools": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem:     regionPoolElem,
					},
				},
			},
		},

		"fixed_response": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"message_body": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringLenBetween(0, 1024),
					},

					"status_code": {
						Type:     schema.TypeInt,
						Optional: true,
					},

					"content_type": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringLenBetween(0, 32),
					},

					"location": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringLenBetween(0, 2048),
					},
				},
			},
		},
	},
}

var popPoolElem = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"pop": {
			Type:     schema.TypeString,
			Required: true,
			// let the api handle validating pops
		},

		"pool_ids": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
		},
	},
}

var regionPoolElem = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"region": {
			Type:     schema.TypeString,
			Required: true,
			// let the api handle validating regions
		},

		"pool_ids": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
		},
	},
}

var localPoolElems = map[string]*schema.Resource{
	"pop":    popPoolElem,
	"region": regionPoolElem,
}

func resourceCloudflareLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	zoneID := d.Get("zone_id").(string)

	enabled := d.Get("enabled").(bool)
	newLoadBalancer := cloudflare.LoadBalancer{
		Name:           d.Get("name").(string),
		FallbackPool:   d.Get("fallback_pool_id").(string),
		DefaultPools:   expandInterfaceToStringList(d.Get("default_pool_ids")),
		Proxied:        d.Get("proxied").(bool),
		Enabled:        &enabled,
		TTL:            d.Get("ttl").(int),
		SteeringPolicy: d.Get("steering_policy").(string),
		Persistence:    d.Get("session_affinity").(string),
	}

	if description, ok := d.GetOk("description"); ok {
		newLoadBalancer.Description = description.(string)
	}

	if ttl, ok := d.GetOk("ttl"); ok {
		newLoadBalancer.TTL = ttl.(int)
	}

	if regionPools, ok := d.GetOk("region_pools"); ok {
		expandedRegionPools, err := expandGeoPools(regionPools, "region")
		if err != nil {
			return err
		}
		newLoadBalancer.RegionPools = expandedRegionPools
	}

	if popPools, ok := d.GetOk("pop_pools"); ok {
		expandedPopPools, err := expandGeoPools(popPools, "pop")
		if err != nil {
			return err
		}
		newLoadBalancer.PopPools = expandedPopPools
	}

	if sessionAffinityTTL, ok := d.GetOk("session_affinity_ttl"); ok {
		newLoadBalancer.PersistenceTTL = sessionAffinityTTL.(int)
	}

	if sessionAffinityAttrs, ok := d.GetOk("session_affinity_attributes"); ok {
		sessionAffinityAttributes, err := expandSessionAffinityAttrs(sessionAffinityAttrs)
		if err != nil {
			return err
		}
		newLoadBalancer.SessionAffinityAttributes = sessionAffinityAttributes
	}

	if rules, ok := d.GetOk("rules"); ok {
		v, err := expandRules(rules)
		if err != nil {
			return err
		}
		newLoadBalancer.Rules = v
	}

	log.Printf("[INFO] Creating Cloudflare Load Balancer from struct: %+v", newLoadBalancer)

	r, err := client.CreateLoadBalancer(context.Background(), zoneID, newLoadBalancer)
	if err != nil {
		return errors.Wrap(err, "error creating load balancer for zone")
	}

	if r.ID == "" {
		return fmt.Errorf("failed to find id in Create response; resource was empty")
	}

	d.SetId(r.ID)

	log.Printf("[INFO] Cloudflare Load Balancer ID: %s", d.Id())

	return resourceCloudflareLoadBalancerRead(d, meta)
}

func resourceCloudflareLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	// since api only supports replace, update looks a lot like create...
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	enabled := d.Get("enabled").(bool)
	loadBalancer := cloudflare.LoadBalancer{
		ID:             d.Id(),
		Name:           d.Get("name").(string),
		FallbackPool:   d.Get("fallback_pool_id").(string),
		DefaultPools:   expandInterfaceToStringList(d.Get("default_pool_ids")),
		Proxied:        d.Get("proxied").(bool),
		Enabled:        &enabled,
		TTL:            d.Get("ttl").(int),
		SteeringPolicy: d.Get("steering_policy").(string),
		Persistence:    d.Get("session_affinity").(string),
	}

	if description, ok := d.GetOk("description"); ok {
		loadBalancer.Description = description.(string)
	}

	if regionPools, ok := d.GetOk("region_pools"); ok {
		expandedRegionPools, err := expandGeoPools(regionPools, "region")
		if err != nil {
			return err
		}
		loadBalancer.RegionPools = expandedRegionPools
	}

	if popPools, ok := d.GetOk("pop_pools"); ok {
		expandedPopPools, err := expandGeoPools(popPools, "pop")
		if err != nil {
			return err
		}
		loadBalancer.PopPools = expandedPopPools
	}

	if sessionAffinityTTL, ok := d.GetOk("session_affinity_ttl"); ok {
		loadBalancer.PersistenceTTL = sessionAffinityTTL.(int)
	}

	if sessionAffinityAttrs, ok := d.GetOk("session_affinity_attributes"); ok {
		sessionAffinityAttributes, err := expandSessionAffinityAttrs(sessionAffinityAttrs)
		if err != nil {
			return err
		}
		loadBalancer.SessionAffinityAttributes = sessionAffinityAttributes
	}

	if rules, ok := d.GetOk("rules"); ok {
		v, err := expandRules(rules)
		if err != nil {
			return err
		}
		loadBalancer.Rules = v
	}

	log.Printf("[INFO] Updating Cloudflare Load Balancer from struct: %+v", loadBalancer)

	_, err := client.ModifyLoadBalancer(context.Background(), zoneID, loadBalancer)
	if err != nil {
		return errors.Wrap(err, "error creating load balancer for zone")
	}

	return resourceCloudflareLoadBalancerRead(d, meta)
}

func expandGeoPools(pool interface{}, geoType string) (map[string][]string, error) {
	cfg := pool.(*schema.Set).List()
	expanded := make(map[string][]string)
	for _, v := range cfg {
		locationConfig := v.(map[string]interface{})
		// lists are of type interface{} by default
		location := locationConfig[geoType].(string)
		if _, present := expanded[location]; !present {
			expanded[location] = expandInterfaceToStringList(locationConfig["pool_ids"])
		} else {
			return nil, fmt.Errorf("duplicate entry specified for %s pool in location %q. each location must only be specified once", geoType, location)
		}
	}
	return expanded, nil
}

func resourceCloudflareLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	loadBalancerID := d.Id()

	loadBalancer, err := client.LoadBalancerDetails(context.Background(), zoneID, loadBalancerID)
	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Load balancer %s in zone %s not found", loadBalancerID, zoneID)
			d.SetId("")
			return nil
		}
		return errors.Wrap(err,
			fmt.Sprintf("Error reading load balancer resource from API for resource %s in zone %s", zoneID, loadBalancerID))
	}

	d.Set("name", loadBalancer.Name)
	d.Set("fallback_pool_id", loadBalancer.FallbackPool)
	d.Set("proxied", loadBalancer.Proxied)
	d.Set("enabled", loadBalancer.Enabled)
	d.Set("description", loadBalancer.Description)
	d.Set("ttl", loadBalancer.TTL)
	d.Set("steering_policy", loadBalancer.SteeringPolicy)
	d.Set("session_affinity", loadBalancer.Persistence)

	d.Set("created_on", loadBalancer.CreatedOn.Format(time.RFC3339Nano))
	d.Set("modified_on", loadBalancer.ModifiedOn.Format(time.RFC3339Nano))

	if _, sessionAffinityAttrsOk := d.GetOk("session_affinity_attributes"); sessionAffinityAttrsOk {
		if err := d.Set("session_affinity_attributes", flattenSessionAffinityAttrs(loadBalancer.SessionAffinityAttributes)); err != nil {
			return fmt.Errorf("failed to set session_affinity_attributes: %s", err)
		}
	}

	if len(loadBalancer.Rules) > 0 {
		fr, err := flattenRules(d, loadBalancer.Rules)
		if err != nil {
			return fmt.Errorf("failed to flatten rules: %s", err)
		}
		if err := d.Set("rules", fr); err != nil {
			return fmt.Errorf("failed to set rules: %s\n %v", err, fr)
		}
	}

	if err := d.Set("default_pool_ids", loadBalancer.DefaultPools); err != nil {
		log.Printf("[WARN] Error setting default_pool_ids on load balancer %q: %s", d.Id(), err)
	}

	if err := d.Set("pop_pools", flattenGeoPools(loadBalancer.PopPools, "pop")); err != nil {
		log.Printf("[WARN] Error setting pop_pools on load balancer %q: %s", d.Id(), err)
	}

	if err := d.Set("region_pools", flattenGeoPools(loadBalancer.RegionPools, "region")); err != nil {
		log.Printf("[WARN] Error setting region_pools on load balancer %q: %s", d.Id(), err)
	}

	if loadBalancer.PersistenceTTL != 0 {
		d.Set("session_affinity_ttl", loadBalancer.PersistenceTTL)
	}

	return nil
}

func flattenGeoPools(pools map[string][]string, geoType string) *schema.Set {
	flattened := make([]interface{}, 0)
	for k, v := range pools {
		geoConf := map[string]interface{}{
			geoType:    k,
			"pool_ids": flattenStringList(v),
		}
		flattened = append(flattened, geoConf)
	}
	return schema.NewSet(schema.HashResource(localPoolElems[geoType]), flattened)
}

func flattenSessionAffinityAttrs(attrs *cloudflare.SessionAffinityAttributes) map[string]interface{} {
	return map[string]interface{}{
		"drain_duration": strconv.Itoa(attrs.DrainDuration),
		"samesite":       attrs.SameSite,
		"secure":         attrs.Secure,
	}
}

func resourceCloudflareLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	loadBalancerID := d.Id()

	log.Printf("[INFO] Deleting Cloudflare Load Balancer: %s in zone: %s", loadBalancerID, zoneID)

	err := client.DeleteLoadBalancer(context.Background(), zoneID, loadBalancerID)
	if err != nil {
		return fmt.Errorf("error deleting Cloudflare Load Balancer: %s", err)
	}

	return nil
}

func resourceCloudflareLoadBalancerImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// split the id so we can lookup
	idAttr := strings.SplitN(d.Id(), "/", 2)
	var zoneID string
	var loadBalancerID string
	if len(idAttr) == 2 {
		zoneID = idAttr[0]
		loadBalancerID = idAttr[1]
	} else {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/loadBalancerID\"", d.Id())
	}

	d.Set("zone_id", zoneID)
	d.SetId(loadBalancerID)

	resourceCloudflareLoadBalancerRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

func flattenRules(d *schema.ResourceData, rules []*cloudflare.LoadBalancerRule) (interface{}, error) {
	if len(rules) == 0 {
		return nil, nil
	}

	cfResources := []map[string]interface{}{}
	for idx, r := range rules {
		m := map[string]interface{}{
			"name":      r.Name,
			"condition": r.Condition,
			"disabled":  r.Disabled,
		}

		if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.priority", idx)); ok {
			m["priority"] = r.Priority
		}
		if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.terminates", idx)); ok {
			m["terminates"] = r.Terminates
		}

		if fr := r.FixedResponse; fr != nil {
			frm := map[string]interface{}{}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.fixed_response.message_body", idx)); ok {
				frm["message_body"] = fr.MessageBody
				m["fixed_response"] = frm // only set if one of these has is true
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.fixed_response.status_code", idx)); ok {
				frm["status_code"] = strconv.FormatInt(int64(fr.StatusCode), 10)
				m["fixed_response"] = frm // only set if one of these has is true
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.fixed_response.content_type", idx)); ok {
				frm["content_type"] = fr.ContentType
				m["fixed_response"] = frm // only set if one of these has is true
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.fixed_response.location", idx)); ok {
				frm["location"] = fr.Location
				m["fixed_response"] = frm // only set if one of these has is true
			}
		}
		if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides", idx)); ok {
			o := r.Overrides
			om := map[string]interface{}{}

			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.session_affinity", idx)); ok {
				om["session_affinity"] = o.Persistence
				m["overrides"] = []interface{}{om}
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.session_affinity_ttl", idx)); ok {
				om["session_affinity"] = o.PersistenceTTL
				m["overrides"] = []interface{}{om}
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.ttl", idx)); ok {
				om["ttl"] = o.TTL
				m["overrides"] = []interface{}{om}
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.steering_policy", idx)); ok {
				om["steering_policy"] = o.SteeringPolicy
				m["overrides"] = []interface{}{om}
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.fallback_pool", idx)); ok {
				om["fallback_pool"] = o.FallbackPool
				m["overrides"] = []interface{}{om}
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.default_pools", idx)); ok {
				om["default_pools"] = o.DefaultPools
				m["overrides"] = []interface{}{om}
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.pop_pools", idx)); ok {
				om["pop_pools"] = flattenGeoPools(o.PoPPools, "pop")
				m["overrides"] = []interface{}{om}
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.region_pools", idx)); ok {
				om["region_pools"] = flattenGeoPools(o.RegionPools, "region")
				m["overrides"] = []interface{}{om}
			}
			if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.session_affinity_attributes", idx)); o.SessionAffinityAttrs != nil && ok {
				saa := map[string]interface{}{}
				om["session_affinity_attributes"] = saa
				m["overrides"] = []interface{}{om}
				if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.session_affinity_attributes.samesite", idx)); ok {
					saa["samesite"] = o.SessionAffinityAttrs.SameSite
				}
				if _, ok := d.GetOkExists(fmt.Sprintf("rules.%d.overrides.0.session_affinity_attributes.secure", idx)); ok {
					saa["secure"] = o.SessionAffinityAttrs.Secure
				}
			}
		}

		cfResources = append(cfResources, m)
	}
	return cfResources, nil
}

func expandRules(rdata interface{}) ([]*cloudflare.LoadBalancerRule, error) {
	var rules []*cloudflare.LoadBalancerRule
	for _, ele := range rdata.([]interface{}) {
		r := ele.(map[string]interface{})
		lbr := &cloudflare.LoadBalancerRule{
			Name: r["name"].(string),
		}
		if v, ok := r["priority"]; ok {
			lbr.Priority = v.(int)
		}
		if d, ok := r["disabled"]; ok {
			lbr.Disabled = d.(bool)
		}
		if c, ok := r["condition"]; ok {
			lbr.Condition = c.(string)
		}
		if t, ok := r["terminates"]; ok {
			lbr.Terminates = t.(bool)
		}

		if overridesData, ok := r["overrides"]; ok && len(overridesData.([]interface{})) > 0 {
			ov := overridesData.([]interface{})[0].(map[string]interface{})

			if sa, ok := ov["session_affinity"]; ok {
				lbr.Overrides.Persistence = sa.(string)
			}

			if sattl, ok := ov["session_affinity_ttl"]; ok {
				v := uint(sattl.(int))
				// a default value of seem to be set into this field bypassing
				// the IntBetween(1800, 604800) validation check ignore
				// this zero values here
				if v != 0 {
					lbr.Overrides.PersistenceTTL = &v
				}
			}

			if saattr, ok := ov["session_affinity_attributes"]; ok {
				attr := saattr.(map[string]interface{})
				v := &cloudflare.LoadBalancerRuleOverridesSessionAffinityAttrs{}
				if ss, ok := attr["samesite"]; ok {
					v.SameSite = ss.(string)
					lbr.Overrides.SessionAffinityAttrs = v
				}
				if sec, ok := attr["secure"]; ok {
					v.Secure = sec.(string)
					lbr.Overrides.SessionAffinityAttrs = v
				}
			}

			if ttl, ok := ov["ttl"]; ok {
				lbr.Overrides.TTL = uint(ttl.(int))
			}

			if sp, ok := ov["steering_policy"]; ok {
				lbr.Overrides.SteeringPolicy = sp.(string)
			}

			if fb, ok := ov["fallback_pool"]; ok {
				lbr.Overrides.FallbackPool = fb.(string)
			}

			if dp, ok := ov["default_pools"]; ok {
				lbr.Overrides.DefaultPools = expandInterfaceToStringList(dp)
			}

			if pp, ok := ov["pop_pools"]; ok {
				expandedPopPools, err := expandGeoPools(pp, "pop")
				if err != nil {
					return nil, err
				}
				lbr.Overrides.PoPPools = expandedPopPools
			}

			if rp, ok := ov["region_pools"]; ok {
				expandedRegionPools, err := expandGeoPools(rp, "region")
				if err != nil {
					return nil, err
				}
				lbr.Overrides.RegionPools = expandedRegionPools
			}
		}

		if fixedResponseData, ok := r["fixed_response"]; ok {
			frd := fixedResponseData.(map[string]interface{})
			// we don't add this into our LB unless one of the cases below is true
			fr := &cloudflare.LoadBalancerFixedResponseData{}
			if mb, ok := frd["message_body"]; ok {
				fr.MessageBody = mb.(string)
				lbr.FixedResponse = fr
			}
			if sc, ok := frd["status_code"]; ok {
				scint, err := strconv.ParseInt(sc.(string), 10, 64)
				if err != nil {
					return nil, err
				}
				fr.StatusCode = int(scint)
				lbr.FixedResponse = fr
			}
			if ct, ok := frd["content_type"]; ok {
				fr.ContentType = ct.(string)
				lbr.FixedResponse = fr
			}
			if l, ok := frd["location"]; ok {
				fr.Location = l.(string)
				lbr.FixedResponse = fr
			}

		}
		rules = append(rules, lbr)
	}
	return rules, nil
}

func expandSessionAffinityAttrs(attrs interface{}) (*cloudflare.SessionAffinityAttributes, error) {
	var cfSessionAffinityAttrs cloudflare.SessionAffinityAttributes

	for k, v := range attrs.(map[string]interface{}) {
		switch k {
		case "secure":
			cfSessionAffinityAttrs.Secure = v.(string)
		case "samesite":
			cfSessionAffinityAttrs.SameSite = v.(string)
		case "drain_duration":
			var err error
			if cfSessionAffinityAttrs.DrainDuration, err = strconv.Atoi(v.(string)); err != nil {
				return nil, err
			}
		}
	}

	return &cfSessionAffinityAttrs, nil
}
