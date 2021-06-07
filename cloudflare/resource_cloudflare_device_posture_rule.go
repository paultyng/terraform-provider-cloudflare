package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceCloudflareDevicePostureRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareDevicePostureRuleCreate,
		Read:   resourceCloudflareDevicePostureRuleRead,
		Update: resourceCloudflareDevicePostureRuleUpdate,
		Delete: resourceCloudflareDevicePostureRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareDevicePostureRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"serial_number", "file", "application"}, false),
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"schedule": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"match": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"platform": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"windows", "mac", "linux", "android", "ios"}, false),
						},
					},
				},
			},
			"input": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"exists": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"thumbprint": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"sha256": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"running": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceCloudflareDevicePostureRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	accountID := d.Get("account_id").(string)

	newDevicePostureRule := cloudflare.DevicePostureRule{
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Description: d.Get("description").(string),
		Schedule:    d.Get("schedule").(string),
	}

	err := setDevicePostureRuleMatch(&newDevicePostureRule, d)
	if err != nil {
		return fmt.Errorf("error creating Device Posture Rule with provided match input: %s", err)
	}

	setDevicePostureRuleInput(&newDevicePostureRule, d)

	log.Printf("[DEBUG] Creating Cloudflare Device Posture Rule from struct: %+v", newDevicePostureRule)

	rule, err := client.CreateDevicePostureRule(context.Background(), accountID, newDevicePostureRule)
	if err != nil {
		return fmt.Errorf("error creating Device Posture Rule for account %q: %s", accountID, err)
	}

	d.SetId(rule.ID)

	return resourceCloudflareDevicePostureRuleRead(d, meta)
}

func resourceCloudflareDevicePostureRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	accountID := d.Get("account_id").(string)

	devicePostureRule, err := client.DevicePostureRule(context.Background(), accountID, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Device Posture Rule %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error finding Device Posture Rule %q: %s", d.Id(), err)
	}

	d.Set("name", devicePostureRule.Name)
	d.Set("description", devicePostureRule.Description)
	d.Set("type", devicePostureRule.Type)
	d.Set("schedule", devicePostureRule.Schedule)
	d.Set("match", convertMatchToSchema(devicePostureRule.Match))
	d.Set("input", convertInputToSchema(devicePostureRule.Input))

	return nil
}

func resourceCloudflareDevicePostureRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	accountID := d.Get("account_id").(string)

	updatedDevicePostureRule := cloudflare.DevicePostureRule{
		ID:          d.Id(),
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Description: d.Get("description").(string),
		Schedule:    d.Get("schedule").(string),
	}

	err := setDevicePostureRuleMatch(&updatedDevicePostureRule, d)
	if err != nil {
		return fmt.Errorf("error creating Device Posture Rule with provided match input: %s", err)
	}

	setDevicePostureRuleInput(&updatedDevicePostureRule, d)

	log.Printf("[DEBUG] Updating Cloudflare Device Posture Rule from struct: %+v", updatedDevicePostureRule)

	devicePostureRule, err := client.UpdateDevicePostureRule(context.Background(), accountID, updatedDevicePostureRule)
	if err != nil {
		return fmt.Errorf("error updating Device Posture Rule for account %q: %s", accountID, err)
	}

	if devicePostureRule.ID == "" {
		return fmt.Errorf("failed to find Device Posture Rule ID in update response; resource was empty")
	}

	return resourceCloudflareDevicePostureRuleRead(d, meta)
}

func resourceCloudflareDevicePostureRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	appID := d.Id()
	accountID := d.Get("account_id").(string)

	log.Printf("[DEBUG] Deleting Cloudflare Device Posture Rule using ID: %s", appID)

	err := client.DeleteDevicePostureRule(context.Background(), accountID, appID)
	if err != nil {
		return fmt.Errorf("error deleting Device Posture Rule for account %q: %s", accountID, err)
	}

	resourceCloudflareDevicePostureRuleRead(d, meta)

	return nil
}

func resourceCloudflareDevicePostureRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"accountID/devicePostureRuleID\"", d.Id())
	}

	accountID, devicePostureRuleID := attributes[0], attributes[1]

	log.Printf("[DEBUG] Importing Cloudflare Device Posture Rule: id %s for account %s", devicePostureRuleID, accountID)

	d.Set("account_id", accountID)
	d.SetId(devicePostureRuleID)

	resourceCloudflareDevicePostureRuleRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

func setDevicePostureRuleInput(rule *cloudflare.DevicePostureRule, d *schema.ResourceData) {
	if _, ok := d.GetOk("input"); ok {
		input := cloudflare.DevicePostureRuleInput{}
		if inputID, ok := d.GetOk("input.0.id"); ok {
			input.ID = inputID.(string)
		}
		if p, ok := d.GetOk("input.0.path"); ok {
			input.Path = p.(string)
		}
		if exists, ok := d.GetOk("input.0.exists"); ok {
			input.Exists = exists.(bool)
		}
		if tp, ok := d.GetOk("input.0.thumbprint"); ok {
			input.Thumbprint = tp.(string)
		}
		if s, ok := d.GetOk("input.0.sha256"); ok {
			input.Sha256 = s.(string)
		}
		if running, ok := d.GetOk("input.0.running"); ok {
			input.Running = running.(bool)
		}

		rule.Input = input
	}
}

func setDevicePostureRuleMatch(rule *cloudflare.DevicePostureRule, d *schema.ResourceData) error {
	if _, ok := d.GetOk("match"); ok {
		match := d.Get("match").([]interface{})
		for _, v := range match {
			jsonString, err := json.Marshal(v.(map[string]interface{}))
			if err != nil {
				return err
			}

			var dprMatch cloudflare.DevicePostureRuleMatch
			err = json.Unmarshal(jsonString, &dprMatch)
			if err != nil {
				return err
			}

			rule.Match = append(rule.Match, dprMatch)
		}
	}

	return nil
}

func convertMatchToSchema(matches []cloudflare.DevicePostureRuleMatch) []map[string]interface{} {
	matchSchema := []map[string]interface{}{}
	for _, match := range matches {
		matchSchema = append(matchSchema, map[string]interface{}{"platform": match.Platform})
	}

	return matchSchema
}

func convertInputToSchema(input cloudflare.DevicePostureRuleInput) []map[string]interface{} {
	m := map[string]interface{}{
		"id":         input.ID,
		"path":       input.Path,
		"exists":     input.Exists,
		"thumbprint": input.Thumbprint,
		"sha256":     input.Sha256,
		"running":    input.Running,
	}

	return []map[string]interface{}{m}
}
