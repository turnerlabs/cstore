//name of the bucket
variable "bucket_name" {
}

//enable versioning
variable "versioning" {
  default = false
}

//bucket and kms key access: user role
variable "saml_role" {
}

//bucket and kms key access: list of federated assumed role users (e.g., aws-account-devops/me@turner.com). Roles must exist in the target account and are case sensitive.
variable "saml_users" {
  type = list(string)
}

variable "iam_users" {
  type = list(string)
}

// bucket and kms key access: list of roles that need access to the bucket
variable "app_roles" {
  type = list(string)
}

# Tags for the infrastructure
variable "tags" {
  type = map(string)
}

//incomplete multipart upload deletion
variable "multipart_delete" {
  default = true
}

variable "multipart_days" {
  default = 3
}

