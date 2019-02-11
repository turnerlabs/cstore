## Set Up Parameter Store ##

Parameter Store requires no infrasructure set up.

cStore will create a parameter in AWS Parameter Store for each variable in the configuration file.

To authenticate with AWS, use one of the [AWS methods](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html).

### Parameter Key Formatting ###

In Parameter Store, each value's key will be generated using one of the following formats. 
- `/{CSTORE_CONTEXT}/{FILE_PATH}/{VAR}` (default)
- `/{CONTEXT}/{VERSION}/{FILE_PATH}/{VAR}` (versioned)

If the file path exceeds AWS Parameter Store's max levels, an error is thrown.

### Versioning Configuration ###

When pushing version of the configuration file, multiple entries will be created in Parameter Store allowing different versions to be updated or managed independently.

### Pushing Configuration Changes ###

When pushing changes, Parameter Store will only be updated when the value of the parameter has changed.

If a parameter has changed in Parameter Store since the last time the configuration was pulled by cStore, cStore will warn before overwriting the changes.

When a variable is removed from the configuration file and the file is pushed, it will also be removed from parameter store completely without a warning.

### Pulling Task Definition Refs ###

When pulling configuration, use `--store-command=refs` flag to restore the configuration as Parameter Store references that can be added to the secrets section of a Task Definition.

### Encryption ###

With the initial configuration push to Parameter Store, encryption settings are saved. To change these settings, purge and re-push configuration with new encryption settings.
