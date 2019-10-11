## Vault Solutions ## 

A comparison of supported vault solutions. Vaults can manage credentials, Access Vault, or configuration secrets, Secrets Vault.

NOTE: Delete functionality is not currently supported by vaults to avoid deleting sensitive information accidentally.


| | [AWS Secrets Manager](SECRETS.md) | OSX Keychain | Environment | Encrypted File | 
|-|-|-|-|-|
| CLI Flag | `-x` | `-c` | `-c` | `-c` |
| CLI Key | `aws-secrets-manager`, `aws-secret-manager` | `osx-keychain` | `env` | `file` |
| Description | Secures config secrets in AWS Secrets Manager. | Secures access credentails in OSX Keychain. | Reads access credentails from environment variables. | Secures access credentials in a local encrypted file. |
| Access Vault | no | yes | yes | yes |
| Secrets Vault | yes | no | no | no |

