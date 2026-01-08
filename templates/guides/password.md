---
page_title: "Guide: Write-only password"
subcategory: "Guides"
description: |-
  Use write-only password for newer deployments
---

# Write-only password

Since v3.1.0, the Sakura provider supports write-only password fields on resources. We recommend using write-only password fields instead of the existing password field. See the Terraform documentation for write-only arguments: https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments

Below is an example showing write-only password fields:

```tf
resource "sakura_database" "foobar" {
  database_type = "mariadb"

  username = "user1"
  password_wo = "password1"
  replica_password_wo = "replicapass1"
  password_wo_version = 1

  // for backward compatibility
  //password = "password1"
  //replica_password_wo = "replicapass1"

  // ...
}
```

Existing resources expose `password_wo` and `password_wo_version` for write-only passwords. Terraform cannot show a diff for write-only fields, so update `password_wo_version` to trigger a password change. Note that `password_wo_version` must be greater than zero.

Newer resources (since v3), such as eventbus/nosql/etc, use write-only passwords by default and don't provide a non-write-only password field.