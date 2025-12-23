# foo.tf
import {
  to = sakura_dns_record.record001
  identity = {
    dns_id = "dns-resource-id"
    name   = "www"
    type   = "A"
    // if name/type are sufficient to identify, value can be omitted
    //value  = "192.168.0.1"
  }
}

import {
  to = sakura_dns_record.record002
  identity = {
    dns_id = "dns-resource-id"
    name   = "www2"
    type   = "A"
    value  = "192.168.0.2"
    // other parameters are optional
  }
}

# Run commands
# $ terraform plan -generate-config-out=generated_conf.tf
# $ terraform apply
