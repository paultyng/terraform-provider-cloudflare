package cloudflare

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

var kvNamespaceBindingResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"namespace_id": {
			Type:     schema.TypeString,
			Required: true,
		},
	},
}

var plainTextBindingResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"text": {
			Type:     schema.TypeString,
			Required: true,
		},
	},
}

var secretTextBindingResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"text": {
			Type:      schema.TypeString,
			Required:  true,
			Sensitive: true,
		},
	},
}

var webAssemblyBindingResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"module": {
			Type:     schema.TypeString,
			Required: true,
		},
	},
}

func resourceCloudflareWorkerScript() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareWorkerScriptCreate,
		Read:   resourceCloudflareWorkerScriptRead,
		Update: resourceCloudflareWorkerScriptUpdate,
		Delete: resourceCloudflareWorkerScriptDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareWorkerScriptImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content": {
				Type:     schema.TypeString,
				Required: true,
			},
			"plain_text_binding": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     plainTextBindingResource,
			},
			"secret_text_binding": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     secretTextBindingResource,
			},
			"kv_namespace_binding": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     kvNamespaceBindingResource,
			},
			"webassembly_binding": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     webAssemblyBindingResource,
			},
		},
	}
}

type ScriptData struct {
	// The script id will be the `name` for named script
	// or the `zone_name` for zone-scoped scripts
	ID     string
	Params cloudflare.WorkerRequestParams
}

func getScriptData(d *schema.ResourceData, client *cloudflare.API) (ScriptData, error) {
	scriptName := d.Get("name").(string)

	params := cloudflare.WorkerRequestParams{
		ScriptName: scriptName,
	}

	return ScriptData{
		scriptName,
		params,
	}, nil
}

type ScriptBindings map[string]cloudflare.WorkerBinding

func getWorkerScriptBindings(scriptName string, client *cloudflare.API) (ScriptBindings, error) {
	resp, err := client.ListWorkerBindings(context.Background(), &cloudflare.WorkerRequestParams{ScriptName: scriptName})
	if err != nil {
		return nil, fmt.Errorf("cannot list script bindings: %v", err)
	}

	bindings := make(ScriptBindings, len(resp.BindingList))

	for _, b := range resp.BindingList {
		bindings[b.Name] = b.Binding
	}

	return bindings, nil
}

func parseWorkerBindings(d *schema.ResourceData, bindings ScriptBindings) {
	for _, rawData := range d.Get("kv_namespace_binding").(*schema.Set).List() {
		data := rawData.(map[string]interface{})
		bindings[data["name"].(string)] = cloudflare.WorkerKvNamespaceBinding{
			NamespaceID: data["namespace_id"].(string),
		}
	}

	for _, rawData := range d.Get("plain_text_binding").(*schema.Set).List() {
		data := rawData.(map[string]interface{})
		bindings[data["name"].(string)] = cloudflare.WorkerPlainTextBinding{
			Text: data["text"].(string),
		}
	}

	for _, rawData := range d.Get("secret_text_binding").(*schema.Set).List() {
		data := rawData.(map[string]interface{})
		bindings[data["name"].(string)] = cloudflare.WorkerSecretTextBinding{
			Text: data["text"].(string),
		}
	}

	for _, rawData := range d.Get("webassembly_binding").(*schema.Set).List() {
		data := rawData.(map[string]interface{})
		module := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data["module"].(string)))
		bindings[data["name"].(string)] = cloudflare.WorkerWebAssemblyBinding{
			Module: module,
		}
	}
}

func resourceCloudflareWorkerScriptCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	scriptData, err := getScriptData(d, client)
	if err != nil {
		return err
	}

	// make sure that the worker does not already exist
	r, _ := client.DownloadWorker(context.Background(), &scriptData.Params)
	if r.WorkerScript.Script != "" {
		return fmt.Errorf("script already exists")
	}

	scriptBody := d.Get("content").(string)
	if scriptBody == "" {
		return fmt.Errorf("script content cannot be empty")
	}

	log.Printf("[INFO] Creating Cloudflare Worker Script from struct: %+v", &scriptData.Params)

	bindings := make(ScriptBindings)

	parseWorkerBindings(d, bindings)

	scriptParams := cloudflare.WorkerScriptParams{
		Script:   scriptBody,
		Bindings: bindings,
	}

	_, err = client.UploadWorkerWithBindings(context.Background(), &scriptData.Params, &scriptParams)
	if err != nil {
		return errors.Wrap(err, "error creating worker script")
	}

	d.SetId(scriptData.ID)

	return nil
}

func resourceCloudflareWorkerScriptRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	scriptData, err := getScriptData(d, client)
	if err != nil {
		return err
	}

	r, err := client.DownloadWorker(context.Background(), &scriptData.Params)
	if err != nil {
		// If the resource is deleted, we should set the ID to "" and not
		// return an error according to the terraform spec
		if strings.Contains(err.Error(), "HTTP status 404") {
			d.SetId("")
			return nil
		}

		return errors.Wrap(err,
			fmt.Sprintf("Error reading worker script from API for resource %+v", &scriptData.Params))
	}

	existingBindings := make(ScriptBindings)

	parseWorkerBindings(d, existingBindings)

	bindings, err := getWorkerScriptBindings(d.Get("name").(string), client)
	if err != nil {
		return err
	}

	kvNamespaceBindings := &schema.Set{F: schema.HashResource(kvNamespaceBindingResource)}
	plainTextBindings := &schema.Set{F: schema.HashResource(plainTextBindingResource)}
	secretTextBindings := &schema.Set{F: schema.HashResource(secretTextBindingResource)}
	webAssemblyBindings := &schema.Set{F: schema.HashResource(webAssemblyBindingResource)}

	for name, binding := range bindings {
		switch v := binding.(type) {
		case cloudflare.WorkerKvNamespaceBinding:
			kvNamespaceBindings.Add(map[string]interface{}{
				"name":         name,
				"namespace_id": v.NamespaceID,
			})
		case cloudflare.WorkerPlainTextBinding:
			plainTextBindings.Add(map[string]interface{}{
				"name": name,
				"text": v.Text,
			})
		case cloudflare.WorkerSecretTextBinding:
			value := v.Text
			switch v := existingBindings[name].(type) {
			case cloudflare.WorkerSecretTextBinding:
				value = v.Text
			}
			secretTextBindings.Add(map[string]interface{}{
				"name": name,
				"text": value,
			})
		case cloudflare.WorkerWebAssemblyBinding:
			module, err := ioutil.ReadAll(v.Module)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("cannot read contents of wasm bindings (%s)", name))
			}
			webAssemblyBindings.Add(map[string]interface{}{
				"name":   name,
				"module": base64.StdEncoding.EncodeToString(module),
			})
		}
	}

	if err := d.Set("content", r.Script); err != nil {
		return fmt.Errorf("cannot set content: %v", err)
	}

	if err := d.Set("kv_namespace_binding", kvNamespaceBindings); err != nil {
		return fmt.Errorf("cannot set kv namespace bindings (%s): %v", d.Id(), err)
	}

	if err := d.Set("plain_text_binding", plainTextBindings); err != nil {
		return fmt.Errorf("cannot set plain text bindings (%s): %v", d.Id(), err)
	}

	if err := d.Set("secret_text_binding", secretTextBindings); err != nil {
		return fmt.Errorf("cannot set secret text bindings (%s): %v", d.Id(), err)
	}

	if err := d.Set("webassembly_binding", webAssemblyBindings); err != nil {
		return fmt.Errorf("cannot set webassembly bindings (%s): %v", d.Id(), err)
	}

	return nil
}

func resourceCloudflareWorkerScriptUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	scriptData, err := getScriptData(d, client)
	if err != nil {
		return err
	}

	scriptBody := d.Get("content").(string)
	if scriptBody == "" {
		return fmt.Errorf("script content cannot be empty")
	}

	log.Printf("[INFO] Updating Cloudflare Worker Script from struct: %+v", &scriptData.Params)

	bindings := make(ScriptBindings)

	parseWorkerBindings(d, bindings)

	scriptParams := cloudflare.WorkerScriptParams{
		Script:   scriptBody,
		Bindings: bindings,
	}

	_, err = client.UploadWorkerWithBindings(context.Background(), &scriptData.Params, &scriptParams)
	if err != nil {
		return errors.Wrap(err, "error updating worker script")
	}

	return nil
}

func resourceCloudflareWorkerScriptDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	scriptData, err := getScriptData(d, client)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Deleting Cloudflare Worker Script from struct: %+v", &scriptData.Params)

	_, err = client.DeleteWorker(context.Background(), &scriptData.Params)
	if err != nil {
		// If the resource is already deleted, we should return without an error
		// according to the terraform spec
		if strings.Contains(err.Error(), "HTTP status 404") {
			return nil
		}

		return errors.Wrap(err, "error deleting worker script")
	}

	return nil
}

func resourceCloudflareWorkerScriptImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	scriptID := d.Id()
	_ = d.Set("name", scriptID)

	_ = resourceCloudflareWorkerScriptRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
