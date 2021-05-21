---
layout: "cloudflare"
page_title: "Cloudflare: cloudflare_zone"
sidebar_current: "docs-cloudflare-resource-zone"
description: |-
  Provides a Cloudflare resource to create and modify a zone.
---

# cloudflare_zone

Provides a Cloudflare Zone resource. Zone is the basic resource for working with Cloudflare and is roughly equivalent to a domain name that the user purchases.

## Example Usage

```hcl
resource "cloudflare_zone" "example" {
    zone = "example.com"
}
```

## Argument Reference

The following arguments are supported:

* `zone` - (Required) The DNS zone name which will be added.
* `paused` - (Optional) Boolean of whether this zone is paused (traffic bypasses Cloudflare). Default: false.
* `jump_start` - (Optional) Boolean of whether to scan for DNS records on creation. Ignored after zone is created. Default: false.
* `plan` - (Optional) The name of the commercial plan to apply to the zone, can be updated once the zone is created; one of `free`, `pro`, `business`, `enterprise`.
* `type` - A full zone implies that DNS is hosted with Cloudflare. A partial zone is typically a partner-hosted zone or a CNAME setup. Valid values: `full`, `partial`. Default is `full`.

## Attributes Reference

The following attributes are exported:

* `id` - The zone ID.
* `plan` - The name of the commercial plan to apply to the zone.
* `vanity_name_servers` - List of Vanity Nameservers (if set).
* `meta.wildcard_proxiable` - Indicates whether wildcard DNS records can receive Cloudflare security and performance features.
* `meta.phishing_detected` - Indicates if URLs on the zone have been identified as hosting phishing content.
* `status` - Status of the zone. Valid values: `active`, `pending`, `initializing`, `moved`, `deleted`, `deactivated`.
* `name_servers` - Cloudflare-assigned name servers. This is only populated for zones that use Cloudflare DNS.
* `verification_key` - Contains the TXT record value to validate domain ownership. This is only populated for zones of type `partial`. 

## Import

Zone resource can be imported using a zone ID, e.g.

```
$ terraform import cloudflare_zone.example d41d8cd98f00b204e9800998ecf8427e
```

where:

* `d41d8cd98f00b204e9800998ecf8427e` - zone ID, as returned from [API](https://api.cloudflare.com/#zone-list-zones)
