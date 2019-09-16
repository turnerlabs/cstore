resource "aws_kms_key" "bucket_key" {
  policy = "${data.template_file.key_policy.rendered}"
  enable_key_rotation  = "true"

  tags {
    team          = "${var.tag_team}"
    application   = "${var.tag_application}"
    environment   = "${var.tag_environment}"
    contact-email = "${var.tag_contact-email}"
    customer      = "${var.tag_customer}"
  }
}

resource "aws_kms_alias" "bucket_key_alias" {
  name          = "alias/${var.bucket_name}-key"
  target_key_id = "${aws_kms_key.bucket_key.key_id}"
}

//render dynamic list of role for kms
data "template_file" "kms_role_principal" {
  count    = "${length(var.roles)}"
  template = "arn:aws:iam::$${account}:role/$${role}"

  vars {
    account = "${data.aws_caller_identity.current.account_id}"
    role    = "${var.roles[count.index]}"
  }
}

//render KMS key policy including dynamic principals
data "template_file" "key_policy" {
  template = <<EOF
{
  "Version": "2012-10-17",
  "Id": "key-default-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::{{ACCOUNT_ID}}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    },
    {
      "Sid": "Allow roles to encrypt and decrypt",
      "Effect": "Allow",
      "Principal": { "AWS": $${principals} },
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt"
      ],
      "Resource": "*" 
    }
  ]
}
EOF

  vars {
    principals = "${jsonencode(data.template_file.kms_role_principal.*.rendered)}"
  }
}
