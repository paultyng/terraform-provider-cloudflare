package cloudflare

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
	"log"
	"strings"
)

func resourceCloudflareMagicFirewallRuleset() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareMagicFirewallRulesetCreate,
		Read:   resourceCloudflareMagicFirewallRulesetRead,
		Update: resourceCloudflareMagicFirewallRulesetUpdate,
		Delete: resourceCloudflareMagicFirewallRulesetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareMagicFirewallRulesetImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// the order of firewall rules has to be maintained, so we are using a list of maps here and validate the
			// map using a custom validator
			"rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     ruleElem,
			},
		},
	}
}

var ruleElem = &schema.Schema{
	Type: schema.TypeMap,
	Elem: &schema.Schema{
		Type: schema.TypeString,
	},
	ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
		ruleElemValidators := ruleElemValidators()
		ruleFields, ok := val.(map[string]interface{})

		if !ok {
			errs = append(errs, fmt.Errorf("got invalid map for rule element"))
			return
		}

		for k, v := range ruleFields {
			if _, ok := ruleElemValidators[k]; !ok {
				errs = append(errs, fmt.Errorf("%s is not supported in a rule", k))
			}

			validationFunc := ruleElemValidators[k]
			delete(ruleElemValidators, k)
			if validationFunc == nil {
				continue
			}

			w, e := validationFunc(v, k)
			warns = append(warns, w...)
			errs = append(errs, e...)
		}

		// attributes with non-nil validators must be set
		for k, v := range ruleElemValidators {
			if v == nil {
				continue
			}
			errs = append(errs, fmt.Errorf("%s must be set in a rule", k))
		}

		return
	},
}

func resourceCloudflareMagicFirewallRulesetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	client.AccountID = d.Get("account_id").(string)

	rules, err := buildMagicFirewallRulesetRulesFromResource(d.Get("rules"))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error building ruleset from resource"))
	}

	ruleset, err := client.CreateMagicFirewallRuleset(context.Background(),
		d.Get("name").(string),
		d.Get("description").(string),
		rules)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error creating firewall ruleset %s", d.Get("name").(string)))
	}

	d.SetId(ruleset.ID)

	return resourceCloudflareMagicFirewallRulesetRead(d, meta)
}

func resourceCloudflareMagicFirewallRulesetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*cloudflare.API)
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"accountID/rulesetID\"", d.Id())
	}

	accountID, rulesetID := attributes[0], attributes[1]
	d.SetId(rulesetID)
	d.Set("account_id", accountID)
	client.AccountID = accountID

	resourceCloudflareMagicFirewallRulesetRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

func resourceCloudflareMagicFirewallRulesetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	client.AccountID = d.Get("account_id").(string)

	ruleset, err := client.GetMagicFirewallRuleset(context.Background(), d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "could not find ruleset") {
			log.Printf("[INFO] Magic Firewall Ruleset %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return errors.Wrap(err, fmt.Sprintf("error reading Magic Firewall Ruleset ID %q", d.Id()))
	}

	d.Set("name", ruleset.Name)
	d.Set("description", ruleset.Description)
	d.Set("rules", buildStateFromMagicFirewallRulesetRules(ruleset.Rules))

	return nil
}

func resourceCloudflareMagicFirewallRulesetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	client.AccountID = d.Get("account_id").(string)

	rules, err := buildMagicFirewallRulesetRulesFromResource(d.Get("rules"))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error building ruleset from resource"))
	}

	_, err = client.UpdateMagicFirewallRuleset(context.Background(), d.Id(), d.Get("description").(string), rules)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error updating Magic Firewall ruleset with ID %q", d.Id()))
	}

	return resourceCloudflareMagicFirewallRulesetRead(d, meta)
}

func resourceCloudflareMagicFirewallRulesetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	client.AccountID = d.Get("account_id").(string)

	err := client.DeleteMagicFirewallRuleset(context.Background(), d.Id())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error deleting Magic Firewall ruleset with ID %q", d.Id()))
	}

	return nil
}

func ruleElemValidators() map[string]schema.SchemaValidateFunc {
	v := make(map[string]schema.SchemaValidateFunc)

	v["action"] = validation.StringInSlice([]string{"allow", "block"}, false)
	v["expression"] = validation.StringIsNotEmpty
	v["description"] = nil
	v["enabled"] = validation.StringInSlice([]string{"true", "false"}, false)

	return v
}

// receives the current rules and returns an interface for the state file
func buildStateFromMagicFirewallRulesetRules(r []cloudflare.MagicFirewallRulesetRule) interface{} {
	var ruleset []interface{}
	var rulesetRule map[string]interface{}

	for _, rule := range r {
		rulesetRule = make(map[string]interface{})

		rulesetRule["expression"] = rule.Expression

		if rule.Description != "" {
			rulesetRule["description"] = rule.Description
		}

		if rule.Enabled == true {
			rulesetRule["enabled"] = "true"
		} else {
			rulesetRule["enabled"] = "false"
		}

		if rule.Action == "skip" {
			rulesetRule["action"] = "allow"
		} else {
			rulesetRule["action"] = "block"
		}

		ruleset = append(ruleset, rulesetRule)
	}

	return ruleset
}

// receives the resource config and builds a ruleset rule array
func buildMagicFirewallRulesetRulesFromResource(r interface{}) ([]cloudflare.MagicFirewallRulesetRule, error) {
	var rulesetRules []cloudflare.MagicFirewallRulesetRule

	rules, ok := r.([]interface{})
	if !ok {
		return nil, errors.New("unable to create interface array type assertion")
	}

	for _, v := range rules {
		var rule cloudflare.MagicFirewallRulesetRule

		resourceRule, ok := v.(map[string]interface{})
		if !ok {
			return nil, errors.New("unable to create interface map type assertion for rule")
		}

		rule.Expression = resourceRule["expression"].(string)

		if resourceRule["description"] != nil {
			rule.Description = resourceRule["description"].(string)
		}

		if resourceRule["enabled"].(string) == "true" {
			rule.Enabled = true
		} else {
			rule.Enabled = false
		}

		if resourceRule["action"].(string) == "allow" {
			rule.Action = "skip"
			rule.ActionParameters = &cloudflare.MagicFirewallRulesetRuleActionParameters{
				Ruleset: "current",
			}
		} else {
			rule.Action = "block"
		}

		rulesetRules = append(rulesetRules, rule)
	}

	return rulesetRules, nil
}
