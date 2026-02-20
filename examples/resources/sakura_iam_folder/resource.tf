resource "sakura_iam_folder" "foobar" {
  name = "foobar"
  description = "description"
  // for nested folder
  //parent_id = sakura_iam_folder.parent.id
}
