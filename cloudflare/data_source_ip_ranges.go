package cloudflare

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	urlIPV4s = "https://www.cloudflare.com/ips-v4"
	urlIPV6s = "https://www.cloudflare.com/ips-v6"
)

func dataSourceCloudflareIPRanges() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCloudflareIPRangesRead,

		Schema: map[string]*schema.Schema{
			"cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ipv4_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ipv6_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"china_ipv4_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"china_ipv6_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceCloudflareIPRangesRead(d *schema.ResourceData, meta interface{}) error {
	ranges, err := cloudflare.IPs()
	if err != nil {
		return fmt.Errorf("failed to fetch Cloudflare IP ranges: %s", err)
	}

	IPv4s := ranges.IPv4CIDRs
	IPv6s := ranges.IPv6CIDRs
	chinaIPv4s := ranges.ChinaIPv4CIDRs
	chinaIPv6s := ranges.ChinaIPv6CIDRs

	sort.Strings(IPv4s)
	sort.Strings(IPv6s)
	sort.Strings(chinaIPv4s)
	sort.Strings(chinaIPv6s)

	all := append([]string{}, IPv4s...)
	all = append(all, IPv6s...)
	sort.Strings(all)

	d.SetId(strconv.Itoa(hashCodeString(strings.Join(all, "|"))))

	if err := d.Set("cidr_blocks", all); err != nil {
		return fmt.Errorf("error setting all cidr blocks: %s", err)
	}

	if err := d.Set("ipv4_cidr_blocks", IPv4s); err != nil {
		return fmt.Errorf("error setting ipv4 cidr blocks: %s", err)
	}

	if err := d.Set("ipv6_cidr_blocks", IPv6s); err != nil {
		return fmt.Errorf("error setting ipv6 cidr blocks: %s", err)
	}

	if err := d.Set("china_ipv4_cidr_blocks", chinaIPv4s); err != nil {
		return fmt.Errorf("error setting china ipv4 cidr blocks: %s", err)
	}

	if err := d.Set("china_ipv6_cidr_blocks", chinaIPv6s); err != nil {
		return fmt.Errorf("error setting china ipv6 cidr blocks: %s", err)
	}

	return nil
}
