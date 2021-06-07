package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceCloudflareHealthcheck() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareHealthcheckCreate,
		Read:   resourceCloudflareHealthcheckRead,
		Update: resourceCloudflareHealthcheckUpdate,
		Delete: resourceCloudflareHealthcheckDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareHealthcheckImport,
		},

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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"suspended": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"consecutive_fails": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"consecutive_successes": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"retries": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
			},
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  5,
			},
			"interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
			},
			"check_regions": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"WNAM", "ENAM", "WEU", "EEU", "NSAM", "SSAM", "OC", "ME", "NAF", "SAF", "IN", "SEAS", "NEAS", "ALL_REGIONS"}, false),
				},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"TCP", "HTTP", "HTTPS"}, false),
			},
			"method": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"connection_established", "GET", "HEAD"}, false),
			},
			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      80,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"expected_codes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"expected_body": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"follow_redirects": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allow_insecure": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
			},
			"notification_suspended": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"notification_email_addresses": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

func resourceCloudflareHealthcheckRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	healthcheck, err := client.Healthcheck(context.Background(), zoneID, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "object does not exist") {
			log.Printf("[INFO] Healthcheck %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return errors.Wrap(err, fmt.Sprintf("error reading healthcheck information for %q", d.Id()))
	}

	switch healthcheck.Type {
	case "TCP":
		d.Set("method", healthcheck.TCPConfig.Method)
		d.Set("port", int(healthcheck.TCPConfig.Port))
	case "HTTP", "HTTPS":
		d.Set("method", healthcheck.HTTPConfig.Method)
		d.Set("port", int(healthcheck.HTTPConfig.Port))
		d.Set("path", healthcheck.HTTPConfig.Path)
		d.Set("expected_codes", healthcheck.HTTPConfig.ExpectedCodes)
		d.Set("expected_body", healthcheck.HTTPConfig.ExpectedBody)
		d.Set("follow_redirects", healthcheck.HTTPConfig.FollowRedirects)
		d.Set("allow_insecure", healthcheck.HTTPConfig.AllowInsecure)
		d.Set("notification_email_addresses", healthcheck.Notification.EmailAddresses)

		if err := d.Set("header", flattenHealthcheckHeader(healthcheck.HTTPConfig.Header)); err != nil {
			log.Printf("[WARN] Error setting header for standalone healthcheck %q: %s", d.Id(), err)
		}
	}

	d.Set("name", healthcheck.Name)
	d.Set("description", healthcheck.Description)
	d.Set("suspended", healthcheck.Suspended)
	d.Set("notification_suspended", healthcheck.Notification.Suspended)
	d.Set("address", healthcheck.Address)
	d.Set("consecutive_fails", healthcheck.ConsecutiveFails)
	d.Set("consecutive_successes", healthcheck.ConsecutiveSuccesses)
	d.Set("retries", healthcheck.Retries)
	d.Set("timeout", healthcheck.Timeout)
	d.Set("interval", healthcheck.Interval)
	d.Set("type", healthcheck.Type)
	d.Set("created_on", healthcheck.CreatedOn.Format(time.RFC3339Nano))
	d.Set("modified_on", healthcheck.ModifiedOn.Format(time.RFC3339Nano))
	d.Set("check_regions", healthcheck.CheckRegions)

	return nil
}

func resourceCloudflareHealthcheckCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	healthcheck, err := healthcheckSetStruct(d)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error creating healthcheck struct"))
	}

	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		hc, err := client.CreateHealthcheck(context.Background(), zoneID, healthcheck)
		if err != nil {
			if strings.Contains(err.Error(), "no such host") {
				return resource.RetryableError(fmt.Errorf("hostname resolution failed"))
			}

			return resource.NonRetryableError(errors.Wrap(err, fmt.Sprintf("error creating standalone healthcheck")))
		}

		d.SetId(hc.ID)

		return resource.NonRetryableError(resourceCloudflareHealthcheckRead(d, meta))
	})
}

func resourceCloudflareHealthcheckUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	healthcheck, err := healthcheckSetStruct(d)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error creating healthcheck struct"))
	}

	_, err = client.UpdateHealthcheck(context.Background(), zoneID, d.Id(), healthcheck)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error creating healthcheck"))
	}

	return resourceCloudflareHealthcheckRead(d, meta)
}

func resourceCloudflareHealthcheckDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	err := client.DeleteHealthcheck(context.Background(), zoneID, d.Id())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error deleting standalone healthcheck"))
	}

	return nil
}

func resourceCloudflareHealthcheckImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/HealthcheckId\"", d.Id())
	}

	zoneID, HealthcheckID := attributes[0], attributes[1]

	d.Set("zone_id", zoneID)
	d.SetId(HealthcheckID)

	resourceCloudflareHealthcheckRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

func healthcheckSetStruct(d *schema.ResourceData) (cloudflare.Healthcheck, error) {
	healthcheck := cloudflare.Healthcheck{
		Name:                 d.Get("name").(string),
		Address:              d.Get("address").(string),
		Type:                 d.Get("type").(string),
		Retries:              d.Get("retries").(int),
		Timeout:              d.Get("timeout").(int),
		Interval:             d.Get("interval").(int),
		ConsecutiveFails:     d.Get("consecutive_fails").(int),
		ConsecutiveSuccesses: d.Get("consecutive_successes").(int),
	}

	if description, ok := d.GetOk("description"); ok {
		healthcheck.Description = description.(string)
	}

	if suspended, ok := d.GetOk("suspended"); ok {
		healthcheck.Suspended = suspended.(bool)
	}

	if region, ok := d.GetOk("check_regions"); ok {
		healthcheck.CheckRegions = expandInterfaceToStringList(region)
	}

	if notificationSuspended, ok := d.GetOk("notification_suspended"); ok {
		healthcheck.Notification.Suspended = notificationSuspended.(bool)
	}

	if notificationEmailAddresses, ok := d.GetOk("notification_email_addresses"); ok {
		healthcheck.Notification.EmailAddresses = expandInterfaceToStringList(notificationEmailAddresses)
	}

	switch healthcheck.Type {
	case "TCP":
		tcpConfig := new(cloudflare.HealthcheckTCPConfig)

		if method, ok := d.GetOk("method"); ok {
			if method != "connection_established" {
				return cloudflare.Healthcheck{}, errors.New(fmt.Sprintf("cannot use %s as method for TCP healthchecks", method))
			}
			tcpConfig.Method = method.(string)
		} else {
			tcpConfig.Method = "connection_established"
		}

		if port, ok := d.GetOk("port"); ok {
			tcpConfig.Port = uint16(port.(int))
		}

		healthcheck.TCPConfig = tcpConfig
	case "HTTP", "HTTPS":
		httpConfig := new(cloudflare.HealthcheckHTTPConfig)

		if method, ok := d.GetOk("method"); ok {
			if method != "GET" && method != "HEAD" {
				return cloudflare.Healthcheck{}, errors.New(fmt.Sprintf("cannot use %s as method for HTTP/HTTPS healthchecks", method))
			}
			httpConfig.Method = method.(string)
		} else {
			httpConfig.Method = "GET"
		}

		if port, ok := d.GetOk("port"); ok {
			httpConfig.Port = uint16(port.(int))
		}

		if path, ok := d.GetOk("path"); ok {
			httpConfig.Path = path.(string)
		}

		if expectedCode, ok := d.GetOk("expected_codes"); ok {
			httpConfig.ExpectedCodes = expandInterfaceToStringList(expectedCode)
		}

		if expectedBody, ok := d.GetOk("expected_body"); ok {
			httpConfig.ExpectedBody = expectedBody.(string)
		}

		if followRedirects, ok := d.GetOk("follow_redirects"); ok {
			httpConfig.FollowRedirects = followRedirects.(bool)
		}

		if allowInsecure, ok := d.GetOk("allow_insecure"); ok {
			httpConfig.AllowInsecure = allowInsecure.(bool)
		}

		if header, ok := d.GetOk("header"); ok {
			httpConfig.Header = expandHealthcheckHeader(header)
		}

		healthcheck.HTTPConfig = httpConfig
	}

	return healthcheck, nil
}

func flattenHealthcheckHeader(header map[string][]string) *schema.Set {
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

func expandHealthcheckHeader(cfgSet interface{}) map[string][]string {
	header := make(map[string][]string)
	cfgList := cfgSet.(*schema.Set).List()
	for _, item := range cfgList {
		cfg := item.(map[string]interface{})
		header[cfg["header"].(string)] = expandInterfaceToStringList(cfg["values"].(*schema.Set).List())
	}
	return header
}
