package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	"time"

	"regexp"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceCloudflareLoadBalancerPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareLoadBalancerPoolCreate,
		Update: resourceCloudflareLoadBalancerPoolUpdate,
		Read:   resourceCloudflareLoadBalancerPoolRead,
		Delete: resourceCloudflareLoadBalancerPoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("[-_a-zA-Z0-9]+"), "Only alphanumeric characters, hyphens and underscores are allowed."),
			},

			"origins": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     originsElem,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"minimum_origins": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},

			"check_regions": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},

			"monitor": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 32),
			},

			"notification_email": {
				Type:     schema.TypeString,
				Optional: true,
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

var originsElem = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},

		"address": {
			Type:     schema.TypeString,
			Required: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validateStringIP,
			},
		},

		"weight": {
			Type:         schema.TypeFloat,
			Optional:     true,
			Default:      1.0,
			ValidateFunc: floatBetween(0.0, 1.0),
		},

		"enabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},

		"header": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"header": {
						Type:     schema.TypeString,
						Required: true,
					},
					"values": {
						Type:     schema.TypeSet,
						Required: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
			Set: HashByMapKey("header"),
		},
	},
}

func resourceCloudflareLoadBalancerPoolCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	loadBalancerPool := cloudflare.LoadBalancerPool{
		Name:           d.Get("name").(string),
		Origins:        expandLoadBalancerOrigins(d.Get("origins").(*schema.Set)),
		Enabled:        d.Get("enabled").(bool),
		MinimumOrigins: d.Get("minimum_origins").(int),
	}

	if checkRegions, ok := d.GetOk("check_regions"); ok {
		loadBalancerPool.CheckRegions = expandInterfaceToStringList(checkRegions.(*schema.Set).List())
	}

	if description, ok := d.GetOk("description"); ok {
		loadBalancerPool.Description = description.(string)
	}

	if monitor, ok := d.GetOk("monitor"); ok {
		loadBalancerPool.Monitor = monitor.(string)
	}

	if notificationEmail, ok := d.GetOk("notification_email"); ok {
		loadBalancerPool.NotificationEmail = notificationEmail.(string)
	}

	log.Printf("[DEBUG] Creating Cloudflare Load Balancer Pool from struct: %+v", loadBalancerPool)

	r, err := client.CreateLoadBalancerPool(context.Background(), loadBalancerPool)
	if err != nil {
		return errors.Wrap(err, "error creating load balancer pool")
	}

	if r.ID == "" {
		return fmt.Errorf("cailed to find id in create response; resource was empty")
	}

	d.SetId(r.ID)

	log.Printf("[INFO] New Cloudflare Load Balancer Pool created with  ID: %s", d.Id())

	return resourceCloudflareLoadBalancerPoolRead(d, meta)
}

func resourceCloudflareLoadBalancerPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	loadBalancerPool := cloudflare.LoadBalancerPool{
		ID:             d.Id(),
		Name:           d.Get("name").(string),
		Origins:        expandLoadBalancerOrigins(d.Get("origins").(*schema.Set)),
		Enabled:        d.Get("enabled").(bool),
		MinimumOrigins: d.Get("minimum_origins").(int),
	}

	if checkRegions, ok := d.GetOk("check_regions"); ok {
		loadBalancerPool.CheckRegions = expandInterfaceToStringList(checkRegions.(*schema.Set).List())
	}

	if description, ok := d.GetOk("description"); ok {
		loadBalancerPool.Description = description.(string)
	}

	if monitor, ok := d.GetOk("monitor"); ok {
		loadBalancerPool.Monitor = monitor.(string)
	}

	if notificationEmail, ok := d.GetOk("notification_email"); ok {
		loadBalancerPool.NotificationEmail = notificationEmail.(string)
	}

	log.Printf("[DEBUG] Updating Cloudflare Load Balancer Pool from struct: %+v", loadBalancerPool)

	_, err := client.ModifyLoadBalancerPool(context.Background(), loadBalancerPool)
	if err != nil {
		return errors.Wrap(err, "error updating load balancer pool")
	}

	return resourceCloudflareLoadBalancerPoolRead(d, meta)
}

func expandLoadBalancerPoolHeader(cfgSet interface{}) map[string][]string {
	header := make(map[string][]string)
	cfgList := cfgSet.(*schema.Set).List()
	for _, item := range cfgList {
		cfg := item.(map[string]interface{})
		header[cfg["header"].(string)] = expandInterfaceToStringList(cfg["values"].(*schema.Set).List())
	}
	return header
}

func flattenLoadBalancerPoolHeader(header map[string][]string) *schema.Set {
	flattened := make([]interface{}, 0)
	for k, v := range header {
		cfg := map[string]interface{}{
			"header": k,
			"values": schema.NewSet(schema.HashString, flattenStringList(v)),
		}
		flattened = append(flattened, cfg)
	}
	return schema.NewSet(HashByMapKey("header"), flattened)
}

func expandLoadBalancerOrigins(originSet *schema.Set) (origins []cloudflare.LoadBalancerOrigin) {
	for _, iface := range originSet.List() {
		o := iface.(map[string]interface{})
		origin := cloudflare.LoadBalancerOrigin{
			Name:    o["name"].(string),
			Address: o["address"].(string),
			Enabled: o["enabled"].(bool),
			Weight:  o["weight"].(float64),
		}

		if header, ok := o["header"]; ok {
			origin.Header = expandLoadBalancerPoolHeader(header)
		}

		origins = append(origins, origin)
	}
	return
}

func resourceCloudflareLoadBalancerPoolRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	loadBalancerPool, err := client.LoadBalancerPoolDetails(context.Background(), d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Load balancer pool %s no longer exists", d.Id())
			d.SetId("")
			return nil
		} else {
			return errors.Wrap(err,
				fmt.Sprintf("Error reading load balancer pool from API for resource %s ", d.Id()))
		}
	}
	log.Printf("[DEBUG] Read Cloudflare Load Balancer Pool from API as struct: %+v", loadBalancerPool)

	d.Set("name", loadBalancerPool.Name)
	d.Set("enabled", loadBalancerPool.Enabled)
	d.Set("minimum_origins", loadBalancerPool.MinimumOrigins)
	d.Set("description", loadBalancerPool.Description)
	d.Set("monitor", loadBalancerPool.Monitor)
	d.Set("notification_email", loadBalancerPool.NotificationEmail)
	d.Set("created_on", loadBalancerPool.CreatedOn.Format(time.RFC3339Nano))
	d.Set("modified_on", loadBalancerPool.ModifiedOn.Format(time.RFC3339Nano))

	if err := d.Set("origins", flattenLoadBalancerOrigins(d, loadBalancerPool.Origins)); err != nil {
		log.Printf("[WARN] Error setting origins on load balancer pool %q: %s", d.Id(), err)
	}

	if err := d.Set("check_regions", schema.NewSet(schema.HashString, flattenStringList(loadBalancerPool.CheckRegions))); err != nil {
		log.Printf("[WARN] Error setting check_regions on load balancer pool %q: %s", d.Id(), err)
	}

	return nil
}

func flattenLoadBalancerOrigins(d *schema.ResourceData, origins []cloudflare.LoadBalancerOrigin) *schema.Set {
	flattened := make([]interface{}, 0)
	for _, o := range origins {
		cfg := map[string]interface{}{
			"name":    o.Name,
			"address": o.Address,
			"enabled": o.Enabled,
			"weight":  o.Weight,
			"header":  flattenLoadBalancerPoolHeader(o.Header),
		}

		flattened = append(flattened, cfg)
	}
	return schema.NewSet(schema.HashResource(originsElem), flattened)
}

func resourceCloudflareLoadBalancerPoolDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	log.Printf("[INFO] Deleting Cloudflare Load Balancer Pool: %s ", d.Id())

	err := client.DeleteLoadBalancerPool(context.Background(), d.Id())
	if err != nil {
		return errors.Wrap(err, "error deleting Cloudflare Load Balancer Pool")
	}

	return nil
}

// floatBetween returns a validate function that can be used in schema definitions.
func floatBetween(min, max float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(float64)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be float64", k))
			return
		}

		if v < min || v > max {
			es = append(es, fmt.Errorf("expected %s to be within %v and %v", k, min, max))
			return
		}

		return
	}
}
