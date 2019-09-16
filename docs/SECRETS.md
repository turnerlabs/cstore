### Secret Storage and Injection ###

Most configuration contains secrets of some kind such as database passwords or OAuth tokens. cStore supports secret injection into tokens within any `*.json` or `*.env` file. The secrets are stored and retreived from AWS Secrets Manager by default.

IMPORTANT: Secrets are created and updated in Secrets Manager, but never deleted by cStore. Due to the sensitive nature of secrets, a user must delete the secrets through the console.

IMPORTANT: Ensure the users or roles performing the following actions have access to the AWS Secrets Manager secrets and the KMS Key ID used by Secrets Manager.

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



2. Push the file to AWS S3 or Parameter Store using the `-m` flag to store secrets. This action will remove all secrets from the file.
```
$ cstore push {{file}} -m
```
3. Pull the file from AWS S3 or Parameter Store using the `-i` flag enabling secret injection. This will create a side car file called `*.secrets` containing the injected secrets.
```
$ cstore pull {{file}} -i
```
NOTE: When using the `-e` and `-i` flag together on a `.env` file the secrets will be injected and exported. 