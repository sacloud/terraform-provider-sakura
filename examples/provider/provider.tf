terraform {
  required_providers {
    sakura = {
      # terraform-provider-sakura supports Terraform 1.11 and later
      source = "sacloud/sakura"

      # We recommend pinning to the specific version of the Sakura Provider you're using
      # since new versions are released frequently
      version = "3.12.1"
      #version = "~> 3"
    }
  }
}

# Configure the Sakura Provider
provider "sakura" {
  # More information on the authentication methods supported by
  # the Sakura Provider can be found here:
  # https://docs.usacloud.jp/terraform/provider/

  # profile = "..."
}
