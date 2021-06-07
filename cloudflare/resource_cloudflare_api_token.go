package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceCloudflareApiToken() *schema.Resource {
	p := schema.Resource{
		Schema: map[string]*schema.Schema{
			"resources": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"permission_groups": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"effect": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "allow",
				ValidateFunc: validation.StringInSlice([]string{"allow", "deny"}, false),
			},
		},
	}

	return &schema.Resource{
		Create: resourceCloudflareApiTokenCreate,
		Read:   resourceCloudflareApiTokenRead,
		Update: resourceCloudflareApiTokenUpdate,
		Delete: resourceCloudflareApiTokenDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:     schema.TypeSet,
				Set:      schema.HashResource(&p),
				Required: true,
				Elem:     &p,
			},
			"condition": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"request_ip": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"in": {
										Type:     schema.TypeList,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
									},
									"not_in": {
										Type:     schema.TypeList,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"value": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issued_on": {
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

func buildAPIToken(d *schema.ResourceData) cloudflare.APIToken {
	token := cloudflare.APIToken{}

	token.Name = d.Get("name").(string)
	token.Policies = resourceDataToApiTokenPolices(d)

	ipsIn := []string{}
	ipsNotIn := []string{}
	if ips, ok := d.GetOk("condition.0.request_ip.0.in"); ok {
		ipsIn = expandInterfaceToStringList(ips)
	}

	if ips, ok := d.GetOk("condition.0.request_ip.0.not_in"); ok {
		ipsNotIn = expandInterfaceToStringList(ips)
	}

	if len(ipsIn) > 0 || len(ipsNotIn) > 0 {
		token.Condition = &cloudflare.APITokenCondition{
			RequestIP: &cloudflare.APITokenRequestIPCondition{},
		}

		if len(ipsIn) > 0 {
			token.Condition.RequestIP.In = ipsIn
		}

		if len(ipsNotIn) > 0 {
			token.Condition.RequestIP.NotIn = ipsNotIn
		}
	}

	return token
}

func resourceCloudflareApiTokenCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	name := d.Get("name").(string)

	log.Printf("[INFO] Creating Cloudflare API Token: name %s", name)

	t := buildAPIToken(d)
	t, err := client.CreateAPIToken(context.Background(), t)
	if err != nil {
		return fmt.Errorf("error creating Cloudflare API Token %q: %s", name, err)
	}

	d.SetId(t.ID)
	d.Set("status", t.Status)
	d.Set("issued_on", t.IssuedOn.Format(time.RFC3339Nano))
	d.Set("modified_on", t.ModifiedOn.Format(time.RFC3339Nano))
	d.Set("value", t.Value)

	return resourceCloudflareApiTokenRead(d, meta)
}

func resourceDataToApiTokenPolices(d *schema.ResourceData) []cloudflare.APITokenPolicies {
	policies := d.Get("policy").(*schema.Set).List()
	var cfPolicies []cloudflare.APITokenPolicies

	for _, p := range policies {
		policy := p.(map[string]interface{})

		permissionGroups := expandInterfaceToStringList(policy["permission_groups"])
		if len(permissionGroups) == 0 {
			continue
		}
		var cfPermissionGroups []cloudflare.APITokenPermissionGroups
		for _, pg := range permissionGroups {
			cfPermissionGroups = append(cfPermissionGroups, cloudflare.APITokenPermissionGroups{
				ID: pg,
			})
		}

		cfResources := map[string]interface{}{}
		for k, v := range policy["resources"].(map[string]interface{}) {
			// value can be object or just a string ("*"), try to convert it to map
			obj := map[string]string{}
			if err := json.Unmarshal([]byte(v.(string)), &obj); err == nil {
				cfResources[k] = obj
			} else {
				cfResources[k] = v
			}
		}

		cfPolicies = append(cfPolicies, cloudflare.APITokenPolicies{
			Effect:           policy["effect"].(string),
			Resources:        cfResources,
			PermissionGroups: cfPermissionGroups,
		})
	}

	return cfPolicies
}

func resourceCloudflareApiTokenRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	tokenID := d.Id()

	t, err := client.GetAPIToken(context.Background(), tokenID)

	log.Printf("[DEBUG] Cloudflare API Token: %+v", t)

	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Cloudflare API Token %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error finding Cloudflare API Token %q: %s", d.Id(), err)
	}

	policies := []map[string]interface{}{}

	for _, p := range t.Policies {
		permissionGroups := []string{}
		for _, v := range p.PermissionGroups {
			permissionGroups = append(permissionGroups, v.ID)
		}

		policies = append(policies, map[string]interface{}{
			"resources":         p.Resources,
			"permission_groups": permissionGroups,
			"effect":            p.Effect,
		})
	}

	d.Set("name", t.Name)
	d.Set("policy", policies)
	d.Set("status", t.Status)
	d.Set("issued_on", t.IssuedOn.Format(time.RFC3339Nano))
	d.Set("modified_on", time.Now().Format(time.RFC3339Nano))

	var ipIn []string
	var ipNotIn []string
	if t.Condition != nil && t.Condition.RequestIP != nil && t.Condition.RequestIP.In != nil {
		ipIn = t.Condition.RequestIP.In
	}

	if t.Condition != nil && t.Condition.RequestIP != nil && t.Condition.RequestIP.NotIn != nil {
		ipNotIn = t.Condition.RequestIP.NotIn
	}

	if len(ipIn) > 0 || len(ipNotIn) > 0 {
		d.Set("condition", []map[string]interface{}{{
			"request_ip": []map[string]interface{}{{
				"not_in": ipNotIn,
				"in":     ipIn,
			}},
		}})
	}

	return nil
}

func resourceCloudflareApiTokenUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	name := d.Get("name").(string)
	tokenID := d.Id()

	t := buildAPIToken(d)

	log.Printf("[INFO] Updating Cloudflare API Token: name %s", name)

	t, err := client.UpdateAPIToken(context.Background(), tokenID, t)
	if err != nil {
		return fmt.Errorf("error updating Cloudflare API Token %q: %s", name, err)
	}

	return resourceCloudflareApiTokenRead(d, meta)
}

func resourceCloudflareApiTokenDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	tokenID := d.Id()

	log.Printf("[INFO] Deleting Cloudflare API Token: id %s", tokenID)

	err := client.DeleteAPIToken(context.Background(), tokenID)
	if err != nil {
		return fmt.Errorf("error deleting Cloudflare API Token: %s", err)
	}

	return nil
}
