package cloudflare

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflarePageRule_Basic(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-basic.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigBasic(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
					testAccCheckCloudflarePageRuleAttributesBasic(&pageRule),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "zone_id", zoneID),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "target", fmt.Sprintf("%s/", target)),
				),
			},
		},
	})
}

func TestAccCloudflarePageRule_FullySpecified(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-fully-specified.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigFullySpecified(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "zone_id", zoneID),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "target", fmt.Sprintf("%s/", target)),
				),
			},
		},
	})
}

func TestAccCloudflarePageRule_ForwardingOnly(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-fwd-only.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigForwardingOnly(zoneID, target, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "zone_id", zoneID),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "target", fmt.Sprintf("%s/", target)),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test",
						"actions.0.forwarding_url.0.url",
						fmt.Sprintf("http://%s/forward", domain),
					),
				),
			},
		},
	})
}

func TestAccCloudflarePageRule_ForwardingAndOthers(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-fwd-others.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigForwardingAndOthers(zoneID, target, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "zone_id", zoneID),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "target", target),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "target", fmt.Sprintf("%s/", target)),
				),

				ExpectError: regexp.MustCompile("\"forwarding_url\" cannot be set with any other actions"),
			},
		},
	})
}

func TestAccCloudflarePageRule_Updated(t *testing.T) {
	var before, after cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-updated.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigBasic(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &before),
				),
			},
			{
				Config: testAccCheckCloudflarePageRuleConfigNewValue(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &after),
					testAccCheckCloudflarePageRuleAttributesUpdated(&after),
					testAccCheckCloudflarePageRuleIDUnchanged(&before, &after),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "target", fmt.Sprintf("%s/updated", target)),
				),
			},
		},
	})
}

func TestAccCloudflarePageRule_CreateAfterManualDestroy(t *testing.T) {
	var before, after cloudflare.PageRule
	var initialID string
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-updated.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigBasic(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &before),
					testAccManuallyDeletePageRule("cloudflare_page_rule.test", &initialID),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCheckCloudflarePageRuleConfigNewValue(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &after),
					testAccCheckCloudflarePageRuleRecreated(&before, &after),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "target", fmt.Sprintf("%s/updated", target)),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "actions.0.always_online", "off"),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "actions.0.browser_check", "on"),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "actions.0.rocket_loader", "on"),
					resource.TestCheckResourceAttr(
						"cloudflare_page_rule.test", "actions.0.ssl", "strict"),
				),
			},
		},
	})
}

func TestAccCloudflarePageRule_UpdatingZoneForcesNewResource(t *testing.T) {
	var before, after cloudflare.PageRule
	oldZoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	newZoneID := os.Getenv("CLOUDFLARE_ALT_ZONE_ID")
	oldTarget := fmt.Sprintf("test-updating-zone-value.%s", os.Getenv("CLOUDFLARE_DOMAIN"))
	newTarget := fmt.Sprintf("test-updating-zone-value.%s", os.Getenv("CLOUDFLARE_ALT_DOMAIN"))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAltDomain(t)
			testAccPreCheckAltZoneID(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigBasic(oldZoneID, oldTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &before),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "zone_id", oldZoneID),
				),
			},
			{
				Config: testAccCheckCloudflarePageRuleConfigBasic(newZoneID, newTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &after),
					testAccCheckCloudflarePageRuleRecreated(&before, &after),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "zone_id", newZoneID),
				),
			},
		},
	})
}

func TestAccCloudflarePageRuleMinifyAction(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-action-minify.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigMinify(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.minify.0.css", "on"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.minify.0.js", "off"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.minify.0.html", "on"),
				),
			},
		},
	})
}

func TestTranformForwardingURL(t *testing.T) {
	key, val, err := transformFromCloudflarePageRuleAction(&cloudflare.PageRuleAction{
		ID: "forwarding_url",
		Value: map[string]interface{}{
			"url":         "http://test.com/forward",
			"status_code": 302,
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error transforming page rule action: %s", err)
	}

	if key != "forwarding_url" {
		t.Fatalf("Unexpected key transforming page rule action. Expected \"forwarding_url\", got \"%s\"", key)
	}

	// the transformed value for a forwarding_url should be [{url: "", "status_code": 302}] (single item slice where the
	// element in the slice is a map)
	if sl, isSlice := val.([]interface{}); !isSlice {
		t.Fatalf("Unexpected value type from transforming page rule action. Expected slice, got %s", reflect.TypeOf(val).Kind())
	} else if len(sl) != 1 {
		t.Fatalf("Unexpected slice length after transforming page rule action. Expected 1, got %d", len(sl))
	} else if _, isMap := sl[0].(map[string]interface{}); !isMap {
		t.Fatalf("Unexpected type in slice after tranforming page rule action. Expected map[string]interface{}, got %s", reflect.TypeOf(sl[0]).Kind())
	}
}

// This test ensures there is no crash while encountering a nil query_string section, which may happen when updating
// existing Page Rule that didn't have this value set previously
func TestCacheKeyFieldsNilValue(t *testing.T) {
	pageRuleAction, err := transformToCloudflarePageRuleAction(
		"cache_key_fields",
		[]interface{}{
			map[string]interface{}{
				"cookie": []interface{}{
					map[string]interface{}{
						"include":        schema.NewSet(schema.HashString, []interface{}{}),
						"check_presence": schema.NewSet(schema.HashString, []interface{}{"next-i18next"}),
					},
				},
				"header": []interface{}{
					map[string]interface{}{
						"check_presence": schema.NewSet(schema.HashString, []interface{}{}),
						"exclude":        schema.NewSet(schema.HashString, []interface{}{}),
						"include":        schema.NewSet(schema.HashString, []interface{}{"x-forwarded-host"}),
					},
				},
				"host": []interface{}{
					map[string]interface{}{
						"resolved": false,
					},
				},
				"query_string": []interface{}{
					interface{}(nil),
				},
				"user": []interface{}{
					map[string]interface{}{
						"device_type": true,
						"geo":         true,
						"lang":        true,
					},
				},
			},
		},
		nil,
	)

	if err != nil {
		t.Fatalf("Unexpected error transforming page rule action: %s", err)
	}

	if !reflect.DeepEqual(pageRuleAction.Value.(map[string]interface{})["query_string"], map[string]interface{}{"include": "*"}) {
		t.Fatalf("Unexpected transformToCloudflarePageRuleAction result, expected %#v, got %#v", map[string]interface{}{"include": "*"}, pageRuleAction.Value.(map[string]interface{})["query_string"])
	}
}

func TestAccCloudflarePageRule_CreatesBrowserCacheTTLIntegerValues(t *testing.T) {
	var pageRule cloudflare.PageRule
	testAccRunResourceTestSteps(t, []resource.TestStep{
		{
			Config: buildPageRuleConfig("test", "browser_cache_ttl = 1"),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
				testAccCheckCloudflarePageRuleHasAction(&pageRule, "browser_cache_ttl", float64(1)),
				resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.browser_cache_ttl", "1"),
			),
		},
	})
}

func TestAccCloudflarePageRule_CreatesBrowserCacheTTLThatRespectsExistingHeaders(t *testing.T) {
	var pageRule cloudflare.PageRule
	testAccRunResourceTestSteps(t, []resource.TestStep{
		{
			Config: buildPageRuleConfig("test", "browser_cache_ttl = 0"),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
				resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.browser_cache_ttl", "0"),
				testAccCheckCloudflarePageRuleHasAction(&pageRule, "browser_cache_ttl", float64(0)),
			),
		},
	})
}

func TestAccCloudflarePageRule_UpdatesBrowserCacheTTLToSameValue(t *testing.T) {
	var pageRule cloudflare.PageRule
	testAccRunResourceTestSteps(t, []resource.TestStep{
		{
			Config: buildPageRuleConfig("test", "browser_cache_ttl = 1"),
		},
		{
			Config: buildPageRuleConfig("test", `browser_cache_ttl = 1
browser_check = "on"`),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
				testAccCheckCloudflarePageRuleHasAction(&pageRule, "browser_cache_ttl", float64(1)),
				resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.browser_cache_ttl", "1"),
			),
		},
	})
}

func TestAccCloudflarePageRule_UpdatesBrowserCacheTTLThatRespectsExistingHeaders(t *testing.T) {
	var pageRule cloudflare.PageRule
	testAccRunResourceTestSteps(t, []resource.TestStep{
		{
			Config: buildPageRuleConfig("test", "browser_cache_ttl = 1"),
		},
		{
			Config: buildPageRuleConfig("test", "browser_cache_ttl = 0"),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
				testAccCheckCloudflarePageRuleHasAction(&pageRule, "browser_cache_ttl", float64(0)),
				resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.browser_cache_ttl", "0"),
			),
		},
	})
}

func TestAccCloudflarePageRule_DeletesBrowserCacheTTLThatRespectsExistingHeaders(t *testing.T) {
	var pageRule cloudflare.PageRule
	testAccRunResourceTestSteps(t, []resource.TestStep{
		{
			Config: buildPageRuleConfig("test", "browser_cache_ttl = 0"),
		},
		{
			Config: buildPageRuleConfig("test", `browser_check = "on"`),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
				resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.browser_cache_ttl", ""),
			),
		},
	})
}

func TestAccCloudflarePageRuleEdgeCacheTTLNotClobbered(t *testing.T) {
	var before, after cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-edge-cache-ttl-not-clobbered.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigWithEdgeCacheTtl(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &before),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.edge_cache_ttl", "10"),
				),
			},
			{
				Config: testAccCheckCloudflarePageRuleConfigWithEdgeCacheTtlAndAlwaysOnline(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &after),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.edge_cache_ttl", "10"),
				),
			},
		},
	})
}

func TestAccCloudflarePageRuleCacheKeyFieldsBasic(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-cache-key-fields.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigCacheKeyFields(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.cookie.0.check_presence.#", "1"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.cookie.0.include.#", "1"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.header.0.check_presence.#", "1"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.header.0.include.#", "1"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.host.0.resolved", "true"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.query_string.0.exclude.#", "1"),
				),
			},
		},
	})
}

func TestAccCloudflarePageRuleCacheKeyFieldsIgnoreQueryStringOrdering(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	pageRuleTarget := fmt.Sprintf("%s.%s", rnd, domain)
	resourceName := fmt.Sprintf("cloudflare_page_rule.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigCacheKeyFieldsWithUnorderedEntries(zoneID, rnd, pageRuleTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists(resourceName, &pageRule),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.cookie.0.check_presence.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.cookie.0.include.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.header.0.check_presence.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.header.0.include.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.host.0.resolved", "true"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.query_string.0.include.#", "7"),
				),
			},
		},
	})
}

func TestAccCloudflarePageRuleCacheKeyFieldsExcludeAllQueryString(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	pageRuleTarget := fmt.Sprintf("%s.%s", rnd, domain)
	resourceName := fmt.Sprintf("cloudflare_page_rule.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigCacheKeyFieldsIgnoreAllQueryString(zoneID, rnd, pageRuleTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists(resourceName, &pageRule),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.cookie.0.check_presence.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.cookie.0.include.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.header.0.check_presence.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.header.0.include.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.host.0.resolved", "true"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.query_string.0.ignore", "true"),
				),
			},
		},
	})
}

func TestAccCloudflarePageRuleCacheKeyFieldsExcludeMultipleValuesQueryString(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	pageRuleTarget := fmt.Sprintf("%s.%s", rnd, domain)
	resourceName := fmt.Sprintf("cloudflare_page_rule.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigCacheKeyFieldsExcludeMultipleValuesQueryString(zoneID, rnd, pageRuleTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists(resourceName, &pageRule),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.cookie.0.check_presence.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.cookie.0.include.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.header.0.check_presence.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.header.0.include.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.host.0.resolved", "true"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.query_string.0.exclude.#", "2"),
				),
			},
		},
	})
}

func TestAccCloudflarePageRuleCacheKeyFieldsNoQueryStringValuesDefined(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-cache-key-fields.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigCacheKeyFieldsNoQueryStringValuesDefined(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.header.0.exclude.#", "1"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.host.0.resolved", "false"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.user.0.device_type", "true"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.user.0.geo", "true"),
				),
			},
		},
	})
}

func TestAccCloudflarePageRuleCacheKeyFieldsIncludeAllQueryStringValues(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("test-cache-key-fields.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigCacheKeyFieldsIncludeAllQueryStringValues(zoneID, target),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists("cloudflare_page_rule.test", &pageRule),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.header.0.exclude.#", "1"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.host.0.resolved", "false"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.user.0.device_type", "true"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.user.0.geo", "true"),
					resource.TestCheckResourceAttr("cloudflare_page_rule.test", "actions.0.cache_key_fields.0.query_string.0.ignore", "false"),
				),
			},
		},
	})
}

func TestAccCloudflarePageRuleCacheKeyFieldsIncludeMultipleValuesQueryString(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	pageRuleTarget := fmt.Sprintf("%s.%s", rnd, domain)
	resourceName := fmt.Sprintf("cloudflare_page_rule.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigCacheKeyFieldsIncludeMultipleValuesQueryString(zoneID, rnd, pageRuleTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists(resourceName, &pageRule),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.cookie.0.check_presence.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.cookie.0.include.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.header.0.check_presence.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.header.0.include.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.host.0.resolved", "true"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.cache_key_fields.0.query_string.0.include.#", "2"),
				),
			},
		},
	})
}

func TestAccCloudflarePageRuleCacheTTLByStatus(t *testing.T) {
	var pageRule cloudflare.PageRule
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_page_rule.%s", rnd)
	target := fmt.Sprintf("test-cache-ttl-by-status.%s", domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflarePageRuleConfigCacheTTLByStatus(zoneID, target, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflarePageRuleExists(name, &pageRule),
				),
			},
		},
	})
}

func testAccCheckCloudflarePageRuleRecreated(before, after *cloudflare.PageRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID == after.ID {
			return fmt.Errorf("Expected change of PageRule Ids, but both were %v", before.ID)
		}
		return nil
	}
}

func testAccCheckCloudflarePageRuleIDUnchanged(before, after *cloudflare.PageRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID != after.ID {
			return fmt.Errorf("ID should not change suring in place update, but got change %s -> %s", before.ID, after.ID)
		}
		return nil
	}
}

func testAccCheckCloudflarePageRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_page_rule" {
			continue
		}

		_, err := client.PageRule(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("PageRule still exists")
		}
	}

	return nil
}

func testAccCheckCloudflarePageRuleAttributesBasic(pageRule *cloudflare.PageRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		// check the api only has attributes we set non-empty values for
		// this covers on/off attribute types and setting enum-type strings

		actionMap := pageRuleActionsToMap(pageRule.Actions)

		if val, ok := actionMap["always_online"]; ok {
			if _, ok := val.(string); !ok || val != "on" { // lots of booleans get mapped to on/off at api
				return fmt.Errorf("'always_online' not specified correctly at api, found: '%v'", val)
			}
		} else {
			return fmt.Errorf("'always_online' not specified at api")
		}

		if val, ok := actionMap["ssl"]; ok {
			if _, ok := val.(string); !ok || val != "flexible" {
				return fmt.Errorf("'ssl' not specified correctly at api, found: %q", val)
			}
		} else {
			return fmt.Errorf("'ssl' not specified at api")
		}

		if len(pageRule.Actions) != 2 {
			return fmt.Errorf("api should only have attributes we set non-empty (%d) but got %d: %#v",
				2, len(pageRule.Actions), pageRule.Actions)
		}

		return nil
	}
}

func testAccCheckCloudflarePageRuleAttributesUpdated(pageRule *cloudflare.PageRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		actionMap := pageRuleActionsToMap(pageRule.Actions)

		if _, ok := actionMap["disable_apps"]; ok {
			return fmt.Errorf("'disable_apps' found at api, but we should have removed it")
		}

		if val, ok := actionMap["always_online"]; ok {
			if _, ok := val.(string); !ok || val != "off" { // lots of booleans get mapped to on/off at api
				return fmt.Errorf("'always_online' not specified correctly at api, found: '%v'", val)
			}
		} else {
			return fmt.Errorf("'always_online' not specified at api")
		}

		if val, ok := actionMap["browser_check"]; ok {
			if _, ok := val.(string); !ok || val != "on" { // lots of booleans get mapped to on/off at api
				return fmt.Errorf("'browser_check' not specified correctly at api, found: '%v'", val)
			}
		} else {
			return fmt.Errorf("'browser_check' not specified at api")
		}

		if val, ok := actionMap["ssl"]; ok {
			if _, ok := val.(string); !ok || val != "strict" {
				return fmt.Errorf("'ssl' not specified correctly at api, found: %q", val)
			}
		} else {
			return fmt.Errorf("'ssl' not specified at api")
		}

		if val, ok := actionMap["rocket_loader"]; ok {
			if _, ok := val.(string); !ok || val != "on" {
				return fmt.Errorf("'rocket_loader' not specified correctly at api, found: %q", val)
			}
		} else {
			return fmt.Errorf("'rocket_loader' not specified at api")
		}

		if len(pageRule.Actions) != 4 {
			return fmt.Errorf("api should only have attributes we set non-empty (%d) but got %d: %#v",
				4, len(pageRule.Actions), pageRule.Actions)
		}

		return nil
	}
}

func testAccCheckCloudflarePageRuleExists(n string, pageRule *cloudflare.PageRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No PageRule ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		foundPageRule, err := client.PageRule(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundPageRule.ID != rs.Primary.ID {
			return fmt.Errorf("PageRule not found")
		}

		*pageRule = foundPageRule

		return nil
	}
}

func testAccManuallyDeletePageRule(name string, initialID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		*initialID = rs.Primary.ID
		err := client.DeletePageRule(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckCloudflarePageRuleConfigMinify(zoneID, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s"
	actions {
		minify {
			js = "off"
			css = "on"
			html = "on"
		}
	}
}`, zoneID, target)
}

func testAccCheckCloudflarePageRuleConfigBasic(zoneID, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s"
	actions {
		always_online = "on"
		ssl = "flexible"
	}
}`, zoneID, target)
}

func testAccCheckCloudflarePageRuleConfigNewValue(zoneID, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s/updated"
	actions {
		always_online = "off"
		browser_check = "on"
		ssl = "strict"
		rocket_loader = "on"
	}
}`, zoneID, target)
}

func testAccCheckCloudflarePageRuleConfigFullySpecified(zoneID, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s"
	actions {
		always_online = "on"
		browser_check = "on"
		browser_cache_ttl = 0
		email_obfuscation = "on"
		ip_geolocation = "on"
		server_side_exclude = "on"
		disable_apps = true
		disable_performance = true
		disable_security = true
		cache_level = "bypass"
		security_level = "essentially_off"
		ssl = "flexible"
	}
}`, zoneID, target)
}

func testAccCheckCloudflarePageRuleConfigForwardingOnly(zoneID, target, zoneName string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s"
	actions {
		// on/off options cannot even be set to off without causing error
		forwarding_url {
			url = "http://%s/forward"
			status_code = 301
		}
	}
}`, zoneID, target, zoneName)
}

func testAccCheckCloudflarePageRuleConfigForwardingAndOthers(zoneID, target, zoneName string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s"
	actions {
		disable_security = true
		forwarding_url {
			url = "http://%s/forward"
			status_code = 301
		}
	}
}`, zoneID, target, zoneName)
}

func testAccCheckCloudflarePageRuleConfigWithEdgeCacheTtl(zoneID, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s"
	actions {
		always_online = "on"
		ssl = "flexible"
		edge_cache_ttl = 10
	}
}`, zoneID, target)
}

func testAccCheckCloudflarePageRuleConfigWithEdgeCacheTtlAndAlwaysOnline(zoneID, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s"
	actions {
		always_online = "on"
		edge_cache_ttl = 10
	}
}`, zoneID, target)
}

func testAccCheckCloudflarePageRuleConfigCacheKeyFields(zoneID, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s"
	actions {
		cache_key_fields {
			cookie {
				check_presence = ["cookie_presence"]
				include = ["cookie_include"]
			}
			header {
				check_presence = ["header_presence"]
				include = ["header_include"]
			}
			host {
				resolved = true
			}
			query_string {
				exclude = ["qs_exclude"]
			}
			user {}
		}
	}
}`, zoneID, target)
}

func testAccCheckCloudflarePageRuleConfigCacheKeyFieldsWithUnorderedEntries(zoneID, rnd, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "%[3]s" {
	zone_id = "%[1]s"
	target = "%[3]s"
	actions {
		cache_key_fields {
			cookie {
				check_presence = ["cookie_presence"]
				include = ["cookie_include"]
			}
			header {
				check_presence = ["header_presence"]
				include = ["header_include"]
			}
			host {
				resolved = true
			}
			query_string {
				include = [
          "test.anothertest",
          "test.regiontest",
          "test.devicetest",
          "test.testthis",
          "test.hello",
          "test.segmenttest",
          "test.usertype"
				]
			}
			user {}
		}
	}
}`, zoneID, target, rnd)
}

func testAccCheckCloudflarePageRuleConfigCacheKeyFieldsIgnoreAllQueryString(zoneID, rnd, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "%[3]s" {
	zone_id = "%[1]s"
	target = "%[3]s"
	actions {
		cache_key_fields {
			cookie {
				check_presence = ["cookie_presence"]
				include = ["cookie_include"]
			}
			header {
				check_presence = ["header_presence"]
				include = ["header_include"]
			}
			host {
				resolved = true
			}
			query_string {
				ignore = true
			}
			user {}
		}
	}
}`, zoneID, target, rnd)
}

func testAccCheckCloudflarePageRuleConfigCacheKeyFieldsExcludeMultipleValuesQueryString(zoneID, rnd, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "%[3]s" {
	zone_id = "%[1]s"
	target = "%[3]s"
	actions {
		cache_key_fields {
			cookie {
				check_presence = ["cookie_presence"]
				include = ["cookie_include"]
			}
			header {
				check_presence = ["header_presence"]
				include = ["header_include"]
			}
			host {
				resolved = true
			}
			query_string {
				exclude = ["query1", "query2"]
			}
			user {}
		}
	}
}`, zoneID, target, rnd)
}

func testAccCheckCloudflarePageRuleConfigCacheKeyFieldsNoQueryStringValuesDefined(zoneID, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s"
	actions {
		cache_key_fields {
			cookie {}
			header {
				exclude = ["origin"]
			}
			host {}
			query_string {}
			user {
				device_type = true
				geo = true
			}
		}
	}
}`, zoneID, target)
}

func testAccCheckCloudflarePageRuleConfigCacheKeyFieldsIncludeAllQueryStringValues(zoneID, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "test" {
	zone_id = "%s"
	target = "%s"
	actions {
		cache_key_fields {
			cookie {}
			header {
				exclude = ["origin"]
			}
			host {}
			query_string {
				ignore = false
			}
			user {
				device_type = true
				geo = true
			}
		}
	}
}`, zoneID, target)
}

func testAccCheckCloudflarePageRuleConfigCacheKeyFieldsIncludeMultipleValuesQueryString(zoneID, rnd, target string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "%[3]s" {
	zone_id = "%[1]s"
	target = "%[3]s"
	actions {
		cache_key_fields {
			cookie {
				check_presence = ["cookie_presence"]
				include = ["cookie_include"]
			}
			header {
				check_presence = ["header_presence"]
				include = ["header_include"]
			}
			host {
				resolved = true
			}
			query_string {
				include = ["query1", "query2"]
			}
			user {}
		}
	}
}`, zoneID, target, rnd)
}

func testAccCheckCloudflarePageRuleConfigCacheTTLByStatus(zoneID, target, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_page_rule" "%[3]s" {
	zone_id = "%[1]s"
	target = "%[2]s"
	actions {
		cache_ttl_by_status {
			codes = "200-299"
			ttl = 300
		}
		cache_ttl_by_status {
			codes = "300-399"
			ttl = 60
		}
		cache_ttl_by_status {
			codes = "400-403"
			ttl = -1
		}
		cache_ttl_by_status {
			codes = "404"
			ttl = 30
		}
		cache_ttl_by_status {
			codes = "405-499"
			ttl = -1
		}
		cache_ttl_by_status {
			codes = "500-599"
			ttl = 0
		}
	}
}`, zoneID, target, rnd)
}
func buildPageRuleConfig(resourceName string, actions string) string {
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	target := fmt.Sprintf("terraform-test.%s", domain)

	return fmt.Sprintf(`
		resource "cloudflare_page_rule" "%s" {
			zone_id = "%s"
			target = "%s"
			actions {
				%s
			}
		}`,
		resourceName,
		zoneID,
		target,
		actions)
}

func testAccRunResourceTestSteps(t *testing.T, testSteps []resource.TestStep) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflarePageRuleDestroy,
		Steps:        testSteps,
	})
}

func testAccCheckCloudflarePageRuleHasAction(pageRule *cloudflare.PageRule, key string, value interface{}) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, pageRuleAction := range pageRule.Actions {
			if pageRuleAction.ID == key && pageRuleAction.Value == value {
				return nil
			}
		}
		return fmt.Errorf("cloudflare page rule action not found %#v:%#v\nAction State\n%#v", key, value, pageRule.Actions)
	}
}
