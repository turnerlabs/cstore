## Secrets Manager Store ##

When storing configuration in Secrets Manager, two storate solutions are available.

| CLI Key | Description | Supports | Secret Key |
|-|-|-|-|
|`aws-secret`| All config values are stored in a single secret. | `.env`, `.json`|`/{config_context}/{FILE_PATH}` |
|`aws-secrets`| Each config value is stored in a separate secret. | `.env`, `.json` | `/{config_context}/{file_path}/{var}` |

### Authentication ###

To authenticate with AWS, use one of the [AWS methods](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html).

### Updating Configuration ###

Secrets Manager is updated only when the value or encryption of the configuration has changed.

The user is warned when a secret has changed in Secrets Manager since the last time the configuration was retrieved.

Deleted secrets will be put in Secrets Manager's "Pending Deletion" state for thirty days days.

### Encryption ###

The initial configuration push to Secrets Manager prompts the user for a KMS key. To change the key, edit the KMS key in the catalog file and re-push.

### Version Configuration ###

Versioning is currently not supported.

### AWS Access Policy ###

Set the `config_context` avariable to the cstore context and apply this policy to any resource role to allow AWS Secrets Manager access.

```yml
data "aws_iam_policy_document" "app_policy" {
   statement {
    effect = "Allow"

    actions = [
      "secretsmanager:GetSecretValue",
    ]

    resources = [
      "arn:aws:secretsmanager:us-east-1:${var.account_id}:secret:${var.config_context}/*",
    ]
  }
}

variable account_id {}

variable config_context {}
```
