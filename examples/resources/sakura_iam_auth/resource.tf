resource "sakura_iam_auth" "foobar" {
  password_policy = {
    min_length = 12
    require_uppercase = true
    require_lowercase = true
    require_symbols = false
  }
  conditions = {
    ip_restriction = {
      mode = "allow_list"
      source_network = ["192.168.10.1"]
    }
    require_two_factor_auth = false
    datetime_restriction = {
      after = "2026-01-01T00:00:00+09:00",
      before = "2027-02-01T00:00:00+09:00",
    }
  }
}
