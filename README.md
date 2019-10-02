# README

The cStore CLI provides a command to push config files `$ cstore push service/dev/.env` to remote [storage](docs/STORES.md). The pushed files are replaced by a, `cstore.yml` file, that remembers the storage location, file encryption, and other details making restoration locally or by a resource as simple as `$ cstore pull -t dev`.

`*.env` and `*.json` files are special file types whose secrets can be [tokenized](docs/SECRETS.md), encrypted, stored separately from the configuration, and injected at runtime.

<details>
  <summary>Repository Example</summary>

```
├── project
│   ├── components
│   ├── models
│   ├── main.go
│   ├── Dockerfile 
│   ├── cstore.yml (catalog)
│   └── service
│       └── dev
│       │   └── .env (stored)
│       |   └── .cstore (ghost)
│       |   └── fargate.yml
│       |   └── docker-compose.yml
│       │
│       └── prod
│           └── .env (stored)
│           └── .cstore (ghost)
│           └── fargate.yml
│           └── docker-compose.yml
```
The `cstore.yml` catalog and hidden `.cstore` ghost files reference the stored `*.env` files. Secrets no longer need to be checked into source control.

When the repository has been cloned or the project shared, running `$ cstore pull` in the same directory as the `cstore.yml` catalog file or any of the `.cstore` ghost files will locate, download, and decrypt the configuration files to their respective original location restoring the project's environment configuration.

Example: `cstore.yml`
```yml
version: v3
context: project
files:
- path: service/dev/.env
  store: aws-s3
  isRef: false
  type: env
  data:
    AWS_S3_BUCKET: my-bucket
    AWS_STORE_KMS_KEY_ID: ""
    AWS_VAULT_KMS_KEY_ID: aws/secretsmanager
  tags:
  - service
  - dev
  vaults:
    access: env
    secrets: aws-secrets-manager
  versions: []
- path: service/prod/.env
  store: aws-parameter
  isRef: false
  type: env
  data:
    AWS_STORE_KMS_KEY_ID: aws/ssm
    AWS_VAULT_KMS_KEY_ID: aws/secretsmanager
  tags:
  - service
  - prod
  vaults:
    access: env
    secrets: aws-secrets-manager
  versions: []
```

</details>

<details>
  <summary>Install/Upgrade</summary>

| OS |  |
|----|----|
| Mac | `$ sudo curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.3.1-alpha/cstore_darwin_amd64 && sudo chmod +x /usr/local/bin/cstore` |
| Linux | `$ sudo curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.3.1-alpha/cstore_linux_386 && sudo chmod +x /usr/local/bin/cstore` |
| Windows | `wget https://github.com/turnerlabs/cstore/releases/download/v3.3.1-alpha/cstore_windows_amd64.exe` (add download dir to the PATH environment variable) |

</details>

## Authenticate ##

[AWS credential chain](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) is used for Authentication.

```bash
$ export AWS_REGION=us-east-1
$ export AWS_PROFILE=user-profile
```

## Store App Configuration ##

Ensure a [storage](docs/STORES.md) solution is available and supports the configuration file type.

<details>
  <summary>Example .env</summary>

```
HEALTHCHECK=/ping
MONGO_URL=mongodb://{{dev/user::appuser-dev}}:{{dev/password::3lkjr4kfdro4df}}@example-server.mongodb.net:30000/example-dev
API_KEY={{dev/token::82f6f303-9e00-4a8c-be26-b9d06781d844}}
API_URL=https://dev.api.example-service.com
CONTACT=team@example-service.com
```
</details>

```bash
$ cstore push service/dev/.env -s aws-parameter 
```
```bash
$ cstore push service/dev/.env -s aws-s3
```
```bash
$ cstore push service/dev/.env -s source-control
```

<details>
  <summary>Example config.json</summary>

```json
{
    "db_url" : "mongodb://{{dev/user::app_user}}:{{dev/password::4kdnow55jdjnk3nd}}@example-server.mongodb.net:30000/example-dev",
    "api_key": "{{dev/key::82f6f303-9e00-4a8c-be26-b9d06781d844}}",
    "healthcheck": "/ping",
    "contact": "team@example-service.com"
}
```
</details>

```bash
$ cstore push service/dev/config.json -s aws-s3
```

Multiple files can be discovered and pushed in one command. Replace `service` with the environments folder or `.` to search all project sub folders.
```bash
$ cstore push $(find service -name '*.env')
```

When using `-i` during a push, [tokenized](docs/SECRETS.md) secrets are removed and stored in AWS Secrets Manager.

## Restore App Configuration ##

Restore Config Files (any type)

```bash
$ cstore pull service/dev/.env
```
```bash
$ cstore pull -t dev
```

Export Environment Variables (`.env`)
```bash
$ eval $( cstore pull service/dev/.env -g terminal-export )
```

Output Task Definition JSON Formats (`.env`)
```bash
$ cstore pull -t dev -g task-def-env
```
```bash
$ cstore pull -t dev -g task-def-secrets --store-command refs # When using AWS Parameter Store, this command generates the json needed for the task definition allowing secrets to be injected into the container at run time.
```

#### How To ####

* [Inside Docker Container](docs/DOCKER.md)
* [Inside Lambda Function](docs/LAMBDA.md)
* [Using Application Memory](docs/LIBRARY.md)

## More ##

<details>
  <summary>Learning Basics</summary>

* [Terminology](docs/TERMS.md)
* [Storage Solutions](docs/STORES.md)
* [Vault Solutions](docs/VAULTS.md)

| Demo |  |
|---|---|
| [watch](https://youtu.be/QBVoU4kSYeM) | Store Configs in Parameter Store with secrets in Secrets Manager |
| [watch](https://youtu.be/yL5xFBOQ7lg)| Store Configs in S3 with secrets in Secrets Manager |
| [watch](https://youtu.be/vpNii5Y0yNg) | Get Configs With Secrets Injected |

</details>

<details>
  <summary>Useful Options</summary>

* [Tagging Files](docs/TAGGING.md)
* [Storing/Injecting Secrets](docs/SECRETS.md)
* [Versioning Files](docs/VERSIONING.md)
* [Linking Catalogs](docs/LINKING.md)
* [CLI Commands and Flags](docs/CLI.md)
* [S3 Bucket Store Terraform](docs/S3.md)
* [Ghost Files (.cstore)](docs/GHOST.md)
* [Terraform State Files](docs/TERRAFORM.md)
* [Migrate from v1 to v3](docs/MIGRATE.md) (breaking changes)
</details>

<details>
  <summary>Project Details</summary>

* [Goals](docs/GOALS.md)
* [Integration Testing](docs/TESTING.md)
* [Publish Release](docs/PUBLISH.md)
</details>