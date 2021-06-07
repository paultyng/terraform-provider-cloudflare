package cloudflare

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceCloudflareIPList() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareIPListCreate,
		Read:   resourceCloudflareIPListRead,
		Update: resourceCloudflareIPListUpdate,
		Delete: resourceCloudflareIPListDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareIPListImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[0-9a-z_]+$"), "IP List name must only contain lowercase letters, numbers and underscores"),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kind": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"ip"}, false),
				Required:     true,
			},
			"item": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     listItemElem,
			},
		},
	}
}

var listItemElem = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"value": {
			Type:     schema.TypeString,
			Required: true,
		},
		"comment": {
			Type:     schema.TypeString,
			Optional: true,
		},
	},
}

func resourceCloudflareIPListCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	client.AccountID = d.Get("account_id").(string)

	list, err := client.CreateIPList(context.Background(), d.Get("name").(string), d.Get("description").(string), d.Get("kind").(string))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error creating IP List %s", d.Get("name").(string)))
	}

	d.SetId(list.ID)

	if items, ok := d.GetOk("item"); ok {
		IPListItems := buildIPListItemsCreateRequest(items.(*schema.Set).List())
		_, err = client.CreateIPListItems(context.Background(), d.Id(), IPListItems)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error creating IP List Items"))
		}
	}

	return resourceCloudflareIPListRead(d, meta)
}

func resourceCloudflareIPListImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"accountID/listID\"", d.Id())
	}

	accountID, listID := attributes[0], attributes[1]
	d.SetId(listID)
	d.Set("account_id", accountID)

	resourceCloudflareIPListRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

func resourceCloudflareIPListRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	client.AccountID = d.Get("account_id").(string)

	list, err := client.GetIPList(context.Background(), d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "could not find list") {
			log.Printf("[INFO] IP List %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return errors.Wrap(err, fmt.Sprintf("error reading IP List with ID %q", d.Id()))
	}

	d.Set("name", list.Name)
	d.Set("description", list.Description)
	d.Set("kind", list.Kind)

	items, err := client.ListIPListItems(context.Background(), d.Id())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error reading IP List Items"))
	}

	var itemData []map[string]interface{}
	var item map[string]interface{}

	for _, i := range items {
		item = make(map[string]interface{})
		item["id"] = i.ID
		item["value"] = i.IP
		item["comment"] = i.Comment

		itemData = append(itemData, item)
	}

	d.Set("item", itemData)

	return nil
}

func resourceCloudflareIPListUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	client.AccountID = d.Get("account_id").(string)

	_, err := client.UpdateIPList(context.Background(), d.Id(), d.Get("description").(string))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error updating IP List description"))
	}

	if items, ok := d.GetOk("item"); ok {
		IPListItems := buildIPListItemsCreateRequest(items.(*schema.Set).List())
		_, err = client.ReplaceIPListItems(context.Background(), d.Id(), IPListItems)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error creating IP List Items"))
		}
	}

	return resourceCloudflareIPListRead(d, meta)
}

func resourceCloudflareIPListDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	client.AccountID = d.Get("account_id").(string)

	_, err := client.DeleteIPList(context.Background(), d.Id())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error deleting IP List with ID %q", d.Id()))
	}

	return nil
}

func buildIPListItemsCreateRequest(items []interface{}) []cloudflare.IPListItemCreateRequest {
	var IPListItems []cloudflare.IPListItemCreateRequest

	for _, item := range items {
		IPListItems = append(IPListItems, cloudflare.IPListItemCreateRequest{
			IP:      item.(map[string]interface{})["value"].(string),
			Comment: item.(map[string]interface{})["comment"].(string),
		})
	}

	return IPListItems
}
