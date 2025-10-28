resource "sakura_script" "foobar" {
  name    = "foobar"
  content = file("startup-script.sh")
}