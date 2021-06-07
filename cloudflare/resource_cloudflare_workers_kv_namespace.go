package cloudflare

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceCloudflareWorkersKVNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareWorkersKVNamespaceCreate,
		Read:   resourceCloudflareWorkersKVNamespaceRead,
		Update: resourceCloudflareWorkersKVNamespaceUpdate,
		Delete: resourceCloudflareWorkersKVNamespaceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareWorkersKVNamespaceImport,
		},

		Schema: map[string]*schema.Schema{
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCloudflareWorkersKVNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	req := &cloudflare.WorkersKVNamespaceRequest{
		Title: d.Get("title").(string),
	}

	log.Printf("[Info] Creating Cloudflare Workers KV Namespace from struct: %+v", req)

	r, err := client.CreateWorkersKVNamespace(context.Background(), req)
	if err != nil {
		return errors.Wrap(err, "error creating workers kv namespace")
	}

	if r.Result.ID == "" {
		return fmt.Errorf("failed to find id in Create response; resource was empty")
	}

	d.SetId(r.Result.ID)

	log.Printf("[INFO] Cloudflare Workers KV Namespace ID: %s", d.Id())

	return nil
}

func resourceCloudflareWorkersKVNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	namespaceID := d.Id()

	resp, err := client.ListWorkersKVNamespaces(context.Background())
	if err != nil {
		return errors.Wrap(err, "error reading workers kv namespaces")
	}

	var namespace cloudflare.WorkersKVNamespace
	for _, r := range resp {
		if r.ID == namespaceID {
			namespace = r
			break
		}
	}

	if namespace.ID == "" {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceCloudflareWorkersKVNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	namespace := &cloudflare.WorkersKVNamespaceRequest{
		Title: d.Get("title").(string),
	}

	log.Printf("[INFO] Updating Cloudflare Workers KV Namespace from struct %+v", namespace)

	_, err := client.UpdateWorkersKVNamespace(context.Background(), d.Id(), namespace)
	if err != nil {
		return errors.Wrap(err, "error updating workers kv namespace")
	}

	return nil
}

func resourceCloudflareWorkersKVNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	log.Printf("[INFO] Deleting Cloudflare Workers KV Namespace with id: %+v", d.Id())

	_, err := client.DeleteWorkersKVNamespace(context.Background(), d.Id())
	if err != nil {
		return errors.Wrap(err, "error deleting workers kv namespace")
	}

	return nil
}

func resourceCloudflareWorkersKVNamespaceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*cloudflare.API)

	namespaces, err := client.ListWorkersKVNamespaces(context.Background())
	var title string

	for _, n := range namespaces {
		if n.ID == d.Id() {
			title = n.Title
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error finding workers kv namespace %q: %s", d.Id(), err)
	}

	d.Set("title", title)
	d.SetId(d.Id())

	return []*schema.ResourceData{d}, nil
}
