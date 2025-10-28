resource "sakura_ssh_key" "foobar" {
  name       = "foobar"
  public_key = file("~/.ssh/id_rsa.pub")
}