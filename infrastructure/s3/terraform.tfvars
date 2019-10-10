bucket_name = "app-configs"

saml_role = "devops"

# read and write permission (case sensitive)
saml_users = [
  "User.Email@Org.com",
]

# read and write permission
iam_users = [
  "srv_app_user_dev"
]

# read only permission
app_roles = [
  "app-dev",
]

tags = {
  application      = "app"
  environment      = "prod"
  team             = "team-name"
  customer         = "team-name"
  contact-email    = "team@email.com"
  product          = "app"
  project          = "app"
}