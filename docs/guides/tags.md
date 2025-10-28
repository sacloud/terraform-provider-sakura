---
page_title: "Guide: Handling of Resource Tags with SAKURA Provider"
subcategory: "Guides"
description: |-
  Handling of Resource Tags with SAKURA Provider.
---

# Handling of Resource Tags with SAKURA Provider

You can manage resource tags with SAKURA provider as follows:

```tf
resource sakura_server "example" {
  name = "example"
  tags = ["tag1", "tag2"]
}
```

However, some tags cannot be managed from Terraform.

- Database Appliance
  - `@MariaDB-*`
  - `@postgres-*`
- Server / Enhanced Load Balancer / Router  
  - `@previous-id*`
  
### `@MariaDB-*` and `@postgres-*`

Tags such as `@MariaDB-*` and `@postgres-*` are automatically assigned when Database Appliance is created, but this provider will ignores them.
Also, it is not supported to specify tags with these prefixes in the tf file.

### `@previous-id`

`@previous-id` tag is assigned when resource plan is changed with using libsacloud, but this provider will ignore it when storing tags to state.
This tag is used as a search condition for fallback when a 404 error occurs when fetching by resource ID.
Also, it is not supported to specify tags with these prefixes in the tf file.
