---
layout: "cloudflare"
page_title: "Cloudflare: cloudflare_access_service_token"
sidebar_current: "docs-cloudflare-resource-access-service-token"
description: |-
  Provides a Cloudflare Access Service Token resource.
---

# cloudflare_access_service_token

Access Service Tokens are used for service-to-service communication
when an application is behind Cloudflare Access.

## Example Usage

```hcl
resource "cloudflare_access_service_token" "my_app" {
  account_id = "d41d8cd98f00b204e9800998ecf8427e"
  name       = "CI/CD app"
}

# Generate a service token that will renew if terraform is ran within 30 days of expiration
resource "cloudflare_access_service_token" "my_app" {
  account_id = "d41d8cd98f00b204e9800998ecf8427e"
  name       = "CI/CD app renewed"

  min_days_for_renewal = 30

  # This flag is important to set if min_days_for_renewal is defined otherwise 
  # there will be a brief period where the service relying on that token 
  # will not have access due to the resource being deleted
  lifecycle {
    create_before_destroy = true
  }
}
```

## Argument Reference

The following arguments are supported:

-> **Note:** It's required that an `account_id` or `zone_id` is provided and in most cases using either is fine. However, if you're using a scoped access token, you must provide the argument that matches the token's scope. For example, an access token that is scoped to the "example.com" zone needs to use the `zone_id` argument.

* `account_id` - (Optional) The ID of the account where the Access Service is being created. Conflicts with `zone_id`.
* `zone_id` - (Optional) The ID of the zone where the Access Service is being created. Conflicts with `account_id`.
* `name` - (Required) Friendly name of the token's intent.
* `min_days_for_renewal` - (Optional) Regenerates the token if terraform is run within the specified amount of days before expiration

## Attributes Reference

The following attributes are exported:

* `client_id` - UUID client ID associated with the Service Token.
* `client_secret` - A secret for interacting with Access protocols.
* `expires_at` - Date when the token expires

## Import

~> **Important:** If you are importing an Access Service Token you will
not have the `client_secret` available in the state for use. The
`client_secret` is only available once, at creation. In most cases, it
is better to just create a new resource should you need to reference it
in other resources.

Access Service Tokens can be imported using a composite ID formed of
account ID and Service Token ID.

```
$ terraform import cloudflare_access_service_token.my_app cb029e245cfdd66dc8d2e570d5dd3322/d41d8cd98f00b204e9800998ecf8427e
```

where

* `cb029e245cfdd66dc8d2e570d5dd3322` - Account ID
* `d41d8cd98f00b204e9800998ecf8427e` - Access Service Token ID
