## Stores - Supported Storage Locations ##

* [AWS S3 Bucket](S3.md) (aws-s3)
* [AWS Parameter Store](PARAMETER.md) (aws-parameter)

### Configuration ###

To configure a store's credentials or encryption settings use `-p` on the commandline and follow the prompts. Options specified by flags during a `push` command will be saved under the catalog's file entry and options specified by flags used during a `pull` will override a catalog's file entry settings.