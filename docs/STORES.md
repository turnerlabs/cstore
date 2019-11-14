## Storage Solutions ##

A comparison of supported storage solutions.

| | [Source Control](SOURCE_CONTROL.md) | [AWS S3 Bucket](S3.md) | [AWS Parameter Store](PARAMETER.md) | [AWS Secrets Manager](SECRETS_MANAGER.md) | 
|-|-|-|-|-|
| CLI Flag | `-s` | `-s` | `-s` | `-s` | `-s` |
| CLI Key | `source-control` | `aws-s3`  | `aws-parameter` | `aws-secret` `aws-secrets` |
| Supported File Types | `.env`, `.json` | * | `.env` | * |
| Default Secrets Vault | Secrets Manager | Secrets Manager | Secrets Manager | Secrets Manager |
| Config Update Strategy | Build Time | Deploy Time | Deploy Time | Deploy Time |
| Infrastructure | KMS Key | S3 Bucket, KMS Key | KMS Key | KMS Key |
| Setup Complexity | Lower | Moderate | Lower | Lower |
| Cost | Lower | Lower | Moderate | Higher |
| Management GUI | No | No | Yes | Yes |
| Service Limits | [Details](https://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html) | [Details](https://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html)| [Details](https://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html) |  [Details](https://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html) |

