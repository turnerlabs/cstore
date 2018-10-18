## Goals ##

* Create a simple CLI that can be shared among different storage solutions.
* Provide environment configuration management within repository context. (manage configuration as code, CaC)
* Expose store interface that can support multiple storage solutions.
* Expose a vault interface that can support a variety of credential, encryption key, and secret storage solutions.
* Offer a scriptable option for pulling configuration locally, within a container, or from anywhere.
* Store and access file configuration using versions.
* Provide a simple solution for sharing configuration among developers.
* Support `.env` files to align with Docker.
* Allow custom developer specific configuration to be stored in a secure location.
* Make a quick, secure way to get an app up and running with simple configuration management.

## Non-Goals ##

* Support store access logging or auditing. (should be the store's responsibility)
* Support the management (add, edit, delete) of credenatils in any vault. (should be the vault's responsibilty) However, cStore vaults may make it possible to add or edit vault creds for ease of use. 