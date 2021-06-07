package cloudflare

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceCloudflareAccessRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareAccessRuleCreate,
		Read:   resourceCloudflareAccessRuleRead,
		Update: resourceCloudflareAccessRuleUpdate,
		Delete: resourceCloudflareAccessRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareAccessRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"block", "challenge", "whitelist", "js_challenge"}, false),
			},
			"notes": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"configuration": {
				Type:             schema.TypeMap,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: configurationDiffSuppress,
				ValidateFunc:     validateAccessRuleConfiguration,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"ip", "ip6", "ip_range", "asn", "country"}, false),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceCloudflareAccessRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	newRule := cloudflare.AccessRule{
		Notes: d.Get("notes").(string),
		Mode:  d.Get("mode").(string),
	}

	if configuration, configurationOk := d.GetOk("configuration"); configurationOk {
		config := configuration.(map[string]interface{})

		newRule.Configuration = cloudflare.AccessRuleConfiguration{
			Target: config["target"].(string),
			Value:  config["value"].(string),
		}
	}

	var r *cloudflare.AccessRuleResponse
	var err error

	if zoneID == "" {
		if client.AccountID != "" {
			r, err = client.CreateAccountAccessRule(context.Background(), client.AccountID, newRule)
		} else {
			r, err = client.CreateUserAccessRule(context.Background(), newRule)
		}
	} else {
		r, err = client.CreateZoneAccessRule(context.Background(), zoneID, newRule)
	}

	if err != nil {
		return fmt.Errorf("Failed to create access rule: %s", err)
	}

	if r.Result.ID == "" {
		return fmt.Errorf("Failed to find access rule in Create response; ID was empty")
	}

	d.SetId(r.Result.ID)

	return resourceCloudflareAccessRuleRead(d, meta)
}

func resourceCloudflareAccessRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	var accessRuleResponse *cloudflare.AccessRuleResponse
	var err error

	if zoneID == "" {
		if client.AccountID != "" {
			accessRuleResponse, err = client.AccountAccessRule(context.Background(), client.AccountID, d.Id())
		} else {
			accessRuleResponse, err = client.UserAccessRule(context.Background(), d.Id())
		}
	} else {
		accessRuleResponse, err = client.ZoneAccessRule(context.Background(), zoneID, d.Id())
	}

	log.Printf("[DEBUG] accessRuleResponse: %#v", accessRuleResponse)
	log.Printf("[DEBUG] accessRuleResponse error: %#v", err)

	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Access Rule %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error finding access rule %q: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Cloudflare Access Rule read configuration: %#v", accessRuleResponse)

	d.Set("zone_id", zoneID)
	d.Set("mode", accessRuleResponse.Result.Mode)
	d.Set("notes", accessRuleResponse.Result.Notes)
	log.Printf("[DEBUG] read configuration: %#v", d.Get("configuration"))

	configuration := map[string]interface{}{}
	configuration["target"] = accessRuleResponse.Result.Configuration.Target
	configuration["value"] = accessRuleResponse.Result.Configuration.Value

	d.Set("configuration", configuration)

	return nil
}

func resourceCloudflareAccessRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	newRule := cloudflare.AccessRule{
		Notes: d.Get("notes").(string),
		Mode:  d.Get("mode").(string),
	}

	if configuration, configurationOk := d.GetOk("configuration"); configurationOk {
		config := configuration.(map[string]interface{})

		newRule.Configuration = cloudflare.AccessRuleConfiguration{
			Target: config["target"].(string),
			Value:  config["value"].(string),
		}
	}

	// var accessRuleResponse *cloudflare.AccessRuleResponse
	var err error

	if zoneID == "" {
		if client.AccountID != "" {
			_, err = client.UpdateAccountAccessRule(context.Background(), client.AccountID, d.Id(), newRule)
		} else {
			_, err = client.UpdateUserAccessRule(context.Background(), d.Id(), newRule)
		}
	} else {
		_, err = client.UpdateZoneAccessRule(context.Background(), zoneID, d.Id(), newRule)
	}

	if err != nil {
		return fmt.Errorf("Failed to update Access Rule: %s", err)
	}

	return resourceCloudflareAccessRuleRead(d, meta)
}

func resourceCloudflareAccessRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	log.Printf("[INFO] Deleting Cloudflare Access Rule: id %s for zone_id %s", d.Id(), zoneID)

	var err error

	if zoneID == "" {
		if client.AccountID != "" {
			_, err = client.DeleteAccountAccessRule(context.Background(), client.AccountID, d.Id())
		} else {
			_, err = client.DeleteUserAccessRule(context.Background(), d.Id())
		}
	} else {
		_, err = client.DeleteZoneAccessRule(context.Background(), zoneID, d.Id())
	}

	if err != nil {
		return fmt.Errorf("Error deleting Cloudflare Access Rule: %s", err)
	}

	return nil
}

func resourceCloudflareAccessRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*cloudflare.API)
	attributes := strings.Split(d.Id(), "/")

	var (
		accessRuleType           string
		accessRuleTypeIdentifier string
		accessRuleID             string
	)

	if len(attributes) != 3 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"accessRuleType/accessRuleTypeIdentifier/identiferValue\"", d.Id())
	}

	accessRuleType, accessRuleTypeIdentifier, accessRuleID = attributes[0], attributes[1], attributes[2]

	d.SetId(accessRuleID)

	switch accessRuleType {
	case "account":
		client.AccountID = accessRuleTypeIdentifier
	case "zone":
		d.Set("zone_id", accessRuleTypeIdentifier)
	}

	resourceCloudflareAccessRuleRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

func configurationDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	switch {
	case d.Get("configuration.target") == "country" &&
		k == "configuration.value":
		return strings.ToUpper(old) == strings.ToUpper(new)
	case d.Get("configuration.target") == "asn" &&
		k == "configuration.value":

		if !strings.HasPrefix(strings.ToUpper(new), "AS") {
			new = "AS" + strings.ToUpper(new)
		}

		return strings.ToUpper(old) == strings.ToUpper(new)
	}

	return false
}

func validateAccessRuleConfiguration(v interface{}, k string) (warnings []string, errors []error) {
	config := v.(map[string]interface{})

	target := config["target"].(string)
	value := config["value"].(string)

	switch target {
	case "ip_range":
		return validateAccessRuleConfigurationIPRange(value)
	default:
	}

	return warnings, errors
}

func validateAccessRuleConfigurationIPRange(v string) (warnings []string, errors []error) {
	ip, ipNet, err := net.ParseCIDR(v)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to parse value as CIDR: %v", err))
		return warnings, errors
	}

	if ipNet == nil {
		errors = append(errors, fmt.Errorf("ip_range must hold a range"))
		return warnings, errors
	}

	if ip.To4() != nil {
		ones, _ := ipNet.Mask.Size()
		if ones != 16 && ones != 24 {
			errors = append(errors, fmt.Errorf("ip_range with ipv4 address must be a /16 or /24, got a /%d", ones))
			return warnings, errors
		}
	} else {
		ones, _ := ipNet.Mask.Size()
		if ones != 32 && ones != 48 && ones != 64 {
			errors = append(errors, fmt.Errorf("ip_range with ipv6 address must be in (/32, /48, /64), instead got a /%d", ones))
			return warnings, errors
		}
	}

	return warnings, errors
}
