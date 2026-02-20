resource "sakura_iam_organization_id_policy" "foobar" {
  bindings = [{
    principals = [{
      id   = "service-principal-id"
      type = "service-principal"
    }],
    role = {
      id   = "identity-viewer"
      type = "preset"
    }
  },
  {
    principals = [{
      id   = "service-principal-id"
      type = "service-principal"
    }],
    role = {
      id   = "identity-admin"
      type = "preset"
    }
  }]
}
