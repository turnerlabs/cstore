# README

The cStore CLI provides commands to push files `$ cstore push {file}` to remote storage. The pushed files are replaced by a catalog file, `cstore.yml`, that understands folder context, file encryption, and other details making restoration as simple as `$ cstore pull`.

`.env` files are treated as a special file type allowing their contents to be parameterized, encrypted, and pushed to stores like Harbor and AWS Parameter Store.

TL;DR: cStore encrypts and stores environment configuration remotely using stores like AWS S3 and restores the configuration anywhere using a catalog file named `cstore.yml` which can be checked into source control without exposing secrets.

### Important Terms ###

* `store` is a remote location that can store data.
* `vault` is an application or service that holds credentials, secrets, and/or encryption used for store access or secret injection for configuration files.
* `catalog` is a `cstore.yml` file checked into the repo which references the files stored.
* `file` is a local file that is pushed to a remote store optionally using a vault to store encrytion keys or configuration secrets.

### Example ###
```
├── repo
│   ├── components
│   ├── models
│   ├── main.go
│   ├── Dockerfile 
│   ├── docker-compose.yml
│   ├── cstore.yml
│   └── env
│       └── dev
│       │   └── .env
│       │
│       └── qa
│           └── .env
```
The `.env` files for each environment referenced by a `Dockerfile` can be encrypted and stored in an AWS S3 bucket without being checked into a repository.

When the repository has been cloned, running `$ cstore pull` in the same directory as the `cstore.yml` catalog file will locate, download and decrypt the `.env` files to their respective original paths restoring the project's environment.

## Set Up S3 Bucket (default store) ##
1. Create AWS S3 bucket using [Terraform]
(https://github.com/turnerlabs/terraform-s3-employee)
2. cStore will prompt for the S3 bucket name on the initial push of any file.
3. Add users and/or container roles to the Terraform script to provide access to the S3 Bucket.
```yml
 # Email address are case sensitive.
  role_users = [
    "{{AWS_USER_ROLE}}/{{USER_EMAIL_ADDRESS}}",
    "{{AWS_CONTAINER_ROLE}}/*",
  ]
```
4. (optional) Create AWS KMS key for encryption

## How to Use (3 minutes) ##

#### Install/Upgrade ####
mac: `$ curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v1.0.1-rc/cstore_darwin_amd64 && chmod +x /usr/local/bin/cstore`

linux: `$ curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v1.0.1-rc/cstore_linux_386 && chmod +x /usr/local/bin/cstore`

The first push creates a catalog file in the same directory that can be checked into source control. Subsequent commands executed in the same directory will use the existing catalog.

By default, cStore will prompt for an AWS profile and S3 bucket.

### Store Files ###
```bash
$ cstore push dir/filename
```

### Restore Files ###
```bash
$ cstore pull
```

Instead of restoring files locally, export environment variables listed inside the file. 
```bash
$ eval $( cstore pull -e ) # works with stored `.env` files only
```

## Advanced Usage ##

### Store Configuration ###

To configure a store's credentials or encryption settings use `-p` on the commandline and follow the prompts. (Flags used during a `push` will be saved to the catalog and flags used during a `pull` can override catalog settings.)

### Tagging Files / Configuration ###

A single catalog may store multiple files for any given folder context, but a complete restore, `pull`, of the catalog may not always be needed. For this purpose, files can be tagged when pushed using a pipe delimited list of tags `$ cstore push {file} -t "dev|test"`. At this point, files can be listed, purged, or restored by tags. For example, `$ cstore pull -t dev` only restores files containing a `dev` tag.

The commands `list`, `purge`, and `pull` accept a list of tags as well. Pipes between each tag indicate an `or` operation that will only manipulate the files that contain at least one of the specified tags. The `,`/`and` tag operation is not currently supported. 

If no tags are pushed, the file will keep the tags from the last push. If tags are pushed, the files previous tags will be overwritten.

Multiple tags should be encapsulated by quotes. (i.e. `"dev|secure|vscode"`)

### Versioning Files / Configuration ###

A file can be versioned allowing it to be associated to a specific build or deployment. Version specific files/configuration makes code deployments and rollbacks simpler by applying the correct configuration to the correct deployment without having to track configuration separately.

Files/configuration can be versioned and retrieved using simple commands. Versions can also be coupled with tags allowing tag specific versions for environments.

Version file: 
`$ cstore push {file} -t dev -v v0.1.0-beta`

Restore file version: 
`$ cstore pull {file} -t dev -v v0.1.0-beta`

List file versions:
`$ cstore list`

### Linking Catalogs ###

Sometimes, it can be useful to restore multiple catalogs with a single command. For this purpose, catalogs can be linked to a parent catalog which allows restoration of the parent catalog to restore child catalogs at the same time.
```bash
$ cstore push dir/cstore.yml
$ cstore pull
```
Purging of a parent catalog will not purge linked child catalogs, but listing a parent's contents will include the contents of the linked children.

If a linked catalog is tagged, the linked catalog's files will only be accessed when the tag is used. Otherwise, tags will flow down and be applied to linked catalog files.

### CLI Flags and Commands ###

| Flag | Values | User Config | Description |
|------|--------|-------------|-------------|
| `-p` | `true/false` | `prompt` | Prompt user for additional configuration options. (default: `false`) |
| `-s` | `$ cstore stores` | `store` | Set remote store to use during file push. (default: `aws-s3`) |
| `-x` | `$ cstore vaults` | `encrypt` | Set vault used to access store encryption keys. (default: `env` *) |
| `-c` | see `vaults` cmd | `creds` | Set vault used to access store credentials. (default: `env` *) |
| `-f` | `{file}.yml` | `file` | Set file name to target during command. (default: `cstore.yml`) |
| `-t` | <code>"tag-1&#124;tag-2"</code> | | Set pipe delimited list of tags to identify files. |
| `-v` | <code>"v0.2.0-beta"</code> | | Set version of file to pull or push. |
| `-a` | `{file}` | | Set alternate file path and name to clone the file contents to during a restore. |
| `-e` | | | Send environment variables from store to `stdout`. (default: `file`) |
| `-d` | `true/false` || Delete local file(s) after successful push. (default: `false`) |
| `-h` | | | List command documentaion. |

\* When an `env` vault is used, the store will typically default to pulling access information from files or environment variables.

| Command | Args | Flags | Description |
|---------|------|-------|-------------|
| `push` | {file_1} {file_2} ... | `-p -s -x -c -d -f -t -a -v` | Store file(s) remotely. During initial push the store and vaults will be saved. |
| `pull` * | {file_1} {file_2} ... | `-p -e -f -t -c -v` | Restore file(s) locally. |
| `purge` * | {file_1} {file_2} ... | `-p -f -t` | Purge file(s) remotely. |
| `list` | | `-f -t` | List file(s) stored remotely. |
| `stores` * | {store_name} | | List available stores or store details. |
| `vault` * | {vault_name} | | List available vaults or vault details. |
| `version` | | | Display version. |

\* When arguments are not supplied, command applies to all objects.

All commands are executed against the default `cstore.yml` or user specified `-f mycatalog.yml` catalog file and will not affect any other catalogs.


## How to Load Configuration in a Docker container running in AWS ##

1. Add [docker-entrypoint.sh](docker-entrypoint.sh) script to the repo. 
2. Replace `./my-application` in the script with the correct application executable. 
```bash
exec ./my-application
```
3. Use the `ENTRYPOINT` command in place of the `CMD` command in Dockerfile to run the shell script. 
```docker
ENTRYPOINT ["./docker-entrypoint.sh"]
```
4. Update the `Dockerfile` to install [cStore](https://github.com/turnerlabs/cstore/releases/download/v0.3.6-alpha/cstore_linux_amd64) for Linux (or the appropriate os) adding execute permissions.
```docker
RUN curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v1.0.1-rc/cstore_linux_386 && chmod +x /usr/local/bin/cstore
```
5. Update the `docker-compose.yml` file to specify which environment config should be pulled by the `docker-entrypoint.sh` script.    
```docker
    environment:
      CONFIG_ENV: dev
      AWS_REGION: us-east-1
```
6. In the same folder as the `Dockerfile`, use cStore to push the `.env` files to an AWS S3 bucket with a `dev` tag. Check the resulting `cstore.yml` file into the repo.
7. Set up [S3 Bucket](#set-up-s3-bucket-default-store) permissions to allow AWS container role access.
```yml
module "s3_employee" {
  source = "github.com/turnerlabs/terraform-s3-employee?ref=v0.1.0"

  bucket_name = "{{S3_BUCKET}}"

  # Email address are case sensitive.
  role_users = [
    "{{AWS_USER_ROLE}}/{{USER_EMAIL_ADDRESS}}",
    "{{AWS_CONTAINER_ROLE}}/*",
  ]
}

```
8. Set up the AWS container role policy permissions to allow S3 bucket access.
```yml
data "aws_iam_policy_document" "app_policy" {
  statement {
    effect = "Allow"

    actions = [
      "s3:Get*",
    ]

    resources = [
      "{{AWS_S3_BUCKET_ARN}}/*",
    ]
  }
}
```
## User Configuration ##

If `$HOME/.cstore/user.yml` file is created, CStore will use user defaults when specific flags are not specified.

```
# default store to use when pushing a file the first time
store: aws-s3 

# override the file specific credentials and encryption vaults locally.
credentials: env
encryption: osx-keychain

# set a custom file for cstore.yml files
file: mystore.yml
```

## Supported Stores - Storage Locations ##

* AWS S3 Bucket (aws-s3)
* AWS Parameter Store (aws-parameter)
* [Harbor](docs/HARBOR.md) (harbor)

## Supported Vaults - Credential and Encryption Configuration ##

* OSX Keychain (osx-keychain)
* AWS Parameter Store (aws-parameter)
* AWS Profiles (env)
* AWS SDK (aws-sdk)
* Environment Variables (env)
* Encrypted File (file)

## Publish GitHub Release ##

NOTE: semver tag format required `v1.0.0-rc`
```bash
$ git tag {{TAG}}
$ git push origin {{TAG}}
$ ./create_darwin_build.sh
```
Once the build is complete the {{TAG}} release will be published to GitHub. 

IMPORTANT: The Mac build is not part of the build process due to required libraries not included on the linux build server. The `create_darwin_build.sh` script must be run on a Mac and manually uploaded to the new GitHub release to complete the process.

## Goals ##

* Create a simple CLI shared among configuration stores.
* Provide environment configuration management within repository context. (manage configuration as code, CaC)
* Expose a store interface allowing implementation for multiple storage solutions.
* Expose a vault interface allowing implementation for a variety of local credential and encryption key storage solutions.
* Offer a scriptable option for pulling configuration locally or within a container.
* Store and access configuration or files by version numbers.
* Provide a simple solution for sharing configuration among developers.
* Support `.env` files to align with Docker.
* Allow custom developer specific configuration to be stored in a secure location.
* Make a quick, secure way to get an app up and running with simple configuration management.

## Non-Goals ##

* Support store access logging or auditing. (should be the store's responsibility)

## Terraform Files ##

Terraform state files contain secrets and can also be encrypted and stored using cstore. However, a better option may be Terraform remote [state](https://github.com/turnerlabs/terraform-remote-state).
