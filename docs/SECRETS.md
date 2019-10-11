## Secret Storage and Injection ##

Most configuration contains secrets of some kind such as database passwords or OAuth tokens. cStore supports secret injection for any `*.json` or `*.env` file. The secrets are stored and retreived from AWS Secrets Manager by default.

| CLI Key | Description | Supports | Secret Key |
|-|-|-|-|
|`aws-secret-manager`| All config values are stored in a single secret. | `.env`, `.json`|`/{config_context}/{env}` |
|`aws-secrets-manager`| Each config value is stored in a separate secret. | `.env`, `.json` | `/{config_context}/{env}/{var}` |

IMPORTANT: Secrets are created and updated in Secrets Manager, but never deleted by cStore. Due to the sensitive nature of secrets, a user must delete the secrets through the console.

### How To ###

1. Place tokens with secrets in the file using the format `{{ENV/KEY::SECRET}}`.

#### `*.env` example #### 
Tokens are only supported in values not keys.
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

4. Set the `config_context` avariable to the cstore context and apply this policy to any resource role to allow AWS Secrets Manager access.

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