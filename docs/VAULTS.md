## Vaults - Supported Credential, Encryption Key, or Secret Retrieval Locations ##

* Environment Variables (env)
* Encrypted File (file)
* [AWS Secrets Manager](SECRETS.md)(aws-secrets-manager)
* OSX Keychain (osx-keychain)

NOTE: Not all operations like set, get, and delete are currently supported by all vaults. Only operations that were needed at the time of development were implemented.