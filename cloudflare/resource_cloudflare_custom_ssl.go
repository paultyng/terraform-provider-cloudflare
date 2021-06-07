package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceCloudflareCustomSsl() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareCustomSslCreate,
		Read:   resourceCloudflareCustomSslRead,
		Update: resourceCloudflareCustomSslUpdate,
		Delete: resourceCloudflareCustomSslDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareCustomSslImport,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"custom_ssl_priority": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"custom_ssl_options": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate": {
							Type:     schema.TypeString,
							Optional: false,
						},
						"private_key": {
							Type:      schema.TypeString,
							Optional:  false,
							Sensitive: true,
						},
						"bundle_method": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"ubiquitous", "optimal", "force"}, false),
						},
						"geo_restrictions": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"us", "eu", "highest_security"}, false),
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"legacy_custom", "sni_custom"}, false),
						},
					},
				},
			},
			"hosts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"issuer": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signature": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uploaded_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"priority": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceCloudflareCustomSslCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	log.Printf("[DEBUG] zone ID: %s", zoneID)
	zcso, err := expandToZoneCustomSSLOptions(d)
	if err != nil {
		return fmt.Errorf("Failed to create custom ssl cert: %s", err)
	}

	res, err := client.CreateSSL(context.Background(), zoneID, zcso)
	if err != nil {
		return fmt.Errorf("Failed to create custom ssl cert: %s", err)
	}

	if res.ID == "" {
		return fmt.Errorf("Failed to find custom ssl in Create response: id was empty")
	}

	d.SetId(res.ID)

	log.Printf("[INFO] Cloudflare Custom SSL ID: %s", d.Id())

	return resourceCloudflareCustomSslRead(d, meta)
}

func resourceCloudflareCustomSslUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	certID := d.Id()
	var uErr error
	var reErr error
	var updateErr = false
	var reprioritizeErr = false
	log.Printf("[DEBUG] zone ID: %s", zoneID)

	if d.HasChange("custom_ssl_options") {
		zcso, err := expandToZoneCustomSSLOptions(d)
		if err != nil {
			return fmt.Errorf("Failed to update custom ssl cert: %s", err)
		}

		res, uErr := client.UpdateSSL(context.Background(), zoneID, certID, zcso)
		if uErr != nil {
			log.Printf("[DEBUG] Failed to update custom ssl cert: %s", uErr)
			updateErr = true
		} else {
			log.Printf("[DEBUG] Custom SSL set to: %s", res.ID)
		}

	}

	if d.HasChange("custom_ssl_priority") {
		zcsp, err := expandToZoneCustomSSLPriority(d)
		if err != nil {
			log.Printf("Failed to update custom ssl cert: %s", err)
		}

		resList, reErr := client.ReprioritizeSSL(context.Background(), zoneID, zcsp)
		if err != nil {
			log.Printf("Failed to update / reprioritize custom ssl cert: %s", reErr)
			reprioritizeErr = true
		} else {
			log.Printf("[DEBUG] Custom SSL reprioritized to: %#v", resList)
		}
	}

	if updateErr && reprioritizeErr {
		return fmt.Errorf("Failed to update and reprioritize custom ssl cert: %s, %s", uErr, reErr)
	}

	return resourceCloudflareCustomSslRead(d, meta)
}

func resourceCloudflareCustomSslRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	certID := d.Id()

	// update all possible schema attributes with fields from api response
	record, err := client.SSLDetails(context.Background(), zoneID, certID)
	if err != nil {
		log.Printf("[WARN] Removing record from state because it's not found in API")
		d.SetId("")
		return nil
	}
	zcso, err := expandToZoneCustomSSLOptions(d)
	if err != nil {
		log.Printf("[WARN] Problem setting zone options not read from state %s", err)
	}
	zcso.BundleMethod = record.BundleMethod
	customSslOpts := flattenCustomSSLOptions(zcso)

	// fill in fields that the api doesn't return
	data, dataOk := d.GetOk("custom_ssl_options")
	newData := make(map[string]string)
	if dataOk {
		for id, value := range data.(map[string]interface{}) {
			newValue := value.(string)
			newData[id] = newValue
		}
	}
	if val, ok := newData["%"]; ok {
		customSslOpts["%"] = val
	}
	if val, ok := newData["geo_restrictions"]; ok {
		customSslOpts["geo_restrictions"] = val
	}
	if val, ok := newData["type"]; ok {
		customSslOpts["type"] = val
	}

	d.SetId(record.ID)
	d.Set("hosts", record.Hosts)
	d.Set("issuer", record.Issuer)
	d.Set("signature", record.Signature)
	if err := d.Set("custom_ssl_options", customSslOpts); err != nil {
		return fmt.Errorf("[WARN] Error reading custom ssl opts %q: %s", d.Id(), err)
	}
	d.Set("status", record.Status)
	d.Set("uploaded_on", record.UploadedOn.Format(time.RFC3339Nano))
	d.Set("expires_on", record.ExpiresOn.Format(time.RFC3339Nano))
	d.Set("modified_on", record.ModifiedOn.Format(time.RFC3339Nano))
	d.Set("priority", record.Priority)
	return nil
}

func resourceCloudflareCustomSslDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	certID := d.Id()

	log.Printf("[DEBUG] Deleting SSL cert %s for zone %s", certID, zoneID)

	err := client.DeleteSSL(context.Background(), zoneID, certID)
	if err != nil {
		errors.Wrap(err, "failed to delete custom ssl cert setting")
	}
	return nil
}

func resourceCloudflareCustomSslImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// split the id so we can lookup
	idAttr := strings.SplitN(d.Id(), "/", 2)
	if len(idAttr) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/certID\"", d.Id())
	}

	zoneID, certID := idAttr[0], idAttr[1]

	log.Printf("[DEBUG] Importing Cloudflare Custom SSL Cert: id %s for zone %s", certID, zoneID)

	d.Set("zone_id", zoneID)
	d.SetId(certID)

	resourceCloudflareCustomSslRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

func expandToZoneCustomSSLPriority(d *schema.ResourceData) ([]cloudflare.ZoneCustomSSLPriority, error) {
	data, dataOk := d.GetOk("custom_ssl_priority")
	log.Printf("[DEBUG] Custom SSL priority found in config: %#v", data)
	var mtSlice []cloudflare.ZoneCustomSSLPriority
	if dataOk {
		for _, innerData := range data.([]interface{}) {
			newData := make(map[string]interface{})
			for id, value := range innerData.(map[string]interface{}) {
				switch idName := id; idName {
				case "id":
					newValue := value.(string)
					newData["ID"] = newValue
				case "priority":
					newValue := value.(int)
					newData[id] = newValue
				default:
					newValue := value
					newData[id] = newValue
				}
			}
			zcsp := cloudflare.ZoneCustomSSLPriority{}
			zcspJSON, err := json.Marshal(newData)
			if err != nil {
				return mtSlice, fmt.Errorf("Failed to create custom ssl priorities: %s", err)
			}
			// map -> json -> struct
			json.Unmarshal(zcspJSON, &zcsp)
			mtSlice = append(mtSlice, zcsp)
		}
	}
	log.Printf("[DEBUG] Custom SSL priority list creating: %#v", mtSlice)
	return mtSlice, nil
}

func expandToZoneCustomSSLOptions(d *schema.ResourceData) (cloudflare.ZoneCustomSSLOptions, error) {
	data, dataOk := d.GetOk("custom_ssl_options")
	log.Printf("[DEBUG] Custom SSL options found in config: %#v", data)

	newData := make(map[string]interface{})
	if dataOk {
		for id, value := range data.(map[string]interface{}) {
			var newValue interface{}
			if id == "geo_restrictions" {
				newValue = cloudflare.ZoneCustomSSLGeoRestrictions{
					Label: value.(string),
				}
			} else {
				newValue = value.(string)
			}
			newData[id] = newValue
		}
	}

	zcso := cloudflare.ZoneCustomSSLOptions{}
	zcsoJSON, err := json.Marshal(newData)
	if err != nil {
		return zcso, fmt.Errorf("Failed to create custom ssl options: %s", err)
	}

	log.Printf("[DEBUG] Custom SSL JSON: %s", string(zcsoJSON))

	// map -> json -> struct
	json.Unmarshal(zcsoJSON, &zcso)
	log.Printf("[DEBUG] Custom SSL options creating: %#v", zcso)
	return zcso, nil
}

func flattenCustomSSLOptions(sslopt cloudflare.ZoneCustomSSLOptions) map[string]interface{} {
	data := map[string]interface{}{
		"certificate":   sslopt.Certificate,
		"private_key":   sslopt.PrivateKey,
		"bundle_method": sslopt.BundleMethod,
	}

	if sslopt.GeoRestrictions != nil {
		data["geo_restrictions"] = sslopt.GeoRestrictions.Label
	}

	return data
}
