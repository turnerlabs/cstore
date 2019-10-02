## Stores - Supported Storage Solutions ##

| Store | Key | Supported File Types | Default Secrets Vault | Config Update Strategy | Required Infrastructure | Setup Complexity |
|-----|-----|-----|-----|-----|-----|-----|
| [AWS S3 Bucket](S3.md) | `aws-s3` | * | Secrets Manager | Deploy Time | S3 Bucket, KMS Key| Moderate |
| [AWS Parameter Store](PARAMETER.md) | `aws-parameter` | `.env` | Secrets Manager | Deploy Time | KMS Key | Low |
| [Source Control](SOURCE_CONTROL.md) | `source-control` | `.env` `.json` | Secrets Manager | Build Time | KMS Key | Low |
