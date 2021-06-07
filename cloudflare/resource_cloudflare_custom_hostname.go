package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceCloudflareCustomHostname() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareCustomHostnameCreate,
		Read:   resourceCloudflareCustomHostnameRead,
		Update: resourceCloudflareCustomHostnameUpdate,
		Delete: resourceCloudflareCustomHostnameDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareCustomHostnameImport,
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hostname": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"custom_origin_server": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssl": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					SchemaVersion: 1,
					Schema: map[string]*schema.Schema{
						"status": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"method": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"http", "txt", "email"}, false),
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "dv",
							ValidateFunc: validation.StringInSlice([]string{"dv"}, false),
						},
						"certificate_authority": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"lets_encrypt", "digicert"}, false),
						},
						"cname_target": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cname_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"wildcard": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"custom_certificate": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"custom_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"settings": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								SchemaVersion: 1,
								Schema: map[string]*schema.Schema{
									"http2": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
									},
									"tls13": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
									},
									"min_tls_version": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"1.0", "1.1", "1.2", "1.3"}, false),
									},
									"ciphers": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ownership_verification": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Resource{
					SchemaVersion: 1,
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ownership_verification_http": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Resource{
					SchemaVersion: 1,
					Schema: map[string]*schema.Schema{
						"http_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_body": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceCloudflareCustomHostnameRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	hostnameID := d.Id()

	customHostname, err := client.CustomHostname(context.Background(), zoneID, hostnameID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error reading custom hostname %q", hostnameID))
	}

	d.Set("hostname", customHostname.Hostname)
	d.Set("ssl.custom_origin_server", customHostname.CustomOriginServer)

	d.Set("ssl.0.type", customHostname.SSL.Type)
	d.Set("ssl.0.method", customHostname.SSL.Method)
	d.Set("ssl.0.wildcard", customHostname.SSL.Wildcard)
	d.Set("ssl.0.status", customHostname.SSL.Status)
	d.Set("ssl.0.cname_target", customHostname.SSL.CnameTarget)
	d.Set("ssl.0.cname_name", customHostname.SSL.CnameName)
	d.Set("ssl.0.custom_certificate", customHostname.SSL.CustomCertificate)
	d.Set("ssl.0.custom_key", customHostname.SSL.CustomKey)

	d.Set("ssl.0.settings.0.http2", customHostname.SSL.Settings.HTTP2)
	d.Set("ssl.0.settings.0.tls13", customHostname.SSL.Settings.TLS13)
	d.Set("ssl.0.settings.0.min_tls_version", customHostname.SSL.Settings.MinTLSVersion)
	d.Set("ssl.0.settings.0.ciphers", flattenStringList(customHostname.SSL.Settings.Ciphers))

	ownershipVerificationCfg := map[string]interface{}{}
	ownershipVerificationCfg["type"] = customHostname.OwnershipVerification.Type
	ownershipVerificationCfg["value"] = customHostname.OwnershipVerification.Value
	ownershipVerificationCfg["name"] = customHostname.OwnershipVerification.Name
	d.Set("ownership_verification", ownershipVerificationCfg)

	ownershipVerificationHTTPCfg := map[string]interface{}{}
	ownershipVerificationHTTPCfg["http_body"] = customHostname.OwnershipVerificationHTTP.HTTPBody
	ownershipVerificationHTTPCfg["http_url"] = customHostname.OwnershipVerificationHTTP.HTTPUrl
	d.Set("ownership_verification_http", ownershipVerificationHTTPCfg)

	return nil
}

func resourceCloudflareCustomHostnameDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	hostnameID := d.Id()

	err := client.DeleteCustomHostname(context.Background(), zoneID, hostnameID)
	if err != nil {
		return errors.Wrap(err, "failed to delete custom hostname certificate")
	}

	return nil
}

func resourceCloudflareCustomHostnameCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	certificate := buildCustomHostname(d)

	newCertificate, err := client.CreateCustomHostname(context.Background(), zoneID, certificate)
	if err != nil {
		return errors.Wrap(err, "failed to create custom hostname certificate")
	}

	d.SetId(newCertificate.Result.ID)

	return resourceCloudflareCustomHostnameRead(d, meta)
}

func resourceCloudflareCustomHostnameUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	hostnameID := d.Id()
	certificate := buildCustomHostname(d)

	_, err := client.UpdateCustomHostname(context.Background(), zoneID, hostnameID, certificate)
	if err != nil {
		return errors.Wrap(err, "failed to update custom hostname certificate")
	}

	return resourceCloudflareCustomHostnameRead(d, meta)
}

func resourceCloudflareCustomHostnameImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 2)

	if len(idAttr) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/customHostnameID\"", d.Id())
	}

	zoneID, hostnameID := idAttr[0], idAttr[1]

	log.Printf("[DEBUG] Importing Cloudflare Custom Hostname: id %s for zone %s", hostnameID, zoneID)

	d.Set("zone_id", zoneID)
	d.SetId(hostnameID)

	return []*schema.ResourceData{d}, nil
}

// buildCustomHostname takes the existing schema and returns a
// `cloudflare.CustomHostname`.
func buildCustomHostname(d *schema.ResourceData) cloudflare.CustomHostname {
	return cloudflare.CustomHostname{
		Hostname:           d.Get("hostname").(string),
		CustomOriginServer: d.Get("custom_origin_server").(string),
		SSL: cloudflare.CustomHostnameSSL{
			Method:            d.Get("ssl.0.method").(string),
			Type:              d.Get("ssl.0.type").(string),
			Wildcard:          &[]bool{d.Get("ssl.0.wildcard").(bool)}[0],
			CnameTarget:       d.Get("ssl.0.cname_target").(string),
			CnameName:         d.Get("ssl.0.cname_name").(string),
			CustomCertificate: d.Get("ssl.0.custom_certificate").(string),
			CustomKey:         d.Get("ssl.0.custom_key").(string),
			Settings: cloudflare.CustomHostnameSSLSettings{
				HTTP2:         d.Get("ssl.0.settings.0.http2").(string),
				TLS13:         d.Get("ssl.0.settings.0.tls13").(string),
				MinTLSVersion: d.Get("ssl.0.settings.0.min_tls_version").(string),
				Ciphers:       expandInterfaceToStringList(d.Get("ssl.0.settings.0.ciphers").([]interface{})),
			},
		},
	}
}
