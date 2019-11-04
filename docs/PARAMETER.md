## Using Parameter Store ##

cStore will create a parameter in AWS Parameter Store for each variable in the configuration file.

| CLI Flag | CLI Key | Description | Supports | Parameter Name |
|-|-|-|-|-|
| `-s` |`aws-parameter`| Each config value is stored as a separate parameter. | * |`/{config_context}/{file_path}/{var}`, `/{config_context}/{version}/{file_path}/{var}` |

To authenticate with AWS, use one of the [AWS methods](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html).

### Parameter Restrictions ###

If the file path exceeds AWS Parameter Store's max levels, an error is thrown.

### Encryption ###

With the initial configuration push to Parameter Store, encryption settings are saved. To change these settings, purge and re-push configuration with new encryption settings.

### Version Configuration ###

When pushing version of the configuration file, multiple entries will be created in Parameter Store allowing different versions to be updated or managed independently.

### Updating Configuration ###

When pushing changes, Parameter Store will only be updated when the value or encryption of the parameter has changed.

If a parameter has changed in Parameter Store since the last time the configuration was pulled by cStore, cStore will warn before overwriting the changes.

When a variable is removed from the configuration file and the file is pushed, it will also be removed from parameter store completely without a warning.

### Pulling Task Definition Refs ###

When pulling configuration, use `--store-command=refs` flag to restore the configuration as Parameter Store references that can be added to the secrets section of a Task Definition.

### Enabling AWS Parameter Store Access

To make the parameters accessible using the AWS API, Account Roles, or the CLI command, `$ cstore pull`, use the following resource policy statement.

```json
{
    "Sid": "",
    "Effect": "Allow",
    "Action": [
        "ssm:GetParametersByPath",
        "ssm:GetParameter"
    ],
    "Resource": "arn:aws:ssm:::parameter/*" 
}
```

