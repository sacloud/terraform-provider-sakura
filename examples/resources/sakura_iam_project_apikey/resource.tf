resource "sakura_iam_project_apikey" "foobar" {
  name = "foobar"
  description = "description"
  project_id = "project-id" // This project must hav permissions to create API keys
  iam_roles = ["resource-creator"]
}