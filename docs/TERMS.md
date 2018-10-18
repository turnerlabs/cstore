### Important Terms ###

* `store` is a remote location that can store data.
* `vault` is an application or service that holds credentials, secrets, and/or encryption keys used for store access or secret injection for configuration files.
* `catalog` is a `cstore.yml` file that references the files stored remotely. It can be checked into the repo or stored with a project.
* `file` is a local file, can be an `*.env`, that is pushed to a remote store optionally using a vault to store encrytion keys or configuration secrets.
* `ghost file` is a local file, called `.cstore`, created by cStore along side pushed files. It allows cStore commands to be executed in the same directory as the pushed files without being in the same directory as the `cstore.yml` file.