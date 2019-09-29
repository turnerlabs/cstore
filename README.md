# README

The cStore CLI provides commands to push config files `$ cstore push {{FILE}}` to remote storage. The pushed files are replaced by a catalog file, `cstore.yml`, that understands resource needs, storage solution, file encryption, and other details making restoration locally or by a resource as simple as `$ cstore pull {{FILE}}`.

`*.env` and `*.json` files are special file types whose contents can be parameterized with secret tokens, encrypted, and stored.

TL;DR: cStore encrypts and stores configuration remotely using storage solutions like AWS S3, Parameter Store, Secrets Manager; and restores the configuration anywhere including local machines, lambda functions, or Docker containers using a catalog file `cstore.yml` which can be checked into source control safely without exposing secrets.

### Example ###
```
├── project
│   ├── components
│   ├── models
│   ├── main.go
│   ├── Dockerfile 
│   ├── cstore.yml (catalog)
│   └── env
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
The `cstore.yml` catalog and hidden `.cstore` ghost files take the place of the stored `*.env` files. The `*.env` files or secrets within the files can be encrypted and stored remotely no longer checked into source control.

When the repository has been cloned or the project shared, running `$ cstore pull` in the same directory as the `cstore.yml` catalog file or any of the `.cstore` ghost replacement files will locate, download and decrypt the `*.env` files to their respective original location restoring the project's environment configuration.

## How to Use (3 minutes) ##

Ensure a supported [storage](docs/STORES.md) solution is already set up and available.

| OS | Install/Upgrade |
|----|----|
| Mac | `$ sudo curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.1.0--alpha/cstore_darwin_amd64 && sudo chmod +x /usr/local/bin/cstore` |
| Linux | `$ sudo curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.1.0--alpha/cstore_linux_386 && sudo chmod +x /usr/local/bin/cstore` |
| Windows | `wget https://github.com/turnerlabs/cstore/releases/download/v3.1.0--alpha/cstore_windows_amd64.exe` (add download dir to the PATH environment variable) |

The first push creates a catalog file that can be checked into source control. Subsequent commands executed in the same directory will use the existing catalog.

By default, cStore will use the [AWS credential chain](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) and store configuration in AWS Parameter Store.

### Store Files (choose remote storage solution) ###
Using [AWS Parameter Store](docs/PARAMETER.md) for config and [AWS Secrets Manager](docs/SECRETS.md) for credentials. (default)
```bash
$ cstore push {{FILE}} -s aws-paramter 
```
 Using [AWS S3](docs/S3.md) for config and [AWS Secrets Manager](docs/SECRETS.md) for credentials.
```bash
$ cstore push {{FILE}} -s aws-s3
```
Using [Source Control](docs/SOURCE_CONTROL.md) for config and [AWS Secrets Manager](docs/SECRETS.md) for credentials.
```bash
$ cstore push {{FILE}} -s source-control
```

Multiple files can be discovered and pushed in one command. If needed, replace `service` with a custom environments folder or `.` to search all project sub folders.
```bash
$ cstore push $(find service -name '*.env')
```

### Restore Files ###
```bash
$ cstore pull {{FILE}}
```

Instead of restoring files locally, export environment variables listed inside the files. 
```bash
$ eval $( cstore pull {{FILE}} -g terminal-export ) # works for '*.env' files only
```

## Advanced Usage ##

* [Migrate from v1 to v2](docs/MIGRATE.md) (breaking changes)
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
