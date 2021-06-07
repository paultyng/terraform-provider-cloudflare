package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceCloudflareAuthenticatedOriginPullsCertificate() *schema.Resource {
	return &schema.Resource{
		// You cannot edit AOP certificates, rather, only upload new ones.
		Create: resourceCloudflareAuthenticatedOriginPullsCertificateCreate,
		Read:   resourceCloudflareAuthenticatedOriginPullsCertificateRead,
		Delete: resourceCloudflareAuthenticatedOriginPullsCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareAuthenticatedOriginPullsCertificateImport,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			"issuer": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"signature": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"serial_number": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"expires_on": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"uploaded_on": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"type": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"per-zone", "per-hostname"}, false),
				Required:     true,
				ForceNew:     true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
		},
	}
}

func resourceCloudflareAuthenticatedOriginPullsCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	switch aopType, ok := d.GetOk("type"); ok {
	case aopType == "per-zone":
		perZoneAOPCert := cloudflare.PerZoneAuthenticatedOriginPullsCertificateParams{
			Certificate: d.Get("certificate").(string),
			PrivateKey:  d.Get("private_key").(string),
		}
		record, err := client.UploadPerZoneAuthenticatedOriginPullsCertificate(context.Background(), zoneID, perZoneAOPCert)
		if err != nil {
			return fmt.Errorf("error uploading Per-Zone AOP certificate on zone %q: %s", zoneID, err)
		}
		d.SetId(record.ID)

		return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			resp, err := client.GetPerZoneAuthenticatedOriginPullsCertificateDetails(context.Background(), zoneID, record.ID)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("error reading Per Zone AOP certificate details: %s", err))
			}

			if resp.Status != "active" {
				return resource.RetryableError(fmt.Errorf("expected Per Zone AOP certificate to be active but was in state %s", resp.Status))
			}

			return resource.NonRetryableError(resourceCloudflareAuthenticatedOriginPullsCertificateRead(d, meta))
		})
	case aopType == "per-hostname":
		perHostnameAOPCert := cloudflare.PerHostnameAuthenticatedOriginPullsCertificateParams{
			Certificate: d.Get("certificate").(string),
			PrivateKey:  d.Get("private_key").(string),
		}
		record, err := client.UploadPerHostnameAuthenticatedOriginPullsCertificate(context.Background(), zoneID, perHostnameAOPCert)
		if err != nil {
			return fmt.Errorf("error uploading Per-Hostname AOP certificate on zone %q: %s", zoneID, err)
		}
		d.SetId(record.ID)

		return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			resp, err := client.GetPerHostnameAuthenticatedOriginPullsCertificate(context.Background(), zoneID, record.ID)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("error reading Per Hostname AOP certificate details: %s", err))
			}

			if resp.Status != "active" {
				return resource.RetryableError(fmt.Errorf("expected Per Hostname AOP certificate to be active but was in state %s", resp.Status))
			}

			return resource.NonRetryableError(resourceCloudflareAuthenticatedOriginPullsCertificateRead(d, meta))
		})
	}
	return nil
}

func resourceCloudflareAuthenticatedOriginPullsCertificateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	certID := d.Id()

	switch aopType, ok := d.GetOk("type"); ok {
	case aopType == "per-zone":
		record, err := client.GetPerZoneAuthenticatedOriginPullsCertificateDetails(context.Background(), zoneID, certID)
		if err != nil {
			if strings.Contains(err.Error(), "HTTP status 404") {
				log.Printf("[INFO] Per-Zone Authenticated Origin Pull certificate %s no longer exists", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error finding Per-Zone Authenticated Origin Pull certificate %q: %s", d.Id(), err)
		}
		d.Set("issuer", record.Issuer)
		d.Set("signature", record.Signature)
		d.Set("expires_on", record.ExpiresOn.Format(time.RFC3339Nano))
		d.Set("status", record.Status)
		d.Set("uploaded_on", record.UploadedOn.Format(time.RFC3339Nano))
	case aopType == "per-hostname":
		record, err := client.GetPerHostnameAuthenticatedOriginPullsCertificate(context.Background(), zoneID, certID)
		if err != nil {
			if strings.Contains(err.Error(), "HTTP status 404") {
				log.Printf("[INFO] Per-Hostname Authenticated Origin Pull certificate %s no longer exists", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error finding Per-Hostname Authenticated Origin Pull certificate %q: %s", d.Id(), err)
		}
		d.Set("issuer", record.Issuer)
		d.Set("signature", record.Signature)
		d.Set("serial_number", record.SerialNumber)
		d.Set("expires_on", record.ExpiresOn.Format(time.RFC3339Nano))
		d.Set("status", record.Status)
		d.Set("uploaded_on", record.UploadedOn.Format(time.RFC3339Nano))
	}
	return nil
}

func resourceCloudflareAuthenticatedOriginPullsCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	certID := d.Id()

	switch aopType, ok := d.GetOk("type"); ok {
	case aopType == "per-zone":
		_, err := client.DeletePerZoneAuthenticatedOriginPullsCertificate(context.Background(), zoneID, certID)
		if err != nil {
			return fmt.Errorf("Error deleting Per-Zone AOP certificate on zone %q: %s", zoneID, err)
		}
	case aopType == "per-hostname":
		_, err := client.DeletePerHostnameAuthenticatedOriginPullsCertificate(context.Background(), zoneID, certID)
		if err != nil {
			return fmt.Errorf("Error deleting Per-Hostname AOP certificate on zone %q: %s", zoneID, err)
		}
	}
	return nil
}

func resourceCloudflareAuthenticatedOriginPullsCertificateImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// split the id so we can lookup
	idAttr := strings.SplitN(d.Id(), "/", 3)

	if len(idAttr) != 3 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/type/certID\"", d.Id())
	}
	zoneID, aopType, certID := idAttr[0], idAttr[1], idAttr[2]
	d.Set("zone_id", zoneID)
	d.Set("type", aopType)
	d.SetId(certID)

	resourceCloudflareAuthenticatedOriginPullsCertificateRead(d, meta)
	return []*schema.ResourceData{d}, nil
}
