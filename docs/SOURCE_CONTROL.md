## Using Source Control Store ##

Source Control requires a source control repository.

After creating `.env` or `.json` in a repository, `$ cstore push .env -s source-control` will register the file with the catalog, `cstore.yml` and extract/remove [tokenized](SECRETS.md) secrets from the file pushing them to Secrets Manager.

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