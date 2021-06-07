package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceCloudflareCertificatePack() *schema.Resource {
	return &schema.Resource{
		// Intentionally no Update method as certificates require replacement for
		// any changes made.
		Create: resourceCloudflareCertificatePackCreate,
		Read:   resourceCloudflareCertificatePackRead,
		Delete: resourceCloudflareCertificatePackDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareCertificatePackImport,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"custom", "dedicated_custom", "advanced"}, false),
			},
			"hosts": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"validation_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"txt", "http", "email"}, false),
			},
			"validity_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntInSlice([]int{14, 30, 90, 365}),
			},
			"certificate_authority": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"digicert", "lets_encrypt"}, false),
			},
			"cloudflare_branding": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCloudflareCertificatePackCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)
	certificatePackType := d.Get("type").(string)
	certificateHostSet := d.Get("hosts").(*schema.Set)
	certificatePackID := ""

	if certificatePackType == "advanced" {
		validationMethod := d.Get("validation_method").(string)
		validityDays := d.Get("validity_days").(int)
		ca := d.Get("certificate_authority").(string)
		cloudflareBranding := d.Get("cloudflare_branding").(bool)

		cert := cloudflare.CertificatePackAdvancedCertificate{
			Type:                 "advanced",
			Hosts:                expandInterfaceToStringList(certificateHostSet.List()),
			ValidationMethod:     validationMethod,
			ValidityDays:         validityDays,
			CertificateAuthority: ca,
			CloudflareBranding:   cloudflareBranding,
		}
		certPackResponse, err := client.CreateAdvancedCertificatePack(context.Background(), zoneID, cert)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create certificate pack: %s", err))
		}
		certificatePackID = certPackResponse.ID
	} else {
		cert := cloudflare.CertificatePackRequest{
			Type:  certificatePackType,
			Hosts: expandInterfaceToStringList(certificateHostSet.List()),
		}
		certPackResponse, err := client.CreateCertificatePack(context.Background(), zoneID, cert)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create certificate pack: %s", err))
		}
		certificatePackID = certPackResponse.ID
	}

	d.SetId(certificatePackID)

	return resourceCloudflareCertificatePackRead(d, meta)
}

func resourceCloudflareCertificatePackRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	certificatePack, err := client.CertificatePack(context.Background(), zoneID, d.Id())
	if err != nil {
		return errors.Wrap(err, "failed to fetch certificate pack")
	}

	d.Set("type", certificatePack.Type)
	d.Set("hosts", expandStringListToSet(certificatePack.Hosts))

	return nil
}

func resourceCloudflareCertificatePackDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	zoneID := d.Get("zone_id").(string)

	err := client.DeleteCertificatePack(context.Background(), zoneID, d.Id())
	if err != nil {
		return errors.Wrap(err, "failed to delete certificate pack")
	}

	resourceCloudflareCertificatePackRead(d, meta)

	return nil
}

func resourceCloudflareCertificatePackImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/certificatePackID\"", d.Id())
	}

	zoneID, certificatePackID := attributes[0], attributes[1]

	log.Printf("[DEBUG] Importing Cloudflare Certificate Pack: id %s for zone %s", certificatePackID, zoneID)

	d.Set("zone_id", zoneID)
	d.SetId(certificatePackID)

	resourceCloudflareCertificatePackRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
