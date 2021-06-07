package cloudflare

import (
	"context"
	"fmt"
	"log"

	"strings"

	"time"

	"reflect"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceCloudflareZoneSettingsOverride() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareZoneSettingsOverrideCreate,
		Read:   resourceCloudflareZoneSettingsOverrideRead,
		Update: resourceCloudflareZoneSettingsOverrideUpdate,
		Delete: resourceCloudflareZoneSettingsOverrideDelete,

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"settings": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: resourceCloudflareZoneSettingsSchema,
				},
			},

			"initial_settings": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: resourceCloudflareZoneSettingsSchema,
				},
			},

			"initial_settings_read_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"readonly_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"zone_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"zone_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

var resourceCloudflareZoneSettingsSchema = map[string]*schema.Schema{
	"always_online": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"brotli": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"browser_cache_ttl": {
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
		ValidateFunc: validation.IntInSlice([]int{0, 30, 60, 300, 1200, 1800, 3600, 7200, 10800, 14400, 18000, 28800,
			43200, 57600, 72000, 86400, 172800, 259200, 345600, 432000, 691200, 1382400, 2073600, 2678400, 5356800,
			16070400, 31536000}),
		// minimum TTL available depends on the plan level of the zone.
		// - Respect existing headers = 0
		// - Enterprise = 30
		// - Business, Pro, Free = 1800
	},

	"browser_check": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"cache_level": {
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"aggressive", "basic", "simplified"}, false),
	},

	"challenge_ttl": {
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
		ValidateFunc: validation.IntInSlice([]int{300, 900, 1800, 2700, 3600, 7200, 10800, 14400, 28800, 57600,
			86400, 604800, 2592000, 31536000}),
	},

	"development_mode": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"origin_error_page_pass_thru": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"sort_query_string_for_cache": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"email_obfuscation": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"hotlink_protection": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"ip_geolocation": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"ipv6": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"websockets": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"minify": {
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"css": {
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
					Required:     true,
				},

				"html": {
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
					Required:     true,
				},

				"js": {
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
					Required:     true,
				},
			},
		},
	},

	"mobile_redirect": {
		Type:     schema.TypeList, // on/off
		Optional: true,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// which parameters are mandatory is not specified
				"mobile_subdomain": {
					Type:     schema.TypeString,
					Required: true,
				},

				"strip_uri": {
					Type:     schema.TypeBool,
					Required: true,
				},

				"status": {
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
					Required:     true,
				},
			},
		},
	},

	"mirage": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"opportunistic_encryption": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"opportunistic_onion": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"polish": {
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"off", "lossless", "lossy"}, false),
	},

	"webp": {
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
	},

	"prefetch_preload": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"privacy_pass": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"response_buffering": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"rocket_loader": {
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"on", "off", "manual"}, false),
	},

	"security_header": {
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},

				"preload": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},

				"max_age": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},

				"include_subdomains": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},

				"nosniff": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},
			},
		},
	},

	"security_level": {
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"off", "essentially_off", "low", "medium", "high", "under_attack"}, false),
	},

	"server_side_exclude": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"ssl": {
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"off", "flexible", "full", "strict", "origin_pull"}, false), // depends on plan
	},

	"universal_ssl": {
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
	},

	"tls_client_auth": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"true_client_ip_header": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"waf": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"min_tls_version": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"1.0", "1.1", "1.2", "1.3"}, false),
		Optional:     true,
		Computed:     true,
	},

	"tls_1_2_only": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
		Deprecated:   "tls_1_2_only has been deprecated in favour of using `min_tls_version = \"1.2\"` instead.",
	},

	"tls_1_3": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off", "zrt"}, false),
		Optional:     true,
		Computed:     true, // default depends on plan level
	},

	"automatic_https_rewrites": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"http2": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"http3": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},

	"pseudo_ipv4": {
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"off", "add_header", "overwrite_header"}, false),
	},

	"always_use_https": {
		// may cause an error: HTTP status 400: content "{\"success\":false,\"errors\":[{\"code\":1016,\"message\":\"An unknown error has occurred\"}],\"messages\":[],\"result\":null}"
		// but it still gets set at the API
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},

	"cname_flattening": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"flatten_at_root", "flatten_all", "flatten_none"}, false),
		Optional:     true,
		Computed:     true,
	},

	"max_upload": {
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
	},

	"h2_prioritization": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off", "custom"}, false),
		Optional:     true,
		Computed:     true,
	},

	"image_resizing": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off", "open"}, false),
		Optional:     true,
		Computed:     true,
	},

	"zero_rtt": {
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
		Optional:     true,
		Computed:     true,
	},
}

var fetchAsSingleSetting = []string{
	"binary_ast",
	"h2_prioritization",
	"image_resizing",
}

func resourceCloudflareZoneSettingsOverrideCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	zoneID := d.Get("zone_id").(string)
	d.SetId(zoneID)

	log.Printf("[INFO] Creating zone settings resource for zone ID: %s", d.Id())

	// do extra initial read to get initial_settings before updating
	zoneSettings, err := client.ZoneSettings(context.Background(), d.Id())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error reading initial settings for zone %q", d.Id()))
	}

	if err = updateZoneSettingsResponseWithSingleZoneSettings(zoneSettings, d.Id(), client); err != nil {
		return err
	}

	if _, ok := d.GetOk("settings.0.universal_ssl"); ok {
		// pulling USSL status and wrapping it into a cloudflare.ZoneSetting that we can set initial_settings
		if err = updateZoneSettingsResponseWithUniversalSSLSettings(zoneSettings, d.Id(), client); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Read CloudflareZone initial settings: %#v", zoneSettings)

	if err := d.Set("initial_settings", flattenZoneSettings(d, zoneSettings.Result, true)); err != nil {
		log.Printf("[WARN] Error setting initial_settings for zone %q: %s", d.Id(), err)
	}

	d.Set("initial_settings_read_at", time.Now().UTC().Format(time.RFC3339Nano))

	// set readonly setting so that update can behave correctly
	if err := d.Set("readonly_settings", flattenReadOnlyZoneSettings(zoneSettings.Result)); err != nil {
		log.Printf("[WARN] Error setting readonly_settings for zone %q: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Saved CloudflareZone initial settings: %#v", d.Get("initial_settings"))

	return resourceCloudflareZoneSettingsOverrideUpdate(d, meta)
}

func updateZoneSettingsResponseWithSingleZoneSettings(zoneSettings *cloudflare.ZoneSettingResponse, zoneId string, client *cloudflare.API) error {
	for _, settingName := range fetchAsSingleSetting {
		singleSetting, err := client.ZoneSingleSetting(context.Background(), zoneId, settingName)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error reading setting '%q' for zone %q", settingName, zoneId))
		}
		zoneSettings.Result = append(zoneSettings.Result, singleSetting)
	}
	return nil
}

func updateZoneSettingsResponseWithUniversalSSLSettings(zoneSettings *cloudflare.ZoneSettingResponse, zoneId string, client *cloudflare.API) error {
	ussl, err := client.UniversalSSLSettingDetails(context.Background(), zoneId)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error reading initial Universal SSL settings for zone %q", zoneId))
	}

	usslToZoneSetting := cloudflare.ZoneSetting{
		ID:       "universal_ssl",
		Value:    stringFromBool(ussl.Enabled),
		Editable: true,
	}

	zoneSettings.Result = append(zoneSettings.Result, usslToZoneSetting)

	return nil
}

func resourceCloudflareZoneSettingsOverrideRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	zone, err := client.ZoneDetails(context.Background(), d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Zone %q not found", d.Id())
			d.SetId("")
			return nil
		} else {
			return errors.Wrap(err, fmt.Sprintf("Error reading zone %q", d.Id()))
		}
	}

	d.Set("zone_id", d.Id())

	// not all settings are visible to all users, so this might be a subset
	// assume (for now) that user can see/do everything
	zoneSettings, err := client.ZoneSettings(context.Background(), d.Id())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error reading settings for zone %q", d.Id()))
	}

	if err = updateZoneSettingsResponseWithSingleZoneSettings(zoneSettings, d.Id(), client); err != nil {
		return err
	}

	if _, ok := d.GetOk("settings.0.universal_ssl"); ok {
		if err = updateZoneSettingsResponseWithUniversalSSLSettings(zoneSettings, d.Id(), client); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Read CloudflareZone Settings: %#v", zoneSettings)

	d.Set("status", zone.Status)
	d.Set("type", zone.Type)

	newZoneSettings := flattenZoneSettings(d, zoneSettings.Result, false)
	// if polish is off (or we don't know) we need to ignore what comes back from the api for webp
	if polish, ok := newZoneSettings[0]["polish"]; !ok || polish.(string) == "" || polish.(string) == "off" {
		newZoneSettings[0]["webp"] = d.Get("settings.0.webp").(string)
	}

	if err := d.Set("settings", newZoneSettings); err != nil {
		log.Printf("[WARN] Error setting settings for zone %q: %s", d.Id(), err)
	}

	if err := d.Set("readonly_settings", flattenReadOnlyZoneSettings(zoneSettings.Result)); err != nil {
		log.Printf("[WARN] Error setting readonly_settings for zone %q: %s", d.Id(), err)
	}

	return nil
}

func flattenZoneSettings(d *schema.ResourceData, settings []cloudflare.ZoneSetting, flattenAll bool) []map[string]interface{} {
	cfg := map[string]interface{}{}
	for _, s := range settings {
		if s.ID == "0rtt" { // NOTE: 0rtt is an invalid attribute in HCLs grammar.  Remap to `zero_rtt`
			s.ID = "zero_rtt"
		}

		if !settingInSchema(s.ID) {
			log.Printf("[WARN] Value not in schema returned from API zone settings (is it new?) - %q : %#v", s.ID, s.Value)
			continue
		}
		if _, ok := d.GetOkExists(fmt.Sprintf("settings.0.%s", s.ID)); !ok && !flattenAll {
			// don't put settings that were never specified in the update request
			continue
		}

		if s.ID == "minify" || s.ID == "mobile_redirect" {
			cfg[s.ID] = []interface{}{s.Value.(map[string]interface{})}
		} else if s.ID == "security_header" {
			cfg[s.ID] = []interface{}{s.Value.(map[string]interface{})["strict_transport_security"]}
		} else if strValue, ok := s.Value.(string); ok {
			cfg[s.ID] = strValue
		} else if floatValue, ok := s.Value.(float64); ok {
			cfg[s.ID] = int(floatValue)
		} else {
			log.Printf("[WARN] Unexpected value type found in API zone settings - %q : %#v", s.ID, s.Value)
		}
	}

	log.Printf("[DEBUG] Flattened Cloudflare Zone Settings: %#v", cfg)

	return []map[string]interface{}{cfg}
}

func settingInSchema(val string) bool {
	for k, _ := range resourceCloudflareZoneSettingsSchema {
		if val == k {
			return true
		}
	}
	return false
}

func flattenReadOnlyZoneSettings(settings []cloudflare.ZoneSetting) []string {
	ids := make([]string, 0)
	for _, zs := range settings {
		if !zs.Editable {
			ids = append(ids, zs.ID)
		}
	}
	log.Printf("[DEBUG] Flattened Cloudflare Read Only Zone Settings: %#v", ids)

	return ids
}

func updateSingleZoneSettings(zoneSettings []cloudflare.ZoneSetting, client *cloudflare.API, zoneID string) ([]cloudflare.ZoneSetting, error) {
	var indexesToCut []int
	for i, setting := range zoneSettings {
		if contains(fetchAsSingleSetting, setting.ID) {
			_, err := client.UpdateZoneSingleSetting(context.Background(), zoneID, setting.ID, setting)
			if err != nil {
				return zoneSettings, err
			}
			indexesToCut = append(indexesToCut, i)
		}
	}

	for _, indexToCut := range indexesToCut {
		zoneSettings = append(zoneSettings[:indexToCut], zoneSettings[indexToCut+1:]...)
	}
	return zoneSettings, nil
}

func updateUniversalSSLSetting(zoneSettings []cloudflare.ZoneSetting, client *cloudflare.API, zoneID string) ([]cloudflare.ZoneSetting, error) {
	indexToCut := -1
	for i, setting := range zoneSettings {
		// Skipping USSL Update if value is empty, especially when we are reverting to the initial state and we did not had the information
		if setting.ID == "universal_ssl" {
			if setting.Value.(string) != "" {
				_, err := client.EditUniversalSSLSetting(context.Background(), zoneID, cloudflare.UniversalSSLSetting{Enabled: boolFromString(setting.Value.(string))})
				if err != nil {
					return zoneSettings, err
				}
			}
			indexToCut = i
		}
	}

	if indexToCut != -1 {
		zoneSettings = append(zoneSettings[:indexToCut], zoneSettings[indexToCut+1:]...)
	}

	return zoneSettings, nil
}

func resourceCloudflareZoneSettingsOverrideUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	if cfg, ok := d.GetOkExists("settings"); ok && cfg != nil && len(cfg.([]interface{})) > 0 {

		readOnlySettings := expandInterfaceToStringList(d.Get("readonly_settings"))
		zoneSettings, err := expandOverriddenZoneSettings(d, "settings", readOnlySettings)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Cloudflare Zone Settings update configuration: %#v", zoneSettings)

		if zoneSettings, err = updateSingleZoneSettings(zoneSettings, client, d.Id()); err != nil {
			return err
		}

		if zoneSettings, err = updateUniversalSSLSetting(zoneSettings, client, d.Id()); err != nil {
			return err
		}

		if len(zoneSettings) > 0 {
			_, err = client.UpdateZoneSettings(context.Background(), d.Id(), zoneSettings)
			if err != nil {
				return err
			}
		} else {
			log.Printf("[DEBUG] Skipped update call because no settings were set")
		}
	}

	return resourceCloudflareZoneSettingsOverrideRead(d, meta)
}

func expandOverriddenZoneSettings(d *schema.ResourceData, settingsKey string, readOnlySettings []string) ([]cloudflare.ZoneSetting, error) {
	zoneSettings := make([]cloudflare.ZoneSetting, 0)

	keyFormat := fmt.Sprintf("%s.0.%%s", settingsKey)

	for k, _ := range resourceCloudflareZoneSettingsSchema {

		// we only update if the user set the value non-empty before, and its different from the read value
		// note that if user removes an attribute, we don't do anything
		if settingValue, ok := d.GetOkExists(fmt.Sprintf(keyFormat, k)); ok && d.HasChange(fmt.Sprintf(keyFormat, k)) {

			zoneSettingValue, err := expandZoneSetting(d, keyFormat, k, settingValue, readOnlySettings)
			if err != nil {
				return zoneSettings, err
			}

			// Remap zero_rtt key back to Cloudflare's setting name, 0rtt
			if k == "zero_rtt" {
				k = "0rtt"
			}

			if zoneSettingValue != nil {
				newZoneSetting := cloudflare.ZoneSetting{
					ID:    k,
					Value: zoneSettingValue,
				}
				zoneSettings = append(zoneSettings, newZoneSetting)
			}

		}

	}
	return zoneSettings, nil
}

func expandZoneSetting(d *schema.ResourceData, keyFormatString, k string, settingValue interface{}, readOnlySettings []string) (interface{}, error) {

	if contains(readOnlySettings, k) {
		return nil, fmt.Errorf("invalid zone setting %q (value: %v) found - cannot be set as it is read only", k, settingValue)
	}

	var zoneSettingValue interface{}
	switch k {
	case "webp":
		{
			// only ever set webp if polish is on
			polishKey := fmt.Sprintf(keyFormatString, "polish")
			polish := d.Get(polishKey).(string)

			if polish != "" && polish != "off" {
				zoneSettingValue = settingValue
			}
		}
	case "minify", "mobile_redirect":
		{
			listValue := settingValue.([]interface{})
			if len(listValue) > 0 && listValue != nil {
				zoneSettingValue = listValue[0].(map[string]interface{})
			}

		}
	case "security_header":
		{
			listValue := settingValue.([]interface{})
			if len(listValue) > 0 && listValue != nil {
				zoneSettingValue = map[string]interface{}{
					"strict_transport_security": listValue[0].(map[string]interface{}),
				}
			}
		}
	default:
		{
			zoneSettingValue = settingValue
		}
	}
	return zoneSettingValue, nil
}

func resourceCloudflareZoneSettingsOverrideDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	if cfg, ok := d.GetOkExists("settings"); ok && cfg != nil && len(cfg.([]interface{})) > 0 {

		readOnlySettings := expandInterfaceToStringList(d.Get("readonly_settings"))

		zoneSettings, err := expandRevertibleZoneSettings(d, readOnlySettings)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Reverting Cloudflare Zone Settings to initial settings with update configuration: %#v", zoneSettings)

		if zoneSettings, err = updateSingleZoneSettings(zoneSettings, client, d.Id()); err != nil {
			return err
		}

		if zoneSettings, err = updateUniversalSSLSetting(zoneSettings, client, d.Id()); err != nil {
			return err
		}

		if len(zoneSettings) > 0 {
			_, err = client.UpdateZoneSettings(context.Background(), d.Id(), zoneSettings)
			if err != nil {
				return err
			}
		} else {
			log.Printf("[DEBUG] Skipped call to revert settings because no settings were changed")
		}
	}
	return nil
}

func expandRevertibleZoneSettings(d *schema.ResourceData, readOnlySettings []string) ([]cloudflare.ZoneSetting, error) {
	zoneSettings := make([]cloudflare.ZoneSetting, 0)

	keyFormat := fmt.Sprintf("%s.0.%%s", "initial_settings")

	for k, _ := range resourceCloudflareZoneSettingsSchema {

		initialKey := fmt.Sprintf("initial_settings.0.%s", k)
		initialVal := d.Get(initialKey)
		currentKey := fmt.Sprintf("settings.0.%s", k)

		// if the value was never set we don't need to revert it
		if currentVal, ok := d.GetOk(currentKey); ok && !schemaValueEquals(initialVal, currentVal) {

			zoneSettingValue, err := expandZoneSetting(d, keyFormat, k, initialVal, readOnlySettings)
			if err != nil {
				return zoneSettings, err
			}

			if zoneSettingValue != nil {
				newZoneSetting := cloudflare.ZoneSetting{
					ID:    k,
					Value: zoneSettingValue,
				}
				zoneSettings = append(zoneSettings, newZoneSetting)
			}

		}
	}
	return zoneSettings, nil
}

func schemaValueEquals(a, b interface{}) bool {
	// this is the same equality check used in d.HasChange

	// If the type implements the Equal interface, then call that
	// instead of just doing a reflect.DeepEqual. An example where this is
	// needed is *Set
	if eq, ok := a.(schema.Equal); ok {
		return eq.Equal(b)
	}

	return reflect.DeepEqual(a, b)
}
