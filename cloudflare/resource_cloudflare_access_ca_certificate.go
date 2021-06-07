package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudflareAccessCACertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareAccessCACertificateCreate,
		Read:   resourceCloudflareAccessCACertificateRead,
		Update: resourceCloudflareAccessCACertificateUpdate,
		Delete: resourceCloudflareAccessCACertificateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareAccessCACertificateImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"zone_id"},
			},
			"zone_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"account_id"},
			},
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aud": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCloudflareAccessCACertificateCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	var accessCACert cloudflare.AccessCACertificate
	if identifier.Type == AccountType {
		accessCACert, err = client.CreateAccessCACertificate(context.Background(), identifier.Value, d.Get("application_id").(string))
	} else {
		accessCACert, err = client.CreateZoneLevelAccessCACertificate(context.Background(), identifier.Value, d.Get("application_id").(string))
	}
	if err != nil {
		return fmt.Errorf("error creating Access CA Certificate for %s %q: %s", identifier.Type, identifier.Value, err)
	}

	d.SetId(accessCACert.ID)

	return resourceCloudflareAccessCACertificateRead(d, meta)
}

func resourceCloudflareAccessCACertificateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	applicationID := d.Get("application_id").(string)
	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	var accessCACert cloudflare.AccessCACertificate
	if identifier.Type == AccountType {
		accessCACert, err = client.AccessCACertificate(context.Background(), identifier.Value, applicationID)
	} else {
		accessCACert, err = client.ZoneLevelAccessCACertificate(context.Background(), identifier.Value, applicationID)
	}

	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Access CA Certificate %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error finding Access CA Certificate %q: %s", d.Id(), err)
	}

	d.Set("aud", accessCACert.Aud)
	d.Set("public_key", accessCACert.PublicKey)

	return nil
}

func resourceCloudflareAccessCACertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCloudflareAccessCACertificateDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	applicationID := d.Get("application_id").(string)

	log.Printf("[DEBUG] Deleting Cloudflare CA Certificate using ID: %s", d.Id())

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	if identifier.Type == AccountType {
		err = client.DeleteAccessCACertificate(context.Background(), identifier.Value, applicationID)
	} else {
		err = client.DeleteZoneLevelAccessCACertificate(context.Background(), identifier.Value, applicationID)
	}

	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceCloudflareAccessCACertificateImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 3)

	if len(attributes) != 3 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"account/accountID/accessCACertificateID\" or \"zone/zoneID/accessCACertificateID\"", d.Id())
	}

	identifierType, identifierID, accessCACertificateID := attributes[0], attributes[1], attributes[2]

	if AccessIdentifierType(identifierType) != AccountType && AccessIdentifierType(identifierType) != ZoneType {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"account/accountID/accessCACertificateID\" or \"zone/zoneID/accessCACertificateID\"", d.Id())
	}

	log.Printf("[DEBUG] Importing Cloudflare Access CA Certificate: id %s for %s %s", accessCACertificateID, identifierType, identifierID)

	//lintignore:R001
	d.Set(fmt.Sprintf("%s_id", identifierType), identifierID)
	d.SetId(accessCACertificateID)

	resourceCloudflareAccessCACertificateRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
