## Source Control Store ##

This solution allows configuration files to be stored in source control after removing and storing secrets in a secure vault.

| CLI Key | Description | Supports | File Key |
|-|-|-|-|
|`source-control`| Secrets are removed from configuration and stored in a vault. | `.json`, `.env` | reative path |

After creating `.env` or `.json` in a repository, `$ cstore push .env -s source-control` will register the file with the catalog, `cstore.yml` and remove [tokenized](SECRETS.md) secrets from the file storing them in Secrets Manager.

Set the `config_context` variable to the cstore context and apply this policy to any resource role to allow AWS Secrets Manager access.

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