## Using Secrets Manager Store ##

Secrets Manager requires creating a KMS key to use for encryption.

cStore will create a secret in AWS Secrets Manager for each variable in the configuration file.

To authenticate with AWS, use one of the [AWS methods](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html).

### Parameter Key Formatting ###

Each secrets key will be generated using the following format. 
- `/{CSTORE_CONTEXT}/{FILE_PATH}/{VAR}` (default)

### Versioning Configuration ###

Versioning is currently not supported.

### Pushing Configuration Changes ###

When pushing changes, Secrets Manager will only be updated when the value or encryption of the parameter has changed.

If a secret has changed in Secrets Manager since the last time the configuration was pulled by cStore, cStore will warn before overwriting the changes.

When a parameter is removed from the configuration file, and the file is pushed, the secret will set to pending deletion after 30 days. If the parameter is added back to the configuration file, and pushed within thirty days, cStore will request restoration and be able to modify the secret again.

### Encryption ###

With the initial configuration push to Secrets Manager, encryption settings are saved. To change these settings, edit the KMS key in the catalog file and re-push.

### AWS Access Policy ###

Update any resource policy that needs access to Secrets Manager.

```yml
data "aws_iam_policy_document" "app_policy" {
   statement {
    effect = "Allow"

    actions = [
      "secretsmanager:GetSecretValue",
    ]

    resources = [
      "arn:aws:secretsmanager:us-east-1:${var.account_id}:secret:${var.secrets_prefix}/*",
    ]
  }
}
```
