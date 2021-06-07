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

func resourceCloudflareTeamsList() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareTeamsListCreate,
		Read:   resourceCloudflareTeamsListRead,
		Update: resourceCloudflareTeamsListUpdate,
		Delete: resourceCloudflareTeamsListDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareTeamsListImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"SERIAL", "URL", "DOMAIN", "EMAIL"}, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"items": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceCloudflareTeamsListCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	newTeamsList := cloudflare.TeamsList{
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Description: d.Get("description").(string),
	}

	itemValues := d.Get("items").([]interface{})
	for _, v := range itemValues {
		newTeamsList.Items = append(newTeamsList.Items, cloudflare.TeamsListItem{Value: v.(string)})
	}

	log.Printf("[DEBUG] Creating Cloudflare Teams List from struct: %+v", newTeamsList)

	accountID := d.Get("account_id").(string)

	list, err := client.CreateTeamsList(context.Background(), accountID, newTeamsList)
	if err != nil {
		return fmt.Errorf("error creating Teams List for account %q: %s", accountID, err)
	}

	d.SetId(list.ID)

	return resourceCloudflareTeamsListRead(d, meta)
}

func resourceCloudflareTeamsListRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	accountID := d.Get("account_id").(string)

	list, err := client.TeamsList(context.Background(), accountID, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "HTTP status 404") {
			log.Printf("[INFO] Teams List %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error finding Teams List %q: %s", d.Id(), err)
	}

	d.Set("name", list.Name)
	d.Set("type", list.Type)
	d.Set("description", list.Description)

	listItems, _, err := client.TeamsListItems(context.Background(), accountID, d.Id())
	if err != nil {
		return fmt.Errorf("Error finding Teams List %q: %s", d.Id(), err)
	}
	d.Set("items", convertListItemsToSchema(listItems))

	return nil
}

func resourceCloudflareTeamsListUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	updatedTeamsList := cloudflare.TeamsList{
		ID:          d.Id(),
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Description: d.Get("description").(string),
	}

	log.Printf("[DEBUG] Updating Cloudflare Teams List from struct: %+v", updatedTeamsList)

	accountID := d.Get("account_id").(string)

	teamsList, err := client.UpdateTeamsList(context.Background(), accountID, updatedTeamsList)
	if err != nil {
		return fmt.Errorf("error updating Teams List for account %q: %s", accountID, err)
	}
	if teamsList.ID == "" {
		return fmt.Errorf("failed to find Teams List ID in update response; resource was empty")
	}

	if d.HasChange("items") {
		oldItemsIface, newItemsIface := d.GetChange("items")
		oldItems := expandInterfaceToStringList(oldItemsIface)
		newItems := expandInterfaceToStringList(newItemsIface)
		patchTeamsList := cloudflare.PatchTeamsList{ID: d.Id()}
		setListItemDiff(&patchTeamsList, oldItems, newItems)
		l, err := client.PatchTeamsList(context.Background(), accountID, patchTeamsList)
		if err != nil {
			return fmt.Errorf("error updating Teams List for account %q: %s", accountID, err)
		}

		teamsList.Items = l.Items
	}

	return resourceCloudflareTeamsListRead(d, meta)
}

func resourceCloudflareTeamsListDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	appID := d.Id()
	accountID := d.Get("account_id").(string)

	log.Printf("[DEBUG] Deleting Cloudflare Teams List using ID: %s", appID)

	err := client.DeleteTeamsList(context.Background(), accountID, appID)
	if err != nil {
		return fmt.Errorf("error deleting Teams List for account %q: %s", accountID, err)
	}

	resourceCloudflareTeamsListRead(d, meta)

	return nil
}

func resourceCloudflareTeamsListImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"accountID/teamsListID\"", d.Id())
	}

	accountID, teamsListID := attributes[0], attributes[1]

	log.Printf("[DEBUG] Importing Cloudflare Teams List: id %s for account %s", teamsListID, accountID)

	d.Set("account_id", accountID)
	d.SetId(teamsListID)

	resourceCloudflareTeamsListRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

func setListItemDiff(patchList *cloudflare.PatchTeamsList, oldItems []string, newItems []string) {
	counts := make(map[string]int)
	for _, val := range newItems {
		counts[val] += 1
	}
	for _, val := range oldItems {
		counts[val] -= 1
	}

	for key, val := range counts {
		if val > 0 {
			patchList.Append = append(patchList.Append, cloudflare.TeamsListItem{Value: key})
		}
		if val < 0 {
			patchList.Remove = append(patchList.Remove, key)
		}
	}
}

func convertListItemsToSchema(listItems []cloudflare.TeamsListItem) []string {
	itemValues := []string{}
	// The API returns items in reverse order so we iterate backwards for correct ordering.
	for i := len(listItems) - 1; i >= 0; i-- {
		item := listItems[i]
		itemValues = append(itemValues, item.Value)
	}

	return itemValues
}
