## Goals ##

* Create a simple CLI that can be shared among different storage solutions.
* Ensure configuration management is simple for developers, simple for consuming services, and secure for secrets.
* Support `.env` files to align with Docker.
* Provide environment configuration management within repository context. (manage configuration as code, CaC)
* Expose store interface that can support multiple storage solutions.
* Expose a vault interface that can support a variety of credential, encryption key, and secret storage solutions.
* Offer a scriptable option for pulling configuration locally, within a container, inside a lambda function, or from anywhere.
* Store and access file configuration using versions.
* Provide a simple solution for sharing configuration among developers.
* Allow custom developer specific configuration to be stored in a secure location.
* Make a quick, secure way to get an app up and running with simple configuration management.
* Enable configuration to be stored separately from secrets to facilitate having different levels and complexity of access.
* Maintain a separation between application code and configuration retrieval.
* Avoid being opinionated on where to store configuration or how an app should import configuration. 

## Non-Goals ##

* Support store access logging or auditing. (should be the store's responsibility)
* Support the management (add, edit, delete) of credenatils in any vault. (should be the vault's responsibilty) However, cStore vaults may make it possible to add or edit vault creds for ease of use. 