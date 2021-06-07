package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceCloudflareLoadBalancerMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareLoadBalancerPoolMonitorCreate,
		Read:   resourceCloudflareLoadBalancerPoolMonitorRead,
		Update: resourceCloudflareLoadBalancerPoolMonitorUpdate,
		Delete: resourceCloudflareLoadBalancerPoolMonitorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			//
			// common
			//
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
			},
			// interval has to be larger than (retries+1) * probe_timeout:

			"method": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 65535),
			},

			"retries": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				ValidateFunc: validation.IntBetween(1, 5),
			},

			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      5,
				ValidateFunc: validation.IntBetween(1, 10),
			},

			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "http",
				ValidateFunc: validation.StringInSlice([]string{"http", "https", "tcp"}, false),
			},

			"created_on": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"modified_on": {
				Type:     schema.TypeString,
				Computed: true,
			},

			//
			// http/https only
			//
			"allow_insecure": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"expected_body": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"expected_codes": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"follow_redirects": {
				Type:     schema.TypeBool,
				Optional: true,
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

			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"probe_zone": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceCloudflareLoadBalancerPoolMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	loadBalancerMonitor := cloudflare.LoadBalancerMonitor{
		Timeout:  d.Get("timeout").(int),
		Type:     d.Get("type").(string),
		Interval: d.Get("interval").(int),
		Retries:  d.Get("retries").(int),
	}

	if description, ok := d.GetOk("description"); ok {
		loadBalancerMonitor.Description = description.(string)
	}

	if port, ok := d.GetOk("port"); ok {
		loadBalancerMonitor.Port = uint16(port.(int))
	}

	switch loadBalancerMonitor.Type {
	case "tcp":
		if method, ok := d.GetOk("method"); ok {
			loadBalancerMonitor.Method = method.(string)
		} else {
			loadBalancerMonitor.Method = "connection_established"
		}
	case "http", "https":
		if allowInsecure, ok := d.GetOk("allow_insecure"); ok {
			loadBalancerMonitor.AllowInsecure = allowInsecure.(bool)
		}

		expectedBody := d.Get("expected_body")
		loadBalancerMonitor.ExpectedBody = expectedBody.(string)

		if expectedCodes, ok := d.GetOk("expected_codes"); ok {
			loadBalancerMonitor.ExpectedCodes = expectedCodes.(string)
		} else {
			return fmt.Errorf("expected_codes must be set")
		}

		if followRedirects, ok := d.GetOk("follow_redirects"); ok {
			loadBalancerMonitor.FollowRedirects = followRedirects.(bool)
		}

		if method, ok := d.GetOk("method"); ok {
			loadBalancerMonitor.Method = method.(string)
		} else {
			loadBalancerMonitor.Method = "GET"
		}

		if header, ok := d.GetOk("header"); ok {
			loadBalancerMonitor.Header = expandLoadBalancerMonitorHeader(header)
		}

		if path, ok := d.GetOk("path"); ok {
			loadBalancerMonitor.Path = path.(string)
		} else {
			loadBalancerMonitor.Path = "/"
		}

		if probeZone, ok := d.GetOk("probe_zone"); ok {
			loadBalancerMonitor.ProbeZone = probeZone.(string)
		} else {
			loadBalancerMonitor.ProbeZone = ""
		}
	}

	log.Printf("[DEBUG] Creating Cloudflare Load Balancer Monitor from struct: %+v", loadBalancerMonitor)

	r, err := client.CreateLoadBalancerMonitor(context.Background(), loadBalancerMonitor)
	if err != nil {
		return errors.Wrap(err, "error creating load balancer monitor")
	}

	if r.ID == "" {
		return fmt.Errorf("failed to find id in create response; resource was empty")
	}

	d.SetId(r.ID)

	log.Printf("[INFO] New Cloudflare Load Balancer Monitor created with  ID: %s", d.Id())

	return resourceCloudflareLoadBalancerPoolMonitorRead(d, meta)
}

func resourceCloudflareLoadBalancerPoolMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	loadBalancerMonitor := cloudflare.LoadBalancerMonitor{
		ID:       d.Id(),
		Timeout:  d.Get("timeout").(int),
		Type:     d.Get("type").(string),
		Interval: d.Get("interval").(int),
		Retries:  d.Get("retries").(int),
	}

	if description, ok := d.GetOk("description"); ok {
		loadBalancerMonitor.Description = description.(string)
	}

	if port, ok := d.GetOk("port"); ok {
		loadBalancerMonitor.Port = uint16(port.(int))
	}

	switch loadBalancerMonitor.Type {
	case "tcp":
		if method, ok := d.GetOk("method"); ok {
			loadBalancerMonitor.Method = method.(string)
		} else {
			loadBalancerMonitor.Method = "connection_established"
		}
	case "http", "https":
		if allowInsecure, ok := d.GetOk("allow_insecure"); ok {
			loadBalancerMonitor.AllowInsecure = allowInsecure.(bool)
		}

		if expectedBody, ok := d.GetOk("expected_body"); ok {
			loadBalancerMonitor.ExpectedBody = expectedBody.(string)
		} else {
			loadBalancerMonitor.ExpectedBody = ""
		}

		if expectedCodes, ok := d.GetOk("expected_codes"); ok {
			loadBalancerMonitor.ExpectedCodes = expectedCodes.(string)
		} else {
			return fmt.Errorf("expected_codes must be set")
		}

		if header, ok := d.GetOk("header"); ok {
			loadBalancerMonitor.Header = expandLoadBalancerMonitorHeader(header)
		}

		if followRedirects, ok := d.GetOk("follow_redirects"); ok {
			loadBalancerMonitor.FollowRedirects = followRedirects.(bool)
		}

		if method, ok := d.GetOk("method"); ok {
			loadBalancerMonitor.Method = method.(string)
		} else {
			loadBalancerMonitor.Method = "GET"
		}

		if path, ok := d.GetOk("path"); ok {
			loadBalancerMonitor.Path = path.(string)
		} else {
			loadBalancerMonitor.Path = "/"
		}

		if probeZone, ok := d.GetOk("probe_zone"); ok {
			loadBalancerMonitor.ProbeZone = probeZone.(string)
		} else {
			loadBalancerMonitor.ProbeZone = ""
		}
	}

	log.Printf("[DEBUG] Update Cloudflare Load Balancer Monitor from struct: %+v", loadBalancerMonitor)

	_, err := client.ModifyLoadBalancerMonitor(context.Background(), loadBalancerMonitor)
	if err != nil {
		return errors.Wrap(err, "error modifying load balancer monitor")
	}

	log.Printf("[INFO] Cloudflare Load Balancer Monitor %q was modified", d.Id())

	return resourceCloudflareLoadBalancerPoolMonitorRead(d, meta)
}

func expandLoadBalancerMonitorHeader(cfgSet interface{}) map[string][]string {
	header := make(map[string][]string)
	cfgList := cfgSet.(*schema.Set).List()
	for _, item := range cfgList {
		cfg := item.(map[string]interface{})
		header[cfg["header"].(string)] = expandInterfaceToStringList(cfg["values"].(*schema.Set).List())
	}
	return header
}

func resourceCloudflareLoadBalancerPoolMonitorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	loadBalancerMonitor, err := client.LoadBalancerMonitorDetails(context.Background(), d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Load balancer monitor %s no longer exists", d.Id())
			d.SetId("")
			return nil
		} else {
			return errors.Wrap(err,
				fmt.Sprintf("Error reading load balancer monitor from API for resource %s ", d.Id()))
		}
	}
	log.Printf("[DEBUG] Read Cloudflare Load Balancer Monitor from API as struct: %+v", loadBalancerMonitor)

	if loadBalancerMonitor.Type == "http" || loadBalancerMonitor.Type == "https" {
		d.Set("allow_insecure", loadBalancerMonitor.AllowInsecure)
		d.Set("expected_body", loadBalancerMonitor.ExpectedBody)
		d.Set("expected_codes", loadBalancerMonitor.ExpectedCodes)
		d.Set("follow_redirects", loadBalancerMonitor.FollowRedirects)
		d.Set("path", loadBalancerMonitor.Path)
		d.Set("probe_zone", loadBalancerMonitor.ProbeZone)

		if err := d.Set("header", flattenLoadBalancerMonitorHeader(loadBalancerMonitor.Header)); err != nil {
			log.Printf("[WARN] Error setting header for load balancer monitor %q: %s", d.Id(), err)
		}
	}

	d.Set("description", loadBalancerMonitor.Description)
	d.Set("interval", loadBalancerMonitor.Interval)
	d.Set("method", loadBalancerMonitor.Method)
	d.Set("port", int(loadBalancerMonitor.Port))
	d.Set("retries", loadBalancerMonitor.Retries)
	d.Set("timeout", loadBalancerMonitor.Timeout)
	d.Set("type", loadBalancerMonitor.Type)
	d.Set("created_on", loadBalancerMonitor.CreatedOn.Format(time.RFC3339Nano))
	d.Set("modified_on", loadBalancerMonitor.ModifiedOn.Format(time.RFC3339Nano))

	return nil
}

func flattenLoadBalancerMonitorHeader(header map[string][]string) *schema.Set {
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

func resourceCloudflareLoadBalancerPoolMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	log.Printf("[INFO] Deleting Cloudflare Load Balancer Monitor: %s ", d.Id())

	err := client.DeleteLoadBalancerMonitor(context.Background(), d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Load balancer monitor %s no longer exists", d.Id())
			return nil
		} else {
			return errors.Wrap(err, "error deleting cloudflare load balancer monitor")
		}
	}

	return nil
}
