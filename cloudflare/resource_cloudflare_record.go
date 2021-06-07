package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceCloudflareRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareRecordCreate,
		Read:   resourceCloudflareRecordRead,
		Update: resourceCloudflareRecordUpdate,
		Delete: resourceCloudflareRecordDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareRecordImport,
		},

		SchemaVersion: 1,
		MigrateState:  resourceCloudflareRecordMigrateState,
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(i interface{}) string {
					return strings.ToLower(i.(string))
				},
			},

			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"A", "AAAA", "CAA", "CNAME", "TXT", "SRV", "LOC", "MX", "NS", "SPF", "CERT", "DNSKEY", "DS", "NAPTR", "SMIMEA", "SSHFP", "TLSA", "URI", "PTR"}, false),
			},

			"value": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"data"},
			},

			"data": {
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"value"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Properties present in several record types
						"algorithm": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"key_tag": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"flags": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"service": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"certificate": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"usage": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"selector": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"matching_type": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						// SRV record properties
						"proto": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"target": {
							Type:     schema.TypeString,
							Optional: true,
						},

						// LOC record properties
						"size": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"altitude": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"long_degrees": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"lat_degrees": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"precision_horz": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"precision_vert": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"long_direction": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"long_minutes": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"long_seconds": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"lat_direction": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"lat_minutes": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"lat_seconds": {
							Type:     schema.TypeFloat,
							Optional: true,
						},

						// DNSKEY record properties
						"protocol": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"public_key": {
							Type:     schema.TypeString,
							Optional: true,
						},

						// DS record properties
						"digest_type": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"digest": {
							Type:     schema.TypeString,
							Optional: true,
						},

						// NAPTR record properties
						"order": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"preference": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"regex": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"replacement": {
							Type:     schema.TypeString,
							Optional: true,
						},

						// SSHFP record properties
						"fingerprint": {
							Type:     schema.TypeString,
							Optional: true,
						},

						// URI record properties
						"content": {
							Type:     schema.TypeString,
							Optional: true,
						},

						// CAA record properties
						"tag": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"priority": {
				Type:             schema.TypeInt,
				Optional:         true,
				DiffSuppressFunc: suppressPriority,
			},

			"proxied": {
				Optional: true,
				Type:     schema.TypeBool,
			},

			"created_on": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"metadata": {
				Type:     schema.TypeMap,
				Computed: true,
			},

			"modified_on": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"proxiable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

func resourceCloudflareRecordCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	newRecord := cloudflare.DNSRecord{
		Type:   d.Get("type").(string),
		Name:   d.Get("name").(string),
		ZoneID: d.Get("zone_id").(string),
	}

	proxied, proxiedOk := d.GetOkExists("proxied")
	if proxiedOk {
		newRecord.Proxied = &[]bool{proxied.(bool)}[0]
	}

	value, valueOk := d.GetOk("value")
	if valueOk {
		newRecord.Content = value.(string)
	}

	data, dataOk := d.GetOk("data")
	log.Printf("[DEBUG] Data found in config: %#v", data)

	newDataMap := make(map[string]interface{})

	if dataOk {
		for id, value := range data.(map[string]interface{}) {
			newData, err := transformToCloudflareDNSData(newRecord.Type, id, value)
			if err != nil {
				return err
			} else if newData == nil {
				continue
			}
			newDataMap[id] = newData
		}

		newRecord.Data = newDataMap
	}

	if valueOk == dataOk {
		return fmt.Errorf(
			"either 'value' (present: %t) or 'data' (present: %t) must be provided",
			valueOk, dataOk)
	}

	if priority, ok := d.GetOkExists("priority"); ok {
		p := uint16(priority.(int))
		newRecord.Priority = &p
	}

	if ttl, ok := d.GetOk("ttl"); ok {
		if ttl.(int) != 1 && proxiedOk && *newRecord.Proxied {
			return fmt.Errorf("error validating record %s: ttl must be set to 1 when `proxied` is true", newRecord.Name)
		}

		newRecord.TTL = ttl.(int)
	}

	// Validate value based on type
	if err := validateRecordName(newRecord.Type, newRecord.Content); err != nil {
		return fmt.Errorf("Error validating record name %q: %s", newRecord.Name, err)
	}

	var proxiedVal *bool
	if proxiedOk {
		proxiedVal = newRecord.Proxied
	} else {
		proxiedVal = &[]bool{false}[0]
	}

	// Validate type
	if err := validateRecordType(newRecord.Type, *proxiedVal); err != nil {
		return fmt.Errorf("Error validating record type %q: %s", newRecord.Type, err)
	}

	log.Printf("[DEBUG] Cloudflare Record create configuration: %#v", newRecord)

	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateDNSRecord(context.Background(), newRecord.ZoneID, newRecord)
		if err != nil {
			if strings.Contains(err.Error(), "already exist") {
				return resource.RetryableError(fmt.Errorf("expected DNS record to not already be present but already exists"))
			}

			return resource.NonRetryableError(fmt.Errorf("failed to create DNS record: %s", err))
		}

		// In the event that the API returns an empty DNS Record, we verify that the
		// ID returned is not the default ""
		if r.Result.ID == "" {
			return resource.NonRetryableError(fmt.Errorf("Failed to find record in Create response; Record was empty"))
		}

		d.SetId(r.Result.ID)

		log.Printf("[INFO] Cloudflare Record ID: %s", d.Id())

		return resource.NonRetryableError(resourceCloudflareRecordRead(d, meta))
	})
}

func resourceCloudflareRecordRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	record, err := client.DNSRecord(context.Background(), zoneID, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "Invalid dns record identifier") ||
			strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[WARN] Removing record from state because it's not found in API")
			d.SetId("")
			return nil
		}
		return err
	}

	data, dataOk := d.GetOk("data")
	log.Printf("[DEBUG] Data found in config: %#v", data)

	readDataMap := make(map[string]interface{})

	if dataOk {
		for id, value := range data.(map[string]interface{}) {
			newData, err := transformToCloudflareDNSData(record.Type, id, value)
			if err != nil {
				return err
			} else if newData == nil {
				continue
			}
			readDataMap[id] = newData
		}

		record.Data = readDataMap
	}

	d.SetId(record.ID)
	d.Set("hostname", record.Name)
	d.Set("type", record.Type)
	d.Set("value", record.Content)
	d.Set("ttl", record.TTL)
	d.Set("proxied", record.Proxied)
	d.Set("created_on", record.CreatedOn.Format(time.RFC3339Nano))
	d.Set("data", expandStringMap(record.Data))
	d.Set("modified_on", record.ModifiedOn.Format(time.RFC3339Nano))
	if err := d.Set("metadata", expandStringMap(record.Meta)); err != nil {
		log.Printf("[WARN] Error setting metadata: %s", err)
	}
	d.Set("proxiable", record.Proxiable)

	if record.Priority != nil {
		priority := record.Priority
		p := *priority
		d.Set("priority", fmt.Sprintf("%d", int(p)))
	}

	return nil
}

func resourceCloudflareRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	updateRecord := cloudflare.DNSRecord{
		ID:      d.Id(),
		Type:    d.Get("type").(string),
		Name:    d.Get("name").(string),
		Content: d.Get("value").(string),
		ZoneID:  zoneID,
	}

	data, dataOk := d.GetOk("data")
	log.Printf("[DEBUG] Data found in config: %#v", data)

	newDataMap := make(map[string]interface{})

	if dataOk {
		for id, value := range data.(map[string]interface{}) {
			newData, err := transformToCloudflareDNSData(updateRecord.Type, id, value)
			if err != nil {
				return err
			} else if newData == nil {
				continue
			}
			newDataMap[id] = newData
		}

		updateRecord.Data = newDataMap
	}

	if priority, ok := d.GetOkExists("priority"); ok {
		p := uint16(priority.(int))
		updateRecord.Priority = &p
	}

	proxied, proxiedOk := d.GetOkExists("proxied")
	if proxiedOk {
		updateRecord.Proxied = &[]bool{proxied.(bool)}[0]
	}

	if ttl, ok := d.GetOk("ttl"); ok {
		if ttl.(int) != 1 && proxiedOk && *updateRecord.Proxied {
			return fmt.Errorf("error validating record %s: ttl must be set to 1 when `proxied` is true", updateRecord.Name)
		}

		updateRecord.TTL = ttl.(int)
	}

	log.Printf("[DEBUG] Cloudflare Record update configuration: %#v", updateRecord)

	return resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		err := client.UpdateDNSRecord(context.Background(), zoneID, d.Id(), updateRecord)
		if err != nil {
			if strings.Contains(err.Error(), "already exist") {
				return resource.RetryableError(fmt.Errorf("expected DNS record to not already be present but already exists"))
			}

			return resource.NonRetryableError(fmt.Errorf("failed to create DNS record: %s", err))
		}

		return resource.NonRetryableError(resourceCloudflareRecordRead(d, meta))
	})
}

func resourceCloudflareRecordDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	log.Printf("[INFO] Deleting Cloudflare Record: %s, %s", zoneID, d.Id())

	err := client.DeleteDNSRecord(context.Background(), zoneID, d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Cloudflare Record: %s", err)
	}

	return nil
}

func expandStringMap(inVal interface{}) map[string]string {
	// although interface could hold anything
	// we assume that it is either nil or a map of interface values
	outVal := make(map[string]string)
	if inVal == nil {
		return outVal
	}
	for k, v := range inVal.(map[string]interface{}) {
		strValue := fmt.Sprintf("%v", v)
		outVal[k] = strValue
	}
	return outVal
}

func resourceCloudflareRecordImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*cloudflare.API)

	// split the id so we can lookup
	idAttr := strings.SplitN(d.Id(), "/", 2)
	var zoneID string
	var recordID string
	if len(idAttr) == 2 {
		zoneID = idAttr[0]
		recordID = idAttr[1]
	} else {
		return nil, fmt.Errorf("invalid id %q specified, should be in format \"zoneID/recordID\" for import", d.Id())
	}

	record, err := client.DNSRecord(context.Background(), zoneID, recordID)
	if err != nil {
		return nil, fmt.Errorf("Unable to find record with ID %q: %q", d.Id(), err)
	}

	log.Printf("[INFO] Found record: %s", record.Name)
	name := strings.TrimSuffix(record.Name, "."+record.ZoneName)

	d.Set("name", name)
	d.Set("zone_id", zoneID)
	d.SetId(recordID)

	resourceCloudflareZoneRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

var dnsTypeIntFields = []string{
	"algorithm",
	"key_tag",
	"type",
	"usage",
	"selector",
	"matching_type",
	"weight",
	"priority",
	"port",
	"long_degrees",
	"lat_degrees",
	"long_minutes",
	"lat_minutes",
	"protocol",
	"digest_type",
	"order",
	"preference",
}

var dnsTypeFloatFields = []string{
	"size",
	"altitude",
	"precision_horz",
	"precision_vert",
	"long_seconds",
	"lat_seconds",
}

func transformToCloudflareDNSData(recordType string, id string, value interface{}) (newValue interface{}, err error) {
	switch {
	case id == "flags":
		switch {
		case strings.ToUpper(recordType) == "SRV",
			strings.ToUpper(recordType) == "CAA",
			strings.ToUpper(recordType) == "DNSKEY":
			newValue, err = strconv.Atoi(value.(string))
		case strings.ToUpper(recordType) == "NAPTR":
			newValue, err = value.(string), nil
		}
	case contains(dnsTypeIntFields, id):
		newValue, err = strconv.Atoi(value.(string))
	case contains(dnsTypeFloatFields, id):
		newValue, err = strconv.ParseFloat(value.(string), 32)
	default:
		newValue, err = value.(string), nil
	}

	return
}

func suppressPriority(k, old, new string, d *schema.ResourceData) bool {
	recordType := d.Get("type").(string)
	if recordType != "MX" && recordType != "URI" {
		return true
	}
	return false
}
