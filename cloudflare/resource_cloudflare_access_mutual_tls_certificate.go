package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudflareAccessMutualTLSCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareAccessMutualTLSCertificateCreate,
		Read:   resourceCloudflareAccessMutualTLSCertificateRead,
		Update: resourceCloudflareAccessMutualTLSCertificateUpdate,
		Delete: resourceCloudflareAccessMutualTLSCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareAccessMutualTLSCertificateImport,
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"associated_hostnames": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCloudflareAccessMutualTLSCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	newAccessMutualTLSCertificate := cloudflare.AccessMutualTLSCertificate{
		Name:        d.Get("name").(string),
		Certificate: d.Get("certificate").(string),
	}
	newAccessMutualTLSCertificate.AssociatedHostnames = expandInterfaceToStringList(d.Get("associated_hostnames"))

	log.Printf("[DEBUG] Creating Cloudflare Access Mutual TLS certificate from struct: %+v", newAccessMutualTLSCertificate)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	var accessMutualTLSCert cloudflare.AccessMutualTLSCertificate
	if identifier.Type == AccountType {
		accessMutualTLSCert, err = client.CreateAccessMutualTLSCertificate(context.Background(), identifier.Value, newAccessMutualTLSCertificate)
	} else {
		accessMutualTLSCert, err = client.CreateZoneAccessMutualTLSCertificate(context.Background(), identifier.Value, newAccessMutualTLSCertificate)
	}
	if err != nil {
		return fmt.Errorf("error creating Access Mutual TLS Certificate for %s %q: %s", identifier.Type, identifier.Value, err)
	}

	d.SetId(accessMutualTLSCert.ID)

	return resourceCloudflareAccessMutualTLSCertificateRead(d, meta)
}

func resourceCloudflareAccessMutualTLSCertificateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	var accessMutualTLSCert cloudflare.AccessMutualTLSCertificate
	if identifier.Type == AccountType {
		accessMutualTLSCert, err = client.AccessMutualTLSCertificate(context.Background(), identifier.Value, d.Id())
	} else {
		accessMutualTLSCert, err = client.ZoneAccessMutualTLSCertificate(context.Background(), identifier.Value, d.Id())
	}

	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Access Mutal TLS Certificate %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error finding Access Mutual TLS Certificate %q: %s", d.Id(), err)
	}

	d.Set("name", accessMutualTLSCert.Name)
	d.Set("associated_hostnames", accessMutualTLSCert.AssociatedHostnames)
	d.Set("fingerprint", accessMutualTLSCert.Fingerprint)

	return nil
}

func resourceCloudflareAccessMutualTLSCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	updatedAccessMutualTLSCert := cloudflare.AccessMutualTLSCertificate{
		ID:   d.Id(),
		Name: d.Get("name").(string),
	}
	updatedAccessMutualTLSCert.AssociatedHostnames = expandInterfaceToStringList(d.Get("associated_hostnames"))

	log.Printf("[DEBUG] Updating Cloudflare Access Mutal TLS Certificate from struct: %+v", updatedAccessMutualTLSCert)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	if identifier.Type == AccountType {
		_, err = client.UpdateAccessMutualTLSCertificate(context.Background(), identifier.Value, d.Id(), updatedAccessMutualTLSCert)
	} else {
		_, err = client.UpdateZoneAccessMutualTLSCertificate(context.Background(), identifier.Value, d.Id(), updatedAccessMutualTLSCert)
	}
	if err != nil {
		return fmt.Errorf("error updating Access Mutual TLS Certificate for %s %q: %s", identifier.Type, identifier.Value, err)
	}

	return resourceCloudflareAccessMutualTLSCertificateRead(d, meta)
}

func resourceCloudflareAccessMutualTLSCertificateDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*cloudflare.API)
	certID := d.Id()

	log.Printf("[DEBUG] Deleting Cloudflare Access Mutual TLS Certificate using ID: %s", certID)

	identifier, err := initIdentifier(d)
	if err != nil {
		return err
	}

	// To actually delete the certificate, it cannot have any hostnames associated
	// with it so here we perform an update (to remove them) before we continue on
	// with wiping the certificate itself.
	deletedCertificate := cloudflare.AccessMutualTLSCertificate{
		ID:                  d.Id(),
		Name:                d.Get("name").(string),
		AssociatedHostnames: []string{},
	}

	if identifier.Type == AccountType {
		_, err = client.UpdateAccessMutualTLSCertificate(context.Background(), identifier.Value, d.Id(), deletedCertificate)
	} else {
		_, err = client.UpdateZoneAccessMutualTLSCertificate(context.Background(), identifier.Value, d.Id(), deletedCertificate)
	}

	if err != nil {
		return fmt.Errorf("error updating Access Mutual TLS Certificate for %s %q: %s", identifier.Type, identifier.Value, err)
	}

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if identifier.Type == AccountType {
			err = client.DeleteAccessMutualTLSCertificate(context.Background(), identifier.Value, certID)
		} else {
			err = client.DeleteZoneAccessMutualTLSCertificate(context.Background(), identifier.Value, certID)
		}

		if err != nil {
			if strings.Contains(err.Error(), "access.api.error.certificate_has_active_associations") {
				return resource.RetryableError(fmt.Errorf("certificate associations are not yet removed"))
			} else {
				return resource.NonRetryableError(fmt.Errorf("error deleting Access Mutual TLS Certificate for %s %q: %s", identifier.Type, identifier.Value, err))
			}
		}

		d.SetId("")

		return nil
	})

}

func resourceCloudflareAccessMutualTLSCertificateImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 3)

	if len(attributes) != 3 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"account/accountID/accessMutualTLSCertificateID\" or \"zone/zoneID/accessMutualTLSCertificateID\"", d.Id())
	}

	identifierType, identifierID, accessMutualTLSCertificateID := attributes[0], attributes[1], attributes[2]

	if AccessIdentifierType(identifierType) != AccountType && AccessIdentifierType(identifierType) != ZoneType {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"account/accountID/accessMutualTLSCertificateID\" or \"zone/zoneID/accessMutualTLSCertificateID\"", d.Id())
	}

	log.Printf("[DEBUG] Importing Cloudflare Access Mutual TLS Certificate: id %s for %s %s", accessMutualTLSCertificateID, identifierType, identifierID)

	//lintignore:R001
	d.Set(fmt.Sprintf("%s_id", identifierType), identifierID)
	d.SetId(accessMutualTLSCertificateID)

	resourceCloudflareAccessMutualTLSCertificateRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
