## Store Integration Tests ##

To run store integration tests locally, configure store authentication and settings.

### AWS S3 Bucket ###

Internet connectivity and access to an AWS S3 Bucket is required for these tests.

```bash
$ export AWS_REGION={{REGION}} // us-east-1
$ export AWS_PROFILE={{PROFILE}}
$ export AWS_S3_BUCKET={{BUCKET_NAME}}
$ export AWS_STORE_KMS_KEY_ID={{KEY_ID}}
$ go test ./cmd/tests/s3
```