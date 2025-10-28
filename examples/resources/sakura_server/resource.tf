resource "sakura_server" "foobar" {
  name        = "foobar"
  disks       = [sakura_disk.foobar.id]
  description = "description"
  tags        = ["tag1", "tag2"]

  network_interface = [{
    upstream         = "shared"
    packet_filter_id = data.sakura_packet_filter.foobar.id
  }]

  disk_edit_parameter = {
    hostname = "hostname"
    password = "password"
    disable_pw_auth = true

    # ssh_keys    = ["ssh-rsa xxxxx"]
    # ssh_key_ids = ["<ID>", "<ID>"]
    # script = [{
    #  id         = "<ID>"
    #  api_key_id = "<ID>"
    #  variables = {
    #    foo = "bar"
    #  }
    # }]
  }

  # If you use cloud-init instead of disk_edit_parameter

  # user_data = join("\n", [
  #   "#cloud-config",
  #   yamlencode({
  #     hostname: "hostname",
  #     password: "password",
  #     chpasswd: {
  #       expire: false,
  #     }
  #     ssh_pwauth: false,
  #     ssh_authorized_keys: [
  #       file("~/.ssh/id_rsa.pub"),
  #     ],
  #   }),
  # ])
}

data "sakura_packet_filter" "foobar" {
  name = "foobar"
}

data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}

resource "sakura_disk" "foobar" {
  name              = "foobar"
  source_archive_id = data.sakura_archive.ubuntu.id
}