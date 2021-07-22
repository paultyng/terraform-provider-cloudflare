---
layout: "cloudflare"
page_title: "Cloudflare: cloudflare_access_group"
sidebar_current: "docs-cloudflare-resource-access-group"
description: |-
  Provides a Cloudflare Access Group resource.
---

# cloudflare_access_group

Provides a Cloudflare Access Group resource. Access Groups are used
in conjunction with Access Policies to restrict access to a
particular resource based on group membership.

## Example Usage

```hcl
# Allowing access to `test@example.com` email address only
resource "cloudflare_access_group" "test_group" {
  account_id     = "975ecf5a45e3bcb680dba0722a420ad9"
  name           = "staging group"

  include {
    email = ["test@example.com"]
  }
}

# Allowing `test@example.com` to access but only when coming from a
# specific IP.
resource "cloudflare_access_group" "test_group" {
  account_id     = "975ecf5a45e3bcb680dba0722a420ad9"
  name           = "staging group"

  include {
    email = ["test@example.com"]
  }

  require = {
    ip = [var.office_ip]
  }
}
```

## Argument Reference

The following arguments are supported:

-> **Note:** It's required that an `account_id` or `zone_id` is provided and in most cases using either is fine. However, if you're using a scoped access token, you must provide the argument that matches the token's scope. For example, an access token that is scoped to the "example.com" zone needs to use the `zone_id` argument.

* `account_id` - (Optional) The ID of the account the group is associated with. Conflicts with `zone_id`.
* `zone_id` - (Optional) The ID of the zone the group is associated with. Conflicts with `account_id`.
* `name` - (Required) Friendly name of the Access Group.
* `require` - (Optional) A series of access conditions, see below for
  full list.
* `exclude` - (Optional) A series of access conditions, see below for
  full list.
* `include` - (Required) A series of access conditions, see below for
  full list.

## Conditions

`require`, `exclude` and `include` arguments share the available
conditions which can be applied. The conditions are:

* `ip` - (Optional) A list of IP addresses or ranges. Example:
  `ip = ["1.2.3.4", "10.0.0.0/2"]`
* `email` - (Optional) A list of email addresses. Example:
  `email = ["test@example.com"]`
* `email_domain` - (Optional) A list of email domains. Example:
  `email_domain = ["example.com"]`
* `service_token` - (Optional) A list of service token ids. Example:
  `service_token = [cloudflare_access_service_token.demo.id]`
* `any_valid_service_token` - (Optional) Boolean indicating if allow
  all tokens to be granted. Example: `any_valid_service_token = true`
* `group` - (Optional) A list of access group ids. Example:
  `group = [cloudflare_access_group.demo.id]`
* `everyone` - (Optional) Boolean indicating permitting access for all
  requests. Example: `everyone = true`
* `certificate` - (Optional) Whether to use mTLS certificate authentication.
* `common_name` - (Optional) Use a certificate common name to authenticate with.
* `auth_method` - (Optional) A string identifying the authentication
  method code. The list of codes are listed here: https://tools.ietf.org/html/rfc8176#section-2.
  Custom values are also supported. Example: `auth_method = ["swk"]`
* `geo` - (Optional) A list of country codes. Example: `geo = ["US"]`
* `login_method` - (Optional) A list of identity provider ids. Example: `login_method = [cloudflare_access_identity_provider.my_idp.id]`
* `device_posture` - (Optional) A list of device_posture integration_uids. Example: `device_posture = [cloudflare_device_posture_rule.my_posture_rule.id]`
* `gsuite` - (Optional) Use GSuite as the authentication mechanism. Example:

  ```hcl
  # ... other configuration
  include {
    gsuite {
      email = ["admins@example.com"]
      identity_provider_id = "ca298b82-93b5-41bf-bc2d-10493f09b761"
    }
  }
  ```
* `github` - (Optional) Use a GitHub organization as the `include` condition. Example:

  ```hcl
  # ... other configuration
  include {
    github {
      name = "my-github-org-name" # (Required) GitHub organization name
      teams = ["my-github-team-name"] # (Optional) GitHub teams
      identity_provider_id = "ca298b82-93b5-41bf-bc2d-10493f09b761"
    }
  }
  ```
* `azure` - (Optional) Use Azure AD as the `include` condition. Example:

  ```hcl
  # ... other configuration
  include {
    azure {
      id = ["86773093-5feb-48dd-814b-7ccd3676ff50e"]
      identity_provider_id = "ca298b82-93b5-41bf-bc2d-10493f09b761"
    }
  }
  ```
* `okta` - (Optional) Use Okta as the `include` condition. Example:

  ```hcl
  # ... other configuration
  include {
    okta {
      name = ["admins"]
      identity_provider_id = "ca298b82-93b5-41bf-bc2d-10493f09b761"
    }
  }
  ```
* `saml` - (Optional) Use an external SAML setup as the `include` condition.
  Example:

  ```hcl
  # ... other configuration
  include {
    saml {
      attribute_name = "group"
      attribute_value = "admins"
      identity_provider_id = "ca298b82-93b5-41bf-bc2d-10493f09b761"
    }
  }
  ```

## Import

Access Groups can be imported using a composite ID formed of account
ID and group ID.

```
$ terraform import cloudflare_access_group.staging 975ecf5a45e3bcb680dba0722a420ad9/67ea780ce4982c1cfbe6b7293afc765d
```

where

* `975ecf5a45e3bcb680dba0722a420ad9` - Account ID
* `67ea780ce4982c1cfbe6b7293afc765d` - Access Group ID
