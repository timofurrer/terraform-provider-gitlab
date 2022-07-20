resource "gitlab_project" "example" {
  name        = "example"
  description = "My awesome codebase"

  visibility_level = "public"
}

# Project with custom push rules
resource "gitlab_project" "example-two" {
  name = "example-two"

  push_rules {
    author_email_regex     = "@example\\.com$"
    commit_committer_check = true
    member_check           = true
    prevent_secrets        = true
  }
}

# Create a project for a given user (requires admin access)
data "gitlab_user" "peter_parker" {
  username = "peter_parker"
}

resource "gitlab_project" "peters_repo" {
  name         = "peters-repo"
  description  = "This is a description"
  namespace_id = data.gitlab_user.peter_parker.namespace_id
}

resource "gitlab_project" "protected" {
  name                   = "protected-example"
  description            = "This repository has some branches protected"
  default_branch         = "main"
  initialize_with_readme = true

  // Protected branches
  protected_branch {
    branch             = "main"
    merge_access_level = "developer"
    push_access_level  = "developer"
  }

  protected_branch {
    branch             = "release"
    merge_access_level = "maintainer"
    push_access_level  = "maintainer"
  }
}
