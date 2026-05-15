resource "sakura_webaccel_acl" "foobar" {
  site_id = "webaccel-site-id" # e.g. sakura_webaccel.foobar.id
  acl = join("\n", [
    "deny 192.0.2.5/25",
    "deny 198.51.100.0",
    "allow all",
  ])
}