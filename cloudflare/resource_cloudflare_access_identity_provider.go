package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const CONCEALED_STRING = "**********************************"

func resourceCloudflareAccessIdentityProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareAccessIdentityProviderCreate,
		Read:   resourceCloudflareAccessIdentityProviderRead,
		Update: resourceCloudflareAccessIdentityProviderUpdate,
		Delete: resourceCloudflareAccessIdentityProviderDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareAccessIdentityProviderImport,
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
				ForceNew:      true,
				ConflictsWith: []string{"account_id"},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"centrify", "facebook", "google-apps", "oidc", "github", "google", "saml", "linkedin", "azureAD", "okta", "onetimepin", "onelogin", "yandex"}, false),
			},
			"config": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apps_domain": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"attributes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"auth_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"centrify_account": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"centrify_app_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"certs_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"client_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"client_secret": {
							Type:     schema.TypeString,
							Optional: true,
							// client_secret is a write only operation from the Cloudflare API
							// and once it's set, it is no longer accessible. To avoid storing
							// it and messing up the state, hardcode in the concealed version.
							StateFunc: func(val interface{}) string {
								return CONCEALED_STRING
							},
						},
						"directory_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"email_attribute_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"idp_public_cert": {
							Type:     schema.TypeString,
							Optional: true,
							// idp_public_cert is a write only operation from the Cloudflare
							// API and once it's set, it is no longer accessible. To avoid
							// storing it and messing up the state, hardcode in the concealed
							// version.
							StateFunc: func(val interface{}) string {
								return CONCEALED_STRING
							},
						},
						"issuer_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"okta_account": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"onelogin_account": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"redirect_url": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"sign_request": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"sso_target_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"support_groups": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"token_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceCloudflareAccessIdentityProviderRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	var accessIdentityProvider cloudflare.AccessIdentityProvider
	if identifier.Type == AccountType {
		accessIdentityProvider, err = client.AccessIdentityProviderDetails(context.Background(), identifier.Value, d.Id())
	} else {
		accessIdentityProvider, err = client.ZoneLevelAccessIdentityProviderDetails(context.Background(), identifier.Value, d.Id())
	}
	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Access Identity Provider %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("unable to find Access Identity Provider %q: %s", d.Id(), err)
	}

	d.SetId(accessIdentityProvider.ID)
	d.Set("name", accessIdentityProvider.Name)
	d.Set("type", accessIdentityProvider.Type)

	config := convertStructToSchema(d, accessIdentityProvider.Config)
	if configErr := d.Set("config", config); configErr != nil {
		return fmt.Errorf("error setting Access Identity Provider configuration: %s", configErr)
	}

	return nil
}

func resourceCloudflareAccessIdentityProviderCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	IDPConfig, _ := convertSchemaToStruct(d)

	identityProvider := cloudflare.AccessIdentityProvider{
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Config: IDPConfig,
	}

	log.Printf("[DEBUG] Creating Cloudflare Access Identity Provider from struct: %+v", identityProvider)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	var accessIdentityProvider cloudflare.AccessIdentityProvider
	if identifier.Type == AccountType {
		accessIdentityProvider, err = client.CreateAccessIdentityProvider(context.Background(), identifier.Value, identityProvider)
	} else {
		accessIdentityProvider, err = client.CreateZoneLevelAccessIdentityProvider(context.Background(), identifier.Value, identityProvider)
	}
	if err != nil {
		return fmt.Errorf("error creating Access Identity Provider for ID %q: %s", d.Id(), err)
	}

	d.SetId(accessIdentityProvider.ID)

	return resourceCloudflareAccessIdentityProviderRead(d, meta)
}

func resourceCloudflareAccessIdentityProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	IDPConfig, conversionErr := convertSchemaToStruct(d)
	if conversionErr != nil {
		return fmt.Errorf("failed to convert schema into struct: %s", conversionErr)
	}

	log.Printf("[DEBUG] updatedConfig: %+v", IDPConfig)
	updatedAccessIdentityProvider := cloudflare.AccessIdentityProvider{
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Config: IDPConfig,
	}

	log.Printf("[DEBUG] Updating Cloudflare Access Identity Provider from struct: %+v", updatedAccessIdentityProvider)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	var accessIdentityProvider cloudflare.AccessIdentityProvider
	if identifier.Type == AccountType {
		accessIdentityProvider, err = client.UpdateAccessIdentityProvider(context.Background(), identifier.Value, d.Id(), updatedAccessIdentityProvider)
	} else {
		accessIdentityProvider, err = client.UpdateZoneLevelAccessIdentityProvider(context.Background(), identifier.Value, d.Id(), updatedAccessIdentityProvider)
	}
	if err != nil {
		return fmt.Errorf("error updating Access Identity Provider for ID %q: %s", d.Id(), err)
	}

	if accessIdentityProvider.ID == "" {
		return fmt.Errorf("failed to find Access Identity Provider ID in update response; resource was empty")
	}

	return resourceCloudflareAccessIdentityProviderRead(d, meta)
}

func resourceCloudflareAccessIdentityProviderDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	log.Printf("[DEBUG] Deleting Cloudflare Access Identity Provider using ID: %s", d.Id())

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	if identifier.Type == AccountType {
		_, err = client.DeleteAccessIdentityProvider(context.Background(), identifier.Value, d.Id())
	} else {
		_, err = client.DeleteZoneLevelAccessIdentityProvider(context.Background(), identifier.Value, d.Id())
	}
	if err != nil {
		return fmt.Errorf("error deleting Access Identity Provider for ID %q: %s", d.Id(), err)
	}

	d.SetId("")

	return nil
}

func resourceCloudflareAccessIdentityProviderImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"accountID/accessIdentityProviderID\"", d.Id())
	}

	accountID, accessIdentityProviderID := attributes[0], attributes[1]

	log.Printf("[DEBUG] Importing Cloudflare Access Identity Provider: accountID=%s accessIdentityProviderID=%s", accountID, accessIdentityProviderID)

	d.Set("account_id", accountID)
	d.SetId(accessIdentityProviderID)

	resourceCloudflareAccessIdentityProviderRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

func convertSchemaToStruct(d *schema.ResourceData) (cloudflare.AccessIdentityProviderConfiguration, error) {
	IDPConfig := cloudflare.AccessIdentityProviderConfiguration{}

	if _, ok := d.GetOk("config"); ok {
		if _, ok := d.GetOk("config.0.attributes"); ok {
			attrData := make([]string, d.Get("config.0.attributes.#").(int))
			for id := range attrData {
				attrData[id] = d.Get(fmt.Sprintf("config.0.attributes.%d", id)).(string)
			}
			IDPConfig.Attributes = attrData
		}

		IDPConfig.AppsDomain = d.Get("config.0.apps_domain").(string)
		IDPConfig.AuthURL = d.Get("config.0.auth_url").(string)
		IDPConfig.CentrifyAccount = d.Get("config.0.centrify_account").(string)
		IDPConfig.CentrifyAppID = d.Get("config.0.centrify_app_id").(string)
		IDPConfig.CertsURL = d.Get("config.0.certs_url").(string)
		IDPConfig.ClientID = d.Get("config.0.client_id").(string)
		IDPConfig.ClientSecret = d.Get("config.0.client_secret").(string)
		IDPConfig.DirectoryID = d.Get("config.0.directory_id").(string)
		IDPConfig.EmailAttributeName = d.Get("config.0.email_attribute_name").(string)
		IDPConfig.IdpPublicCert = d.Get("config.0.idp_public_cert").(string)
		IDPConfig.IssuerURL = d.Get("config.0.issuer_url").(string)
		IDPConfig.OktaAccount = d.Get("config.0.okta_account").(string)
		IDPConfig.OneloginAccount = d.Get("config.0.onelogin_account").(string)
		IDPConfig.RedirectURL = d.Get("config.0.redirect_url").(string)
		IDPConfig.SignRequest = d.Get("config.0.sign_request").(bool)
		IDPConfig.SsoTargetURL = d.Get("config.0.sso_target_url").(string)
		IDPConfig.SupportGroups = d.Get("config.0.support_groups").(bool)
		IDPConfig.TokenURL = d.Get("config.0.token_url").(string)
	}

	return IDPConfig, nil
}

func convertStructToSchema(d *schema.ResourceData, options cloudflare.AccessIdentityProviderConfiguration) []interface{} {
	if _, ok := d.GetOk("config"); !ok {
		return []interface{}{}
	}

	attributes := make([]string, 0)
	for _, value := range options.Attributes {
		attributes = append(attributes, value)
	}

	m := map[string]interface{}{
		"apps_domain":          options.AppsDomain,
		"attributes":           attributes,
		"auth_url":             options.AuthURL,
		"centrify_account":     options.CentrifyAccount,
		"centrify_app_id":      options.CentrifyAppID,
		"certs_url":            options.CertsURL,
		"client_id":            options.ClientID,
		"client_secret":        options.ClientSecret,
		"directory_id":         options.DirectoryID,
		"email_attribute_name": options.EmailAttributeName,
		"idp_public_cert":      options.IdpPublicCert,
		"issuer_url":           options.IssuerURL,
		"okta_account":         options.OktaAccount,
		"onelogin_account":     options.OneloginAccount,
		"redirect_url":         options.RedirectURL,
		"sign_request":         options.SignRequest,
		"sso_target_url":       options.SsoTargetURL,
		"support_groups":       options.SupportGroups,
		"token_url":            options.TokenURL,
	}

	return []interface{}{m}
}
