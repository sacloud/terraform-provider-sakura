variable "users" {
  type = list(object({
    name             = string
    password         = string
    password_version = optional(number, 1)
    permission       = string
  }))
  default = [
    {
      name             = "user1"
      password         = "password1"
      password_version = 1
      permission       = "all"
    },
    {
      name             = "user2"
      password         = "password2"
      password_version = 1
      permission       = "readwrite"
    }
  ]
}

resource "sakura_container_registry" "foobar" {
  name            = "foobar"
  subdomain_label = "your-subdomain-label"
  description     = "description"
  tags            = ["tag1", "tag2"]

  user = [
    for user in var.users : {
      name                = user.name
      password_wo         = user.password
      password_wo_version = user.password_version
      permission          = user.permission
    }
  ]
}
