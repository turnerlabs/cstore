### Versioning Files / Configuration ###

A file can be versioned allowing it to be associated to a specific build or deployment of code. Version specific files/configuration makes code deployments and rollbacks simpler by applying the correct configuration to the correct deployment without having to track configuration for each version manually.

Files can be versioned and retrieved using simple commands. Versions can also be coupled with tags allowing versions to be created for tagged files.

Version file: 
`$ cstore push {{file}} -v v0.1.0-beta` or `$ cstore push -t dev -v v0.1.0-beta`

Restore file version: 
`$ cstore pull {{file}} -v v0.1.0-beta` or `$ cstore pull -t dev -v v0.1.0-beta`

List file versions:
`$ cstore list -v`

Note: Files can be retrieved with a specified version. If a versioned file entry is not found in the catalog, cStore will attempt to restore that version of all file entries matching the remaining criteria. This provides the ability to get only versioned files when the catalog aware of the version or to store and retrieve versions without the catalog being aware of the version. This is useful, when a version needs to be pushed and pulled, but the catalog file cannot be updated easily.