

### Catalog ###

A list of fields tracked in the `cstore.yml` file to make managing configuration simpler. The catalog should not be modified by hand without understanding the fields.

| Property | Type | Values | State |Description |
|-|-|-|-|-|
| version | `string` | v1, v2, v3, v4 | yes | Catalog version used by the CLI to determine the format of the catalog. |
| context | `string` || yes | The unique id used in the remote store or vault to link the catalog to the remotely stored data. This vault is often the key prefix for remotely stored data. |
| file.path | `string` || yes | The local file path of the remotely stored data relative to the catalog. |
| file.alternatePath | `string` || no | An alternate local file path to restore the data to when a file is retrieved.|
| file.store | `string` | `aws-secret`, `aws-secrets`, `aws-s3`, `aws-parameter`, `source-control` | yes | The CLI key identifying the current remote storage solution. |
| file.isRef| `bool` | `true`,`false` | yes | A flag indicating when the file is referencing another linked catalog or remotely stored data.  |
| file.deleteAfterPush| `bool` | `true`,`false` | no | A flag indicating if the local file should be deleted after pushing the data to the remote store. This avoids secrets living permanently on other machines. |
| file.type | `string` | `env`, `json`, etc...| yes | A flag indicating the local file type. |
| file.data | `map[string]string` |  | yes | Key/value pairs of data specific to the chosen storage or vault solution. |
| file.tags | `[]string` | | no | File tags used to make pulling and pushing data easier. |
| file.vault.access | `string` | `env`, `osx-keychain` | no | The local vault used to save remote store and vault credentials. |
| file.vault.secrets | `string` | `aws-secrets-manager` | yes | The remote vault used to store tokenized secrets when the remote store is not secure. |
| files.version | `[]string` | | yes |A list of data versions stored remotely. Often using semver to specify config versions that should be pulled for specific builds. |

