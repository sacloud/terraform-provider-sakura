---
page_title: "Guide: Use Object Storage for Terraform Backend"
subcategory: "Guides"
description: |-
  Use Sakura's Object Storage for Terraform's S3 Backend
---

# Use Sakura's Object Storage for Terraform's S3 Backend

Terraform can use the S3 backend to store state files in AWS S3. Sakura's Object Storage provides the S3-compatible API, so you can store state files in Sakura's Object Storage by using the S3 backend.

https://developer.hashicorp.com/terraform/language/backend/s3

Below is an example configuration.

- backend.tf

```tf
terraform {
  backend "s3" {
    // Following parameters are depend on your environment
    bucket                   = "bucket-name"
    key                      = "terraform.tfstate"
    region                   = "jp-east-1" // "jp-north-1" for Ishikari
    shared_credentials_files = ["./s3.credentials"] // or access_key/secret_key parameters or ENV vars
    endpoints = {
      s3 = "https://s3.tky01.sakurastorage.jp" // "https://s3.isk01.sakurastorage.jp" for Ishikari
    }

    // Don't edit following parameters
    skip_credentials_validation = true
    skip_metadata_api_check     = true
    skip_region_validation      = true
    skip_requesting_account_id  = true
    skip_s3_checksum            = true
  }
}
```

You can check the values for region/endpoints on the Control Panel site page or via Sakura's Object Storage API.

- s3.credentials

```
[default]
aws_access_key_id = "object-storage-access-key"
aws_secret_access_key = "object-storage-secret-key"
```

Obtain these values by creating a site and permissions.

- Additional Environment Variables

You need to set following environment variables to avoid API error.

```
AWS_REQUEST_CHECKSUM_CALCULATION=WHEN_REQUIRED
AWS_RESPONSE_CHECKSUM_VALIDATION=WHEN_REQUIRED
```

See also: https://cloud.sakura.ad.jp/news/2025/02/04/objectstorage_defectversion/

## Notes

Due to current API limitations, the `use_lockfile` parameter cannot be used when using Sakura's Object Storage.
