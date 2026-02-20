resource "sakura_iam_project" "foobar" {
  name = "foobar"
  code = "foobar-code"
  description = "description"
  // for project under folder
  //parent_folder_id = sakura_iam_folder.foobar.id
}