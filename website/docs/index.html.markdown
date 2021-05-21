---
layout: "cloudflare"
page_title: "Provider: Cloudflare"
sidebar_current: "docs-cloudflare-index"
description: |-
  The Cloudflare provider is used to interact with resources supported by Cloudflare. The provider needs to be configured with the proper credentials before it can be used.
---

# Cloudflare Provider

The Cloudflare provider is used to interact with resources supported by
Cloudflare. The provider needs to be configured with the proper credentials
before it can be used.

Use the navigation to the left to read about the available resources.

## Getting Started

Try the [Host a Static Website with S3 and Cloudflare](https://learn.hashicorp.com/tutorials/terraform/cloudflare-static-website) tutorial on HashiCorp Learn. In this tutorial, you will set up a static website using AWS S3 as an object store and Cloudflare for DNS, SSL and CDN, then create Cloudflare page rules to always redirect HTTPS and temporarily redirect certain paths.

## Example Usage

### Terraform >= 0.13

```hcl
# Configure the Cloudflare provider using the required_providers stanza required with Terraform 0.13 and beyond
# You may optionally use version directive to prevent breaking changes occurring unannounced.
terraform {
  required_providers {
    cloudflare = {
      source = "cloudflare/cloudflare"
      version = "~> 2.0"
    }
  }
}

provider "cloudflare" { 
  email   = var.cloudflare_email
  api_key = var.cloudflare_api_key
}

# Create a record
resource "cloudflare_record" "www" {
  # ...
}

# Create a page rule
resource "cloudflare_page_rule" "www" {
  # ...
}
```

### Terraform < 0.13

```hcl
# Configure the Cloudflare provider.
# You may optionally use version directive to prevent breaking changes occurring unannounced.
provider "cloudflare" {
  version = "~> 2.0"
  email   = "${var.cloudflare_email}"
  api_key = "${var.cloudflare_api_key}"
}

# Create a record
resource "cloudflare_record" "www" {
  # ...
}

# Create a page rule
resource "cloudflare_page_rule" "www" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

* `email` - (Optional) The email associated with the account. This can also be
  specified with the `CLOUDFLARE_EMAIL` shell environment variable.
* `api_key` - (Optional) The Cloudflare API key. This can also be specified
  with the `CLOUDFLARE_API_KEY` shell environment variable.
* `api_token` - (Optional) The Cloudflare API Token. This can also be specified
  with the `CLOUDFLARE_API_TOKEN` shell environment variable. This is an
  alternative to `email`+`api_key`. If both are specified, `api_token` will be
  used over `email`+`api_key` fields.
* `api_user_service_key` - (Optional) The Cloudflare API User Service Key. This can also be specified 
  with the `CLOUDFLARE_API_USER_SERVICE_KEY` shell environment variable. The value is 
  to be used in combination with an `api_token`, or `email` and `api_key`.
  This is used for a specific set of endpoints, such as creating Origin CA certificates.
* `rps` - (Optional) RPS limit to apply when making calls to the API. Default: 4.
  This can also be specified with the `CLOUDFLARE_RPS` shell environment variable.
* `retries` - (Optional) Maximum number of retries to perform when an API request fails. Default: 3.
  This can also be specified with the `CLOUDFLARE_RETRIES` shell environment variable.
* `min_backoff` - (Optional) Minimum backoff period in seconds after failed API calls. Default: 1.
  This can also be specified with the `CLOUDFLARE_MIN_BACKOFF` shell environment variable.
* `max_backoff` - (Optional) Maximum backoff period in seconds after failed API calls Default: 30.
  This can also be specified with the `CLOUDFLARE_MAX_BACKOFF` shell environment variable.
* `api_client_logging` - (Optional) Whether to print logs from the API client (using the default log library logger). Default: false.
  This can also be specified with the `CLOUDFLARE_API_CLIENT_LOGGING` shell environment variable.
* `account_id` - (Optional) Configure API client with this account ID, so calls use the account API rather than the (default) user API.
  This is required for other users in your account to have access to the resources you manage.
  This can also be specified with the `CLOUDFLARE_ACCOUNT_ID` shell environment variable.
