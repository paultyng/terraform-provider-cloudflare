package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudflareAccessGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareAccessGroupCreate,
		Read:   resourceCloudflareAccessGroupRead,
		Update: resourceCloudflareAccessGroupUpdate,
		Delete: resourceCloudflareAccessGroupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareAccessGroupImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"zone_id"},
			},
			"zone_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"account_id"},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"require": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     AccessGroupOptionSchemaElement,
			},
			"exclude": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     AccessGroupOptionSchemaElement,
			},
			"include": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     AccessGroupOptionSchemaElement,
			},
		},
	}
}

// AccessGroupOptionSchemaElement is used by `require`, `exclude` and `include`
// attributes to build out the expected access conditions.
var AccessGroupOptionSchemaElement = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"email": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"email_domain": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ip": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"service_token": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"any_valid_service_token": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"group": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"everyone": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"certificate": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"common_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"auth_method": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"geo": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"login_method": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"device_posture": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"gsuite": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"email": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"identity_provider_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"github": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"teams": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"identity_provider_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"azure": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"identity_provider_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"okta": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"identity_provider_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"saml": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"attribute_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"attribute_value": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"identity_provider_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	},
}

func resourceCloudflareAccessGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	var accessGroup cloudflare.AccessGroup
	if identifier.Type == AccountType {
		accessGroup, err = client.AccessGroup(context.Background(), identifier.Value, d.Id())
	} else {
		accessGroup, err = client.ZoneLevelAccessGroup(context.Background(), identifier.Value, d.Id())
	}

	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Access Group %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error finding Access Group %q: %s", d.Id(), err)
	}

	d.Set("name", accessGroup.Name)

	if err := d.Set("require", TransformAccessGroupForSchema(accessGroup.Require)); err != nil {
		return fmt.Errorf("failed to set require attribute: %s", err)
	}

	if err := d.Set("exclude", TransformAccessGroupForSchema(accessGroup.Exclude)); err != nil {
		return fmt.Errorf("failed to set exclude attribute: %s", err)
	}

	if err := d.Set("include", TransformAccessGroupForSchema(accessGroup.Include)); err != nil {
		return fmt.Errorf("failed to set include attribute: %s", err)
	}

	return nil
}

func resourceCloudflareAccessGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	newAccessGroup := cloudflare.AccessGroup{
		Name: d.Get("name").(string),
	}

	newAccessGroup = appendConditionalAccessGroupFields(newAccessGroup, d)

	log.Printf("[DEBUG] Creating Cloudflare Access Group from struct: %+v", newAccessGroup)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	var accessGroup cloudflare.AccessGroup
	if identifier.Type == AccountType {
		accessGroup, err = client.CreateAccessGroup(context.Background(), identifier.Value, newAccessGroup)
	} else {
		accessGroup, err = client.CreateZoneLevelAccessGroup(context.Background(), identifier.Value, newAccessGroup)
	}
	if err != nil {
		return fmt.Errorf("error creating Access Group for ID %q: %s", accessGroup.ID, err)
	}

	d.SetId(accessGroup.ID)
	resourceCloudflareAccessGroupRead(d, meta)

	return nil
}

func resourceCloudflareAccessGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	updatedAccessGroup := cloudflare.AccessGroup{
		Name: d.Get("name").(string),
		ID:   d.Id(),
	}

	updatedAccessGroup = appendConditionalAccessGroupFields(updatedAccessGroup, d)

	log.Printf("[DEBUG] Updating Cloudflare Access Group from struct: %+v", updatedAccessGroup)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	var accessGroup cloudflare.AccessGroup
	if identifier.Type == AccountType {
		accessGroup, err = client.UpdateAccessGroup(context.Background(), identifier.Value, updatedAccessGroup)
	} else {
		accessGroup, err = client.UpdateZoneLevelAccessGroup(context.Background(), identifier.Value, updatedAccessGroup)
	}
	if err != nil {
		return fmt.Errorf("error updating Access Group for ID %q: %s", d.Id(), err)
	}

	if accessGroup.ID == "" {
		return fmt.Errorf("failed to find Access Group ID in update response; resource was empty")
	}

	return resourceCloudflareAccessGroupRead(d, meta)
}

func resourceCloudflareAccessGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	log.Printf("[DEBUG] Deleting Cloudflare Access Group using ID: %s", d.Id())

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	if identifier.Type == AccountType {
		err = client.DeleteAccessGroup(context.Background(), identifier.Value, d.Id())
	} else {
		err = client.DeleteZoneLevelAccessGroup(context.Background(), identifier.Value, d.Id())
	}
	if err != nil {
		return fmt.Errorf("error deleting Access Group for ID %q: %s", d.Id(), err)
	}

	resourceCloudflareAccessGroupRead(d, meta)

	return nil
}

func resourceCloudflareAccessGroupImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"accountID/accessGroupID\"", d.Id())
	}

	accountID, accessGroupID := attributes[0], attributes[1]

	log.Printf("[DEBUG] Importing Cloudflare Access Group: accountID %q, accessGroupID %q", accountID, accessGroupID)

	d.Set("account_id", accountID)
	d.SetId(accessGroupID)

	resourceCloudflareAccessGroupRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

// appendConditionalAccessGroupFields determines which of the
// conditional group enforcement fields it should append to the
// AccessGroup by iterating over the provided values and generating the
// correct structs.
func appendConditionalAccessGroupFields(group cloudflare.AccessGroup, d *schema.ResourceData) cloudflare.AccessGroup {
	exclude := d.Get("exclude").([]interface{})
	for _, value := range exclude {
		group.Exclude = BuildAccessGroupCondition(value.(map[string]interface{}))
	}

	require := d.Get("require").([]interface{})
	for _, value := range require {
		group.Require = BuildAccessGroupCondition(value.(map[string]interface{}))
	}

	include := d.Get("include").([]interface{})
	for _, value := range include {
		group.Include = BuildAccessGroupCondition(value.(map[string]interface{}))
	}

	return group
}

// BuildAccessGroupCondition iterates the provided `map` of values and
// generates the required (repetitive) structs.
//
// Returns the intended combination structure of Access Groups to build the
// desired policy.
func BuildAccessGroupCondition(options map[string]interface{}) []interface{} {
	var group []interface{}
	for accessGroupType, values := range options {
		if accessGroupType == "everyone" {
			if values == true {
				group = append(group, cloudflare.AccessGroupEveryone{})
			}
		} else if accessGroupType == "any_valid_service_token" {
			if values == true {
				group = append(group, cloudflare.AccessGroupAnyValidServiceToken{})
			}
		} else if accessGroupType == "certificate" {
			if values == true {
				group = append(group, cloudflare.AccessGroupCertificate{})
			}
		} else if accessGroupType == "common_name" {
			if values != "" {
				group = append(group, cloudflare.AccessGroupCertificateCommonName{CommonName: struct {
					CommonName string `json:"common_name"`
				}{CommonName: values.(string)}})
			}
		} else if accessGroupType == "auth_method" {
			if values != "" {
				group = append(group, cloudflare.AccessGroupAuthMethod{AuthMethod: struct {
					AuthMethod string `json:"auth_method"`
				}{AuthMethod: values.(string)}})
			}
		} else if accessGroupType == "gsuite" {
			for _, v := range values.([]interface{}) {
				gsuiteCfg := v.(map[string]interface{})
				for _, email := range gsuiteCfg["email"].([]interface{}) {
					group = append(group, cloudflare.AccessGroupGSuite{Gsuite: struct {
						Email              string `json:"email"`
						IdentityProviderID string `json:"identity_provider_id"`
					}{
						Email:              email.(string),
						IdentityProviderID: gsuiteCfg["identity_provider_id"].(string),
					}})
				}
			}
		} else if accessGroupType == "github" {
			for _, v := range values.([]interface{}) {
				githubCfg := v.(map[string]interface{})
				for _, team := range githubCfg["teams"].([]interface{}) {
					group = append(group, cloudflare.AccessGroupGitHub{GitHubOrganization: struct {
						Name               string `json:"name"`
						Team               string `json:"team,omitempty"`
						IdentityProviderID string `json:"identity_provider_id"`
					}{
						Name:               githubCfg["name"].(string),
						Team:               team.(string),
						IdentityProviderID: githubCfg["identity_provider_id"].(string),
					}})
				}
			}
		} else if accessGroupType == "azure" {
			for _, v := range values.([]interface{}) {
				azureCfg := v.(map[string]interface{})
				for _, id := range azureCfg["id"].([]interface{}) {
					group = append(group, cloudflare.AccessGroupAzure{AzureAD: struct {
						ID                 string `json:"id"`
						IdentityProviderID string `json:"identity_provider_id"`
					}{
						ID:                 id.(string),
						IdentityProviderID: azureCfg["identity_provider_id"].(string),
					}})
				}
			}
		} else if accessGroupType == "okta" {
			for _, v := range values.([]interface{}) {
				oktaCfg := v.(map[string]interface{})
				for _, name := range oktaCfg["name"].([]interface{}) {
					group = append(group, cloudflare.AccessGroupOkta{Okta: struct {
						Name               string `json:"name"`
						IdentityProviderID string `json:"identity_provider_id"`
					}{
						Name:               name.(string),
						IdentityProviderID: oktaCfg["identity_provider_id"].(string),
					}})
				}
			}
		} else if accessGroupType == "saml" {
			for _, v := range values.([]interface{}) {
				samlCfg := v.(map[string]interface{})
				group = append(group, cloudflare.AccessGroupSAML{Saml: struct {
					AttributeName      string `json:"attribute_name"`
					AttributeValue     string `json:"attribute_value"`
					IdentityProviderID string `json:"identity_provider_id"`
				}{
					AttributeName:      samlCfg["attribute_name"].(string),
					AttributeValue:     samlCfg["attribute_value"].(string),
					IdentityProviderID: samlCfg["identity_provider_id"].(string),
				}})
			}
		} else {
			for _, value := range values.([]interface{}) {
				switch accessGroupType {
				case "email":
					group = append(group, cloudflare.AccessGroupEmail{Email: struct {
						Email string `json:"email"`
					}{Email: value.(string)}})
				case "email_domain":
					group = append(group, cloudflare.AccessGroupEmailDomain{EmailDomain: struct {
						Domain string `json:"domain"`
					}{Domain: value.(string)}})
				case "ip":
					group = append(group, cloudflare.AccessGroupIP{IP: struct {
						IP string `json:"ip"`
					}{IP: value.(string)}})
				case "service_token":
					group = append(group, cloudflare.AccessGroupServiceToken{ServiceToken: struct {
						ID string `json:"token_id"`
					}{ID: value.(string)}})
				case "group":
					group = append(group, cloudflare.AccessGroupAccessGroup{Group: struct {
						ID string `json:"id"`
					}{ID: value.(string)}})
				case "geo":
					group = append(group, cloudflare.AccessGroupGeo{Geo: struct {
						CountryCode string `json:"country_code"`
					}{CountryCode: value.(string)}})
				case "login_method":
					group = append(group, cloudflare.AccessGroupLoginMethod{LoginMethod: struct {
						ID string `json:"id"`
					}{ID: value.(string)}})
				case "device_posture":
					group = append(group, cloudflare.AccessGroupDevicePosture{DevicePosture: struct {
						ID string `json:"integration_uid"`
					}{ID: value.(string)}})
				}
			}
		}
	}

	return group
}

// TransformAccessGroupForSchema takes the incoming `accessGroup` from the API
// response and converts it to a usable schema for the conditions.
func TransformAccessGroupForSchema(accessGroup []interface{}) []map[string]interface{} {
	data := []map[string]interface{}{}
	emails := []string{}
	emailDomains := []string{}
	ips := []string{}
	serviceTokens := []string{}
	groups := []string{}
	commonName := ""
	authMethod := ""
	geos := []string{}
	loginMethod := []string{}
	oktaID := ""
	oktaGroups := []string{}
	gsuiteID := ""
	gsuiteEmails := []string{}
	githubName := ""
	githubTeams := []string{}
	githubID := ""
	azureID := ""
	azureIDs := []string{}
	samlAttrName := ""
	samlAttrValue := ""
	devicePostureRuleIDs := []string{}

	for _, group := range accessGroup {
		for groupKey, groupValue := range group.(map[string]interface{}) {
			switch groupKey {
			case "everyone", "any_valid_service_token", "certificate":
				data = append(data, map[string]interface{}{
					groupKey: true,
				})
			case "email":
				for _, email := range groupValue.(map[string]interface{}) {
					emails = append(emails, email.(string))
				}
			case "email_domain":
				for _, domain := range groupValue.(map[string]interface{}) {
					emailDomains = append(emailDomains, domain.(string))
				}
			case "ip":
				for _, ip := range groupValue.(map[string]interface{}) {
					ips = append(ips, ip.(string))
				}
			case "service_token":
				for _, serviceToken := range groupValue.(map[string]interface{}) {
					serviceTokens = append(serviceTokens, serviceToken.(string))
				}
			case "common_name":
				for _, name := range groupValue.(map[string]interface{}) {
					commonName = name.(string)
				}
			case "auth_method":
				for _, method := range groupValue.(map[string]interface{}) {
					authMethod = method.(string)
				}
			case "geo":
				for _, geo := range groupValue.(map[string]interface{}) {
					geos = append(geos, geo.(string))
				}
			case "login_method":
				for _, method := range groupValue.(map[string]interface{}) {
					loginMethod = append(loginMethod, method.(string))
				}
			case "okta":
				oktaCfg := groupValue.(map[string]interface{})
				oktaID = oktaCfg["identity_provider_id"].(string)
				oktaGroups = append(oktaGroups, oktaCfg["name"].(string))
			case "gsuite":
				gsuiteCfg := groupValue.(map[string]interface{})
				gsuiteID = gsuiteCfg["identity_provider_id"].(string)
				gsuiteEmails = append(gsuiteEmails, gsuiteCfg["email"].(string))
			case "github-organization":
				githubCfg := groupValue.(map[string]interface{})
				githubID = githubCfg["identity_provider_id"].(string)
				githubName = githubCfg["name"].(string)
				githubTeams = append(githubTeams, githubCfg["team"].(string))
			case "azure":
				azureCfg := groupValue.(map[string]interface{})
				azureID = azureCfg["identity_provider_id"].(string)
				azureIDs = append(azureIDs, azureCfg["id"].(string))
			case "saml":
				samlCfg := groupValue.(map[string]interface{})
				samlAttrName = samlCfg["attribute_name"].(string)
				samlAttrValue = samlCfg["attribute_value"].(string)
			case "group":
				for _, group := range groupValue.(map[string]interface{}) {
					groups = append(groups, group.(string))
				}
			case "device_posture":
				for _, dprID := range groupValue.(map[string]interface{}) {
					devicePostureRuleIDs = append(devicePostureRuleIDs, dprID.(string))
				}
			default:
				log.Printf("[DEBUG] Access Group key %q not transformed", groupKey)
			}
		}
	}

	if len(emails) > 0 {
		data = append(data, map[string]interface{}{
			"email": emails,
		})
	}

	if len(emailDomains) > 0 {
		data = append(data, map[string]interface{}{
			"email_domain": emailDomains,
		})
	}

	if len(ips) > 0 {
		data = append(data, map[string]interface{}{
			"ip": ips,
		})
	}

	if len(serviceTokens) > 0 {
		data = append(data, map[string]interface{}{
			"service_token": serviceTokens,
		})
	}

	if commonName != "" {
		data = append(data, map[string]interface{}{
			"common_name": commonName,
		})
	}

	if authMethod != "" {
		data = append(data, map[string]interface{}{
			"auth_method": authMethod,
		})
	}

	if len(geos) > 0 {
		data = append(data, map[string]interface{}{
			"geo": geos,
		})
	}

	if len(loginMethod) > 0 {
		data = append(data, map[string]interface{}{
			"login_method": loginMethod,
		})
	}

	if len(oktaGroups) > 0 && oktaID != "" {
		data = append(data, map[string]interface{}{
			"okta": []interface{}{
				map[string]interface{}{
					"identity_provider_id": oktaID,
					"name":                 oktaGroups,
				}},
		})
	}

	if len(gsuiteEmails) > 0 && gsuiteID != "" {
		data = append(data, map[string]interface{}{
			"gsuite": []interface{}{
				map[string]interface{}{
					"identity_provider_id": gsuiteID,
					"email":                gsuiteEmails,
				}},
		})
	}

	if len(githubTeams) > 0 && githubID != "" && githubName != "" {
		data = append(data, map[string]interface{}{
			"github": []interface{}{
				map[string]interface{}{
					"name":                 githubName,
					"teams":                githubTeams,
					"identity_provider_id": githubID,
				}},
		})
	}

	if len(azureIDs) > 0 && azureID != "" {
		data = append(data, map[string]interface{}{
			"azure": []interface{}{
				map[string]interface{}{
					"identity_provider_id": azureID,
					"id":                   azureIDs,
				}},
		})
	}

	if samlAttrName != "" && samlAttrValue != "" {
		data = append(data, map[string]interface{}{
			"saml": []interface{}{
				map[string]interface{}{
					"attribute_name":  samlAttrName,
					"attribute_value": samlAttrValue,
				}},
		})
	}

	if len(groups) > 0 {
		data = append(data, map[string]interface{}{
			"group": groups,
		})
	}

	if len(devicePostureRuleIDs) > 0 {
		data = append(data, map[string]interface{}{
			"device_posture": devicePostureRuleIDs,
		})
	}

	return data
}
