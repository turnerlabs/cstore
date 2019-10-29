### CLI Flags and Commands ###

| Flag | Env Variable | Values | Description |
|------|---------|--------|-------------|
| `-p` | `CSTORE_PROMPT` | `true/false` | Prompt user for additional configuration options. (default: `false`) |
| `-s` | `CSTORE_STORE` | `$ cstore stores` | Set remote store to use during file push. (default: `aws-s3`) |
| `-x` | `CSTORE_SECRETS` | `$ cstore vaults` | Set integration for storing and injecting secrets into configuration. (default: `aws-secrets-manager`) |
| `-c` | `CSTORE_ACCESS` | `$ cstore vaults` | Set integration for retrieving store credentials. (default: `env` *) |
| `-f` | `CSTORE_CATALOG` | `{file}.yml` | Set a different catalog file name to use. (default: `cstore.yml`) |
| `-t` | `CSTORE_TAGS` | <code>"tag-1&#124;tag-2"</code> | Set <code>&#124;</code> or `&` delimited list of tags to identify files. If any <code>&#124;</code> is used during a pull request, only files tagged with all listed tags will be retrieved. (default: file path folder names) |
| `-v` | | <code>"v0.2.0-rc"</code> | Set version of file to pull or push. |
| `-a` | `CSTORE_ALT` | `{path}/{file}` | Set alternate location for the file to be restored. When used during a push, the alternate location will be saved, but when used during a pull, the alternate location will override any stored locations. |
| `-e` | `CSTORE_EXPORT` | | Send environment variables from store prefixed with export commands to `stdout` instead of writing file to disk. (default: `restore file`) |
| `-g` | `CSTORE_FORMAT` | `terminal-export/task-def-secrets/task-def-env/json-object` | Send environment variables from store using specified format to `stdout` instead of writing file to disk. |
| `-n` | `CSTORE_NO-OVERWRITE` | | Skip pulling environment variables already exported in the current environment. (default: `all`) |
| `-d` | `CSTORE_DELETE` | `true/false` | Set automatic deletion of local files after successful push. (default: `false`) |
| `-h` | | | List command documentaion. |
| `-i` | `CSTORE_INJECT-SECRETS` | `false`| Inject secrets into tokenized configuration. [read more](SECRETS.md)|
| `-m` | `CSTORE_MODIFY-SECRETS` | `false`| Inject tokenized secrets into configuration. [read more](SECRETS.md)|
| `-v` | | `false`| Display a list of versions for each file. |
| `-g` | | `false`| Display a list of tags for each file. |
| `-l` | `CSTORE_LOGGING` | `false`| Convert `stderr` output to be more log friendly instead of terminal friendly. |
| `--store-command`| `CSTORE_STORE-COMMAND` | varies by store | Command to send to store. The command is ignored if not supported by a store.|

\* When the `env` vault is used, the store will typically default to pulling access information environment variables.

| Command | Args | Flags | Description |
|---------|------|-------|-------------|
| `push` | {file_1} {file_2} ... | `-p -s -x -c -d -f -t -a -v` | Store file(s) remotely. During initial push the store and vaults will be saved. |
| `pull` * | {file_1} {file_2} ... | `-p -e -n -f -t -c -v -i -m -g --store-command` | Restore file(s) locally. |
| `purge` * | {file_1} {file_2} ... | `-p -f -t` | Purge file(s) remotely. |
| `list` | | `-f -t -k -l` | List file(s) stored remotely. |
| `stores` * | {store_name} | | List available stores or store details. |
| `vault` * | {vault_name} | | List available vaults or vault details. |
| `version` | | | Display version. |

\* When arguments are not supplied, command applies to all objects.

All commands are executed against the default `cstore.yml` or user specified `-f mycatalog.yml` catalog file and will not affect any other catalogs.
