data "sakura_apigw_service" "foobar" {
  name = "foobar"
}

data "sakura_apigw_group" "foobar" {
  name = "foobar"
}

resource "sakura_apigw_route" "foobar" {
  name       = "foobar"
  tags       = ["tag1"]
  service_id = data.sakura_apigw_service.foobar.id
  protocols  = "http,https"
  methods    = ["GET", "POST", "PUT", "HEAD"]
  ip_restriction = {
    protocols = "http,https"
    restricted_by = "allowIps"
    ips = ["192.168.0.1"]
  }
  groups = [{
    id = data.sakura_apigw_group.foobar.id,
    enabled = true
  }]
  request_transformation = {
    http_method = "GET",
    allow = {
      body = ["foo"],
    },
    rename = {
      headers = [{
        from = "X-Old-Header",
        to   = "X-New-Header"
      }]
      query_params = [{
        from = "old_param",
        to   = "new_param"
      }]
      body = [{
        from = "old_body",
        to   = "new_body"
      }]
    },
    // other conditions...
  }
  response_transformation = {
    allow = {
      json_keys = ["foo"],
    },
    remove = {
      if_status_code = [426]
      header_keys = ["X-Remove-Header"],
      json_keys = ["remove_param"],
    },
    add = {
      if_status_code = [426]
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      }]
      json = [{
        key   = "old_param",
        value = "new_param"
      }]
    },
    // other conditions...
  }
}
