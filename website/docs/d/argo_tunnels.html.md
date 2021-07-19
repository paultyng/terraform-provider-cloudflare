---
layout: "cloudflare"
page_title: "Cloudflare: cloudflare_argo_tunnels"
sidebar_current: "docs-cloudflare-datasource-argo-tunnels"
description: |-
  Get information on a Cloudflare Argo Tunnels.
---

# cloudflare_argo_tunnels

Use this data source to look up [Tunnel][1] records.

## Example usage

Given you have the following Argo tunnels in Cloudflare.

- tunnel-example
- origin2-tunnel
- some-other-tunnel

```hcl
# Look for a single tunnel that you know exists using an exact match.
# Provider will perform client side filtering using the provided name and will only match the single tunnel,
# tunnel-example.
data "cloudflare_argo_tunnels" "example" {
  filter {
    name = "tunnel-example"
  }
}
```

```hcl
# Look for all tunnels which start with "origin2-".
# Provider will perform client side filtering using the provided regex and will only match the single zone,
# origin2-tunnel.
data "cloudflare_argo_tunnels" "example" {
  filter {
    match       = "^origin2-"
  }
}
```

### Example usage with other resources

The example below matches a single tunnel with the name "tunnel-example".
The matched tunnel is then referenced in the dns record.

```hcl
data "cloudflare_argo_tunnels" "test" {
  filter {
    name       = "tunnel-example"
  }
}

resource "cloudflare_record" "example" {
  zone_id     = var.zone_id
  name        = "tunnel-example"
  value       = "${lookup(data.cloudflare_argo_tunnels.test.tunnels[0], "id")}.cfargotunnel.com"
  type        = "CNAME"
  proxied     = true
}
```

## Argument Reference
- `filter` - (Optional) One or more values used to look up tunnels. If more than one value is given all
values must match in order to be included, see below for full list.

**filter**

- `name` - (Optional) A string value to search for.
- `match` - (Optional) A RE2 compatible regular expression to filter the
  results. This is performed client side.
- `deleted` - (Optional) Deleted status of the tunnels to search for. Valid values are
  `true` or `false`. Defaults to `false`.

## Attributes Reference

- `tunnels` - A list of tunnels objects. Object format:

**tunnels**

- `id` - The tunnel ID
- `name` - Tunnel name
- `deleted` - Whether or not tunnel is deleted.

[1]: https://api.cloudflare.com/#argo-tunnel-list-argo-tunnels
