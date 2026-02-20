resource "sakura_iam_project" "foobar" {
  name = "foobar"
  code = "foobar-code"
  description = "description"
}

resource "sakura_iam_service_principal" "foobar" {
  name = "foobar"
  description = "description"
  project_id  = sakura_iam_project.foobar.id
}