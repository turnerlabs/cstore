# README

The cStore CLI provides commands to push config files `$ cstore push service/dev/.env` to remote [storage](docs/STORES.md). The pushed files are replaced by a catalog file, `cstore.yml`, that understands resource needs, storage solution, file encryption, and other details making restoration locally or by a resource including lambda functions, docker containers, ec2 instances as simple as `$ cstore pull -t dev`.

`*.env` and `*.json` files are special file types whose contents can be parameterized with secret tokens, encrypted, and stored separately from the configuration.

### Example ###
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
│       └── qa
│           └── .env (stored)
│           └── .cstore (ghost)
│           └── fargate.yml
│           └── docker-compose.yml
```
The `cstore.yml` catalog and hidden `.cstore` ghost files reference the stored `*.env` files. Secrets no longer need to be checked into source control.

When the repository has been cloned or the project shared, running `$ cstore pull` in the same directory as the `cstore.yml` catalog file or any of the `.cstore` ghost files will locate, download, and decrypt the configuration files to their respective original location restoring the project's environment configuration.

| Demo |  | Audio |
|---|---|---|
| [watch](https://youtu.be/QBVoU4kSYeM) | Store Configs in Parameter Store with secrets in Secrets Manager | no |
| [watch](https://youtu.be/yL5xFBOQ7lg)| Store Configs in S3 with secrets in Secrets Manager | no |
| [watch](https://youtu.be/vpNii5Y0yNg) | Get Configs With Secrets Injected | no |

## How to Use (3 minutes) ##

Ensure a supported [storage](docs/STORES.md) solution is already set up and available.

| OS | Install/Upgrade |
|----|----|
| Mac | `$ sudo curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.3.1-alpha/cstore_darwin_amd64 && sudo chmod +x /usr/local/bin/cstore` |
| Linux | `$ sudo curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.3.1-alpha/cstore_linux_386 && sudo chmod +x /usr/local/bin/cstore` |
| Windows | `wget https://github.com/turnerlabs/cstore/releases/download/v3.3.1-alpha/cstore_windows_amd64.exe` (add download dir to the PATH environment variable) |

The first push creates a catalog file that can be checked into source control. Subsequent commands executed in the same directory will use the existing catalog.

By default, cStore will use the [AWS credential chain](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) and store configuration in AWS Parameter Store.

### Store Files (choose remote storage solution) ###
Using [AWS Parameter Store](docs/PARAMETER.md) for config and [AWS Secrets Manager](docs/SECRETS.md) for credentials. (default)
```bash
$ cstore push service/dev/.env -s aws-paramter 
```
 Using [AWS S3](docs/S3.md) for config and [AWS Secrets Manager](docs/SECRETS.md) for credentials.
```bash
$ cstore push service/dev/.env -s aws-s3
```
Using [Source Control](docs/SOURCE_CONTROL.md) for config and [AWS Secrets Manager](docs/SECRETS.md) for credentials.
```bash
$ cstore push service/dev/.env -s source-control
```

Multiple files can be discovered and pushed in one command. If needed, replace `service` with a custom environments folder or `.` to search all project sub folders.
```bash
$ cstore push $(find service -name '*.env')
```

### Restore Files ###
```bash
$ cstore pull service/dev/.env
```
or 
```bash
$ cstore pull -t dev
```

Instead of restoring files locally, export environment variables listed inside the files. 
```bash
$ eval $( cstore pull service/dev/.env -g terminal-export ) # works for '*.env' files only
```

## Advanced Usage ##

* [Migrate from v1 to v3](docs/MIGRATE.md) (breaking changes)
* [Set Up S3 Bucket](docs/S3.md)
* [Access Config inside Docker Container](docs/DOCKER.md)
* [Access Config inside Lambda Function](docs/LAMBDA.md)
* [Access Config using Code Library](docs/LIBRARY.md)
* [Storing/Injecting Secrets](docs/SECRETS.md)
* [Ghost Files (.cstore)](docs/GHOST.md)
* [Tagging Files](docs/TAGGING.md)
* [Versioning Files](docs/VERSIONING.md)
* [Linking Catalogs](docs/LINKING.md)
* [CLI Commands and Flags](docs/CLI.md)

## Additional Info ##

* [Terms](docs/TERMS.md)
* [Stores](docs/STORES.md)
* [Vaults](docs/VAULTS.md)
* [Terraform State Files](docs/TERRAFORM.md)

## Project ##

* [Goals](docs/GOALS.md)
* [Integration Testing](docs/TESTING.md)
* [Publish Release](docs/PUBLISH.md)
