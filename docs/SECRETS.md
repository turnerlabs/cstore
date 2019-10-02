### Secret Storage and Injection ###

Most configuration contains secrets of some kind such as database passwords or OAuth tokens. cStore supports secret injection into tokens within any `*.json` or `*.env` file. The secrets are stored and retreived from AWS Secrets Manager by default.

IMPORTANT: Secrets are created and updated in Secrets Manager, but never deleted by cStore. Due to the sensitive nature of secrets, a user must delete the secrets through the console.

IMPORTANT: Ensure the users or roles performing the following actions have access to the AWS Secrets Manager secrets and the KMS Key ID used by Secrets Manager.

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

1. Place tokens with secrets in the file using the format `{{ENV/KEY::SECRET}}`.

#### `*.env` example #### 
Tokens are only supported in the value.
```
MONGO_URL=mongodb://{{dev/user::my_app_user}}:{{dev/password::123456}}@ds999999.mlab.com:61745/database-name
```
#### `.json` example #### 
Tokens are only supported in objects and nested objects containing properties with string values.
```json
{
    "database" : {
        "url" : "mongodb://{{dev/user::my_app_user}}:{{dev/password::123456}}@ds999999.mlab.com:61745/database-name"
    }
}
```

2. Push the file to a AWS S3, Parameter Store, or Source Control store to extract and store secrets. This action will remove all secrets from the file.
```
$ cstore push {{FILE}}
```
3. Pull the file from the store using the `-i` flag enabling secret injection from the vault. This will create a side car file called `*.secrets` containing the injected secrets.
```
$ cstore pull {{FILE}} -i
```