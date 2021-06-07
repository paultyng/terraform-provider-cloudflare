package cloudflare

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceCloudflareLogpushJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudflareLogpushJobCreate,
		Read:   resourceCloudflareLogpushJobRead,
		Update: resourceCloudflareLogpushJobUpdate,
		Delete: resourceCloudflareLogpushJobDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudflareLogpushJobImport,
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dataset": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"firewall_events", "http_requests", "spectrum_events"}, false),
			},
			"logpull_options": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination_conf": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ownership_challenge": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getJobFromResource(d *schema.ResourceData) (cloudflare.LogpushJob, error) {
	id := 0

	if d.Id() != "" {
		var err error
		if id, err = strconv.Atoi(d.Id()); err != nil {
			return cloudflare.LogpushJob{}, fmt.Errorf("could not extract Logpush job from resource - invalid identifier (%s): %v", d.Id(), err)
		}
	}

	destConf := d.Get("destination_conf").(string)
	ownershipChallenge := d.Get("ownership_challenge").(string)
	var re = regexp.MustCompile(`^((datadog|splunk)://|s3://.+endpoint=)`)

	if ownershipChallenge == "" && !re.MatchString(destConf) {
		return cloudflare.LogpushJob{}, fmt.Errorf("ownership_challenge must be set for the provided destination_conf")
	}

	job := cloudflare.LogpushJob{
		ID:                 id,
		Enabled:            d.Get("enabled").(bool),
		Name:               d.Get("name").(string),
		Dataset:            d.Get("dataset").(string),
		LogpullOptions:     d.Get("logpull_options").(string),
		DestinationConf:    destConf,
		OwnershipChallenge: ownershipChallenge,
	}

	return job, nil
}

func resourceCloudflareLogpushJobRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	jobID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("could not extract Logpush job from resource - invalid identifier (%s): %v", d.Id(), err)
	}

	job, err := client.LogpushJob(context.Background(), d.Get("zone_id").(string), jobID)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			log.Printf("[INFO] Could not find LogpushJob with id: %q", jobID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error finding logpush job %q: %s", jobID, err)
	}

	if job.ID == 0 {
		d.SetId("")
		return nil
	}

	d.Set("name", job.Name)
	d.Set("enabled", job.Enabled)
	d.Set("logpull_options", job.LogpullOptions)
	d.Set("destination_conf", job.DestinationConf)
	d.Set("ownership_challenge", d.Get("ownership_challenge"))

	return nil
}

func resourceCloudflareLogpushJobCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	job, err := getJobFromResource(d)
	if err != nil {
		return fmt.Errorf("error finding logpush job: %v", err)
	}

	log.Printf("[DEBUG] Creating Cloudflare Logpush Job from struct: %+v", job)

	j, err := client.CreateLogpushJob(context.Background(), d.Get("zone_id").(string), job)
	if err != nil {
		return fmt.Errorf("error creating logpush job: %v", err)
	}
	if j.ID == 0 {
		return fmt.Errorf("failed to find ID in Create response; resource was empty")
	}

	d.SetId(strconv.Itoa(j.ID))

	log.Printf("[INFO] Created Cloudflare Logpush Job ID: %s", d.Id())

	return resourceCloudflareLogpushJobRead(d, meta)
}

func resourceCloudflareLogpushJobUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)

	job, err := getJobFromResource(d)
	if err != nil {
		return fmt.Errorf("error finding logpush job: %v", err)
	}

	log.Printf("[INFO] Updating Cloudflare Logpush Job from struct: %+v", job)

	updateErr := client.UpdateLogpushJob(context.Background(), d.Get("zone_id").(string), job.ID, job)

	if updateErr != nil {
		return fmt.Errorf("error updating logpush job: %+v", job.ID)
	}

	return resourceCloudflareLogpushJobRead(d, meta)
}

func resourceCloudflareLogpushJobDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudflare.API)
	job, err := getJobFromResource(d)
	if err != nil {
		return fmt.Errorf("error finding logpush job: %v", err)
	}

	log.Printf("[DEBUG] Deleting Cloudflare Logpush job from zone :%+v with id: %+v", d.Get("zone_id"), job.ID)

	deleteErr := client.DeleteLogpushJob(context.Background(), d.Get("zone_id").(string), job.ID)
	if deleteErr != nil {
		if strings.Contains(err.Error(), "job not found") {
			log.Printf("[INFO] Could not find logpush job with id: %q", job.ID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error deleting logpush job: %+v", job.ID)
	}

	d.SetId("")

	return nil
}

func resourceCloudflareLogpushJobImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// split the id so we can lookup
	idAttr := strings.SplitN(d.Id(), "/", 2)

	if len(idAttr) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"zoneID/logpushJobID\"", d.Id())
	}

	zoneID, logpushJobID := idAttr[0], idAttr[1]

	log.Printf("[DEBUG] Importing Cloudflare Logpush Job: id %s for zone %s", logpushJobID, zoneID)

	d.Set("zone_id", zoneID)
	d.SetId(logpushJobID)

	resourceCloudflareLogpushJobRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
