### Secret Storage and Injection ###

Most configuration contains secrets of some kind such as database passwords or OAuth tokens. cStore supports secret injection into tokens within any `*.json` or `*.env` file. The secrets are stored and retreived from AWS Secrets Manager by default.

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



2. Push the file to AWS S3 using the `-m` flag to store secrets. This action will remove all secrets from the file.
```
$ cstore push {{file}} -m
```
3. Pull the file from AWS S3 using the `-i` flag enabling secret injection. This will create a `*.secrets` file containing the injected secrets along side the file.
```
$ cstore pull {{file}} -i
```
NOTE: When using the `-e` and `-i` flag together on a `.env` file the secrets will be injected and exported. 