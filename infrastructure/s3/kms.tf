locals {
  # KMS write actions
  kms_write_actions = [
    "kms:CancelKeyDeletion",
    "kms:CreateAlias",
    "kms:CreateGrant",
    "kms:CreateKey",
    "kms:DeleteAlias",
    "kms:DeleteImportedKeyMaterial",
    "kms:DisableKey",
    "kms:DisableKeyRotation",
    "kms:EnableKey",
    "kms:EnableKeyRotation",
    "kms:Encrypt",
    "kms:GenerateDataKey",
    "kms:GenerateDataKeyWithoutPlaintext",
    "kms:GenerateRandom",
    "kms:GetKeyPolicy",
    "kms:GetKeyRotationStatus",
    "kms:GetParametersForImport",
    "kms:ImportKeyMaterial",
    "kms:PutKeyPolicy",
    "kms:ReEncryptFrom",
    "kms:ReEncryptTo",
    "kms:RetireGrant",
    "kms:RevokeGrant",
    "kms:ScheduleKeyDeletion",
    "kms:TagResource",
    "kms:UntagResource",
    "kms:UpdateAlias",
    "kms:UpdateKeyDescription",
  ]

  # KMS read actions
  kms_read_actions = [
    "kms:Decrypt",
    "kms:DescribeKey",
    "kms:List*",
  ]

  # list of saml users for policies
  write_user_ids = flatten([
    data.aws_caller_identity.current.user_id,
    data.aws_caller_identity.current.account_id,
    formatlist(
        "%s:%s",
        data.aws_iam_role.saml_role_config.unique_id,
        var.saml_users,
    ),
    values({
        for user in data.aws_iam_user.iam_user_config:
        user.user_id => user.user_id
    }),
  ])

  read_user_ids = flatten([
    local.write_user_ids,
    values({
        for role in data.aws_iam_role.app_role_config:
        role.unique_id => format("%s:*", role.unique_id)
    })
  ])
}

# get the saml user info so we can get the unique_id
data "aws_iam_role" "saml_role_config" {
  name = var.saml_role
}

data "aws_iam_role" "app_role_config" {
  count = length(var.app_roles)

  name          = var.app_roles[count.index]
}

data "aws_iam_user" "iam_user_config" {
  count = length(var.iam_users)

  user_name          = var.iam_users[count.index]
}

# kms key used to encrypt configuration
resource "aws_kms_key" "config" {
  description             = "configuration key for ${var.bucket_name}"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  tags                    = var.tags
  policy                  = data.aws_iam_policy_document.config.json
}

resource "aws_kms_alias" "config" {
  name          = "alias/${var.bucket_name}"
  target_key_id = "${aws_kms_key.config.id}"
}

data "aws_iam_policy_document" "config" {
  statement {
    sid    = "DenyWriteToAllExceptSAMLUsers"
    effect = "Deny"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions   = local.kms_write_actions
    resources = ["*"]

    condition {
      test     = "StringNotLike"
      variable = "aws:userId"
      values   = local.write_user_ids
    }
  }

  statement {
    sid    = "DenyReadToAllExceptRoleAndSAMLUsers"
    effect = "Deny"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions   = local.kms_read_actions
    resources = ["*"]

    condition {
      test     = "StringNotLike"
      variable = "aws:userId"
      values   = local.read_user_ids
    }
  }

  statement {
    sid    = "AllowWriteToSAMLUsers"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions   = local.kms_write_actions
    resources = ["*"]

    condition {
      test     = "StringLike"
      variable = "aws:userId"
      values   = local.write_user_ids
    }
  }

  statement {
    sid    = "AllowReadRoleAndSAMLUsers"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions   = local.kms_read_actions
    resources = ["*"]

    condition {
      test     = "StringLike"
      variable = "aws:userId"
      values   = local.read_user_ids
    }
  }
}

output "config_key_id" {
  value = aws_kms_key.config.key_id
}