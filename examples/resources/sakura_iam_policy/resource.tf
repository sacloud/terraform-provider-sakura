resource "sakura_iam_project" "foobar" {
  name = "foobar"
  code = "foobar-code"
  description = "description"
}

resource "sakura_iam_service_principal" "foobar" {
  name = "foobar"
  description = "description"
  project_id = sakura_iam_project.foobar.id
}

resource "sakura_iam_policy" "foobar" {
  target = "project" // "folder" or "organization" is also available
  target_id = sakura_iam_project.foobar.id
  bindings = [{
    principals = [{
      id   = sakura_iam_service_principal.foobar.id
      type = "service-principal"
    }],
    role = {
      id   = "owner"
      type = "preset"
    }
  },
  {
    principals = [{
      id   = sakura_iam_service_principal.foobar.id
      type = "service-principal"
    }],
    role = {
      id   = "organization-admin"
      type = "preset"
    }
  }]
}