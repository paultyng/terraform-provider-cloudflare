package cloudflare

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	testAccProviders map[string]terraform.ResourceProvider
	testAccProvider  *schema.Provider

	// Integration test account ID
	testAccCloudflareAccountID string = "f037e56e89293a057740de681ac9abbe"

	// Integration test account zone ID
	testAccCloudflareZoneID string = "0da42c8d2132a9ddaf714f9e7c920711"
	// Integration test account zone name
	testAccCloudflareZoneName string = "terraform.cfapi.net"

	// Integration test account alternate zone ID
	testAccCloudflareAltZoneID string = "b72110c08e3382597095c29ba7e661ea"
	// Integration test account alternate zone name
	testAccCloudflareAltZoneName string = "terraform2.cfapi.net"
)

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"cloudflare": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

type preCheckFunc = func(*testing.T)

func testAccPreCheck(t *testing.T) {
	testAccPreCheckEmail(t)
	testAccPreCheckApiKey(t)
	testAccPreCheckDomain(t)

	if v := os.Getenv("CLOUDFLARE_ZONE_ID"); v == "" {
		t.Fatal("CLOUDFLARE_ZONE_ID must be set for this acceptance test")
	}
}

func testAccessAccPreCheck(t *testing.T) {
	testAccPreCheckEmail(t)
	testAccPreCheckApiKey(t)
	testAccPreCheckDomain(t)

	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	if zoneID == "" && accountID == "" {
		t.Fatal("CLOUDFLARE_ZONE_ID or CLOUDFLARE_ACCOUNT_ID must be set for this acceptance test")
	}
}

func testAccPreCheckAltDomain(t *testing.T) {
	if v := os.Getenv("CLOUDFLARE_ALT_DOMAIN"); v == "" {
		t.Fatal("CLOUDFLARE_ALT_DOMAIN must be set for this acceptance test")
	}
}

func testAccPreCheckAltZoneID(t *testing.T) {
	if v := os.Getenv("CLOUDFLARE_ALT_ZONE_ID"); v == "" {
		t.Fatal("CLOUDFLARE_ALT_ZONE_ID must be set for this acceptance test")
	}
}

func testAccPreCheckAccount(t *testing.T) {
	if v := os.Getenv("CLOUDFLARE_ACCOUNT_ID"); v == "" {
		t.Fatal("CLOUDFLARE_ACCOUNT_ID must be set for this acceptance test")
	}
}

func testAccPreCheckEmail(t *testing.T) {
	if v := os.Getenv("CLOUDFLARE_EMAIL"); v == "" {
		t.Fatal("CLOUDFLARE_EMAIL must be set for acceptance tests")
	}
}

func testAccPreCheckApiKey(t *testing.T) {
	if v := os.Getenv("CLOUDFLARE_API_KEY"); v == "" {
		t.Fatal("CLOUDFLARE_API_KEY must be set for acceptance tests")
	}
}

func testAccPreCheckApiUserServiceKey(t *testing.T) {
	if v := os.Getenv("CLOUDFLARE_API_USER_SERVICE_KEY"); v == "" {
		t.Fatal("CLOUDFLARE_API_USER_SERVICE_KEY must be set for acceptance tests")
	}
}

func testAccPreCheckDomain(t *testing.T) {
	if v := os.Getenv("CLOUDFLARE_DOMAIN"); v == "" {
		t.Fatal("CLOUDFLARE_DOMAIN must be set for acceptance tests. The domain is used to create and destroy record against.")
	}
}

func testAccPreCheckLogpushToken(t *testing.T) {
	if v := os.Getenv("CLOUDFLARE_LOGPUSH_OWNERSHIP_TOKEN"); v == "" {
		t.Fatal("CLOUDFLARE_LOGPUSH_OWNERSHIP_TOKEN must be set for this acceptance test")
	}
	if v := os.Getenv("CLOUDFLARE_ZONE_ID"); v == "" {
		t.Fatal("CLOUDFLARE_ZONE_ID must be set for this acceptance test")
	}
}

func testAccPreCheckBYOIPPrefix(t *testing.T) {
	if v := os.Getenv("CLOUDFLARE_BYO_IP_PREFIX_ID"); v == "" {
		t.Skip("Skipping acceptance test as CLOUDFLARE_BYO_IP_PREFIX_ID is not set")
	}
}

func generateRandomResourceName() string {
	return acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
}

// skipMagicTransitTestForNonConfiguredDefaultZone will force an acceptance test
// to skip instead of running and failing due to not having setup Magic Transit.
// This will allow those who intentionally want to run the test to do so while
// keeping CI sane.
func skipMagicTransitTestForNonConfiguredDefaultZone(t *testing.T) {
	if os.Getenv("CLOUDFLARE_ZONE_ID") == testAccCloudflareZoneID {
		t.Skipf("Skipping acceptance test as %s is not configured for Magic Transit", testAccCloudflareZoneID)
	}
}
