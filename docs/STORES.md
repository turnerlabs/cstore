## Stores - Supported Storage Locations ##

| Store | Id | Default Vault | Update Strategy | Version Strategy | Security Level |
|-------|----|---------------|-----------------|------------------|--------|
| [AWS S3 Bucket](S3.md) | `aws-s3` | Secrets Manager | Deploy Time | `$ cstore push -v v1` | Very High |
| [AWS Parameter Store](PARAMETER.md) | `aws-paramter` | Secrets Manager | Deploy Time | `$ cstore push -v v1` | Very High |
| [Source Control](SOURCE_CONTROL.md) | `source-control` | Secrets Manager | Build Time | Source Control | High |