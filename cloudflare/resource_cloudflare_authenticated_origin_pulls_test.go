package cloudflare

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudflareAuthenticatedOriginPullsGlobal(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_authenticated_origin_pulls.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareAuthenticatedOriginPullsGlobalConfig(zoneID, rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccCloudflareAuthenticatedOriginPullsPerZone(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_authenticated_origin_pulls.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareAuthenticatedOriginPullsPerZoneConfig(zoneID, rnd, "per-zone"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccCloudflareAuthenticatedOriginPullsPerHostname(t *testing.T) {
	t.Parallel()
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	hostname := os.Getenv("CLOUDFLARE_DOMAIN")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_authenticated_origin_pulls.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareAuthenticatedOriginPullsConfig(zoneID, rnd, "per-hostname", hostname),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckCloudflareAuthenticatedOriginPullsGlobalConfig(zoneID, name string) string {
	return fmt.Sprintf(`
  resource "cloudflare_authenticated_origin_pulls" "%[2]s" {
	  zone_id        = "%[1]s"
	  enabled = true
  }`, zoneID, name)
}

func testAccCheckCloudflareAuthenticatedOriginPullsPerZoneConfig(zoneID, name, aopType string) string {
	return fmt.Sprintf(`
		resource "cloudflare_authenticated_origin_pulls_certificate" "%[1]s" {
		  zone_id = "%[2]s"
		  certificate = "-----BEGIN CERTIFICATE-----\nMIIEsTCCA5mgAwIBAgISA53fvg2BvlK2QXSkdZewcNo4MA0GCSqGSIb3DQEBCwUA\nMEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD\nExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0yMDA2MjUyMTAzNDdaFw0y\nMDA5MjMyMTAzNDdaMB4xHDAaBgNVBAMTE3RlcnJhZm9ybS5jZmFwaS5uZXQwdjAQ\nBgcqhkjOPQIBBgUrgQQAIgNiAASBYi00+H4E7uUeogweuutTWvuAz8TC6ClQYemH\nCGA6xKrvSgWwjhvVM9joPhGlbUDbINKhVMdZd7q3DgBinVu9GjjKf1Ajxnr6nEsK\naq37tZmtUFawbqnJHAI+O3uTan+jggJpMIICZTAOBgNVHQ8BAf8EBAMCB4AwHQYD\nVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0O\nBBYEFACS0TnEhBjGvOG127Yn2O1/UCOoMB8GA1UdIwQYMBaAFKhKamMEfd265tE5\nt6ZFZe/zqOyhMG8GCCsGAQUFBwEBBGMwYTAuBggrBgEFBQcwAYYiaHR0cDovL29j\nc3AuaW50LXgzLmxldHNlbmNyeXB0Lm9yZzAvBggrBgEFBQcwAoYjaHR0cDovL2Nl\ncnQuaW50LXgzLmxldHNlbmNyeXB0Lm9yZy8wHgYDVR0RBBcwFYITdGVycmFmb3Jt\nLmNmYXBpLm5ldDBMBgNVHSAERTBDMAgGBmeBDAECATA3BgsrBgEEAYLfEwEBATAo\nMCYGCCsGAQUFBwIBFhpodHRwOi8vY3BzLmxldHNlbmNyeXB0Lm9yZzCCAQUGCisG\nAQQB1nkCBAIEgfYEgfMA8QB3AF6nc/nfVsDntTZIfdBJ4DJ6kZoMhKESEoQYdZaB\ncUVYAAABcu2CH2EAAAQDAEgwRgIhAK4dA41POH3dCyi/5CN98MbBRAl8a6LyeQls\nJyZ+y1sIAiEAoMtsQKVgf8APT7/DGj/b4OzMO6EBKWcrGkZpTi7nyyQAdgCyHgXM\ni6LNiiBOh2b5K7mKJSBna9r6cOeySVMt74uQXgAAAXLtgh9PAAAEAwBHMEUCIQC1\nnxSRx2fcqG8gw5z0QK5PGktggqIulg2Jrwr20ZfXKwIgGxNlOEucj1t71h4PaLuy\nnBigJo57ztE5t56o0dlUOzEwDQYJKoZIhvcNAQELBQADggEBACy8MS07SVQLMeGK\na3E7jn7mQciQkt063tnIYbvnUTeYQZVe1Rzk6Tm9GyQoL7MIFAvTHbsB9bNzIRrl\nubefCn4s6PHnVyDGiPY/yQgGjymXyxcsfwVnc3XO3i6N8AN1MQuKMx+Kx69sHVpa\nKq9Qlu1HlStlX/eUWMcoDk1WaCJ7xm17npvdWDweDg71Qlgnl6ukggN+cQwKepw5\n4tMnqmhrzMH+xnH2dTIQ10lgB31AlwBSbOUymhg8XN+BIeXW54mBjdxkBd++7+0q\nv7oFDmljpwQSAC2BMU8ah7lwRhQxgTrG0z10Qdje1CJ8ylRHArIeISlx+jBAwKQh\nulkb7Ck=\n-----END CERTIFICATE-----\n"
		  private_key = "-----BEGIN EC PRIVATE KEY-----\nMIGkAgEBBDD+Um5v/lCBTCvHEcZlLnSz6XX1fEOk5FxfUdiQvcY5x6WXuu3dDgDf\nvKIS0J6AsxygBwYFK4EEACKhZANiAASBYi00+H4E7uUeogweuutTWvuAz8TC6ClQ\nYemHCGA6xKrvSgWwjhvVM9joPhGlbUDbINKhVMdZd7q3DgBinVu9GjjKf1Ajxnr6\nnEsKaq37tZmtUFawbqnJHAI+O3uTan8=\n-----END EC PRIVATE KEY-----\n"
		  type = "%[3]s"
		}
		resource "cloudflare_authenticated_origin_pulls" "%[1]s" {
		  zone_id = "%[2]s"
		  authenticated_origin_pulls_certificate = "${cloudflare_authenticated_origin_pulls_certificate.%[1]s.id}"
		  enabled = true
		}`, name, zoneID, aopType)
}

func testAccCheckCloudflareAuthenticatedOriginPullsConfig(zoneID, name, aopType, hostname string) string {
	return fmt.Sprintf(`
		resource "cloudflare_authenticated_origin_pulls_certificate" "%[1]s" {
		  zone_id = "%[2]s"
		  certificate = "-----BEGIN CERTIFICATE-----\nMIIEsTCCA5mgAwIBAgISA53fvg2BvlK2QXSkdZewcNo4MA0GCSqGSIb3DQEBCwUA\nMEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD\nExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0yMDA2MjUyMTAzNDdaFw0y\nMDA5MjMyMTAzNDdaMB4xHDAaBgNVBAMTE3RlcnJhZm9ybS5jZmFwaS5uZXQwdjAQ\nBgcqhkjOPQIBBgUrgQQAIgNiAASBYi00+H4E7uUeogweuutTWvuAz8TC6ClQYemH\nCGA6xKrvSgWwjhvVM9joPhGlbUDbINKhVMdZd7q3DgBinVu9GjjKf1Ajxnr6nEsK\naq37tZmtUFawbqnJHAI+O3uTan+jggJpMIICZTAOBgNVHQ8BAf8EBAMCB4AwHQYD\nVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0O\nBBYEFACS0TnEhBjGvOG127Yn2O1/UCOoMB8GA1UdIwQYMBaAFKhKamMEfd265tE5\nt6ZFZe/zqOyhMG8GCCsGAQUFBwEBBGMwYTAuBggrBgEFBQcwAYYiaHR0cDovL29j\nc3AuaW50LXgzLmxldHNlbmNyeXB0Lm9yZzAvBggrBgEFBQcwAoYjaHR0cDovL2Nl\ncnQuaW50LXgzLmxldHNlbmNyeXB0Lm9yZy8wHgYDVR0RBBcwFYITdGVycmFmb3Jt\nLmNmYXBpLm5ldDBMBgNVHSAERTBDMAgGBmeBDAECATA3BgsrBgEEAYLfEwEBATAo\nMCYGCCsGAQUFBwIBFhpodHRwOi8vY3BzLmxldHNlbmNyeXB0Lm9yZzCCAQUGCisG\nAQQB1nkCBAIEgfYEgfMA8QB3AF6nc/nfVsDntTZIfdBJ4DJ6kZoMhKESEoQYdZaB\ncUVYAAABcu2CH2EAAAQDAEgwRgIhAK4dA41POH3dCyi/5CN98MbBRAl8a6LyeQls\nJyZ+y1sIAiEAoMtsQKVgf8APT7/DGj/b4OzMO6EBKWcrGkZpTi7nyyQAdgCyHgXM\ni6LNiiBOh2b5K7mKJSBna9r6cOeySVMt74uQXgAAAXLtgh9PAAAEAwBHMEUCIQC1\nnxSRx2fcqG8gw5z0QK5PGktggqIulg2Jrwr20ZfXKwIgGxNlOEucj1t71h4PaLuy\nnBigJo57ztE5t56o0dlUOzEwDQYJKoZIhvcNAQELBQADggEBACy8MS07SVQLMeGK\na3E7jn7mQciQkt063tnIYbvnUTeYQZVe1Rzk6Tm9GyQoL7MIFAvTHbsB9bNzIRrl\nubefCn4s6PHnVyDGiPY/yQgGjymXyxcsfwVnc3XO3i6N8AN1MQuKMx+Kx69sHVpa\nKq9Qlu1HlStlX/eUWMcoDk1WaCJ7xm17npvdWDweDg71Qlgnl6ukggN+cQwKepw5\n4tMnqmhrzMH+xnH2dTIQ10lgB31AlwBSbOUymhg8XN+BIeXW54mBjdxkBd++7+0q\nv7oFDmljpwQSAC2BMU8ah7lwRhQxgTrG0z10Qdje1CJ8ylRHArIeISlx+jBAwKQh\nulkb7Ck=\n-----END CERTIFICATE-----\n"
		  private_key = "-----BEGIN EC PRIVATE KEY-----\nMIGkAgEBBDD+Um5v/lCBTCvHEcZlLnSz6XX1fEOk5FxfUdiQvcY5x6WXuu3dDgDf\nvKIS0J6AsxygBwYFK4EEACKhZANiAASBYi00+H4E7uUeogweuutTWvuAz8TC6ClQ\nYemHCGA6xKrvSgWwjhvVM9joPhGlbUDbINKhVMdZd7q3DgBinVu9GjjKf1Ajxnr6\nnEsKaq37tZmtUFawbqnJHAI+O3uTan8=\n-----END EC PRIVATE KEY-----\n"
		  type = "%[3]s"
		}
		resource "cloudflare_authenticated_origin_pulls" "%[1]s" {
		  zone_id = "%[2]s"
		  authenticated_origin_pulls_certificate = "${cloudflare_authenticated_origin_pulls_certificate.%[1]s.id}"
		  hostname = "%[4]s"
		  enabled = true
		}`, name, zoneID, aopType, name+"."+hostname)
}
