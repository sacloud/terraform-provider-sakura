resource "sakura_iam_user_provisioning" "foobar" {
  name = "foobar"
  # For secret token regeneration, set higher token version.
  # token_version = 2
}