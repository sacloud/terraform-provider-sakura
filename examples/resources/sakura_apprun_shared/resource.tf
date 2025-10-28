resource "sakura_apprun_shared" "foobar" {
  name            = "foobar"
  timeout_seconds = 60
  port            = 8080
  min_scale       = 0
  max_scale       = 1
  components =[{
    name       = "foobar"
    max_cpu    = "0.1"
    max_memory = "256Mi"
    deploy_source = {
      container_registry = {
        image    = "foobar.sakuracr.jp/my-app:latest"
        //server   = "foobar.sakuracr.jp"
        //username = "username"
        //password = "userpassword"
      }
    }
    env = [{
      key   = "key"
      value = "value"
    }]
    probe = {
      http_get = {
        path = "/"
        port = 8080
        headers = [{
          name  = "name"
          value = "value"
        },
        {
          name  = "name2"
          value = "value2"
        }]
      }
    }
  }]
  traffics = [{
    version_index = 0
    percent       = 100
  }]
  packet_filter = {
    enabled = true
    settings = [{
      from_ip               = "192.0.2.0"
      from_ip_prefix_length = "24"
    }]
  }
}
