package cloudflare

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudflareIPRanges(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCloudflareIPRangesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCloudflareIPRanges("data.cloudflare_ip_ranges.some"),
				),
			},
		},
	})
}

func testAccCloudflareIPRanges(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		var (
			cidrBlockSize int
			err           error
		)

		if cidrBlockSize, err = strconv.Atoi(a["cidr_blocks.#"]); err != nil {
			return err
		}

		if cidrBlockSize < 10 {
			return fmt.Errorf("cidr_blocks seem suspiciously low: %d", cidrBlockSize)
		}

		var cidrBlocks sort.StringSlice = make([]string, cidrBlockSize)

		for i := range make([]string, cidrBlockSize) {

			block := a[fmt.Sprintf("cidr_blocks.%d", i)]

			if _, _, err := net.ParseCIDR(block); err != nil {
				return fmt.Errorf("malformed CIDR block %s: %s", block, err)
			}

			cidrBlocks[i] = block

		}

		if !sort.IsSorted(cidrBlocks) {
			return fmt.Errorf("unexpected order of cidr_blocks: %s", cidrBlocks)
		}

		return nil
	}
}

const testAccCloudflareIPRangesConfig = `
data "cloudflare_ip_ranges" "some" {}
`
