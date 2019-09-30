## How to Load Configuration in a Docker container running in AWS ##

Managing configuration from the command line is not enough. Applications need a way to pull environment specific configuration in order to run correctly. 

1. Add [docker-entrypoint.sh](../examples/docker-entrypoint.sh) script to the repo. 
2. Replace `./my-application` in the script with the correct application executable. 
```bash
exec ./my-application
```
3. Use the `ENTRYPOINT` command in place of the `CMD` command in Dockerfile to run the shell script. 
```docker
ENTRYPOINT ["./docker-entrypoint.sh"]
```
4. Update the `Dockerfile` to install [cStore](https://github.com/turnerlabs/cstore/releases/download/v3.3.1-alpha/cstore_linux_amd64) for Linux (or the appropriate os) adding execute permissions.
```docker
RUN curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.3.1-alpha/cstore_linux_386 && chmod +x /usr/local/bin/cstore
```
5. Update the `docker-compose.yml` file to specify which environment config should be pulled by the `docker-entrypoint.sh` script.    
```bash
    environment:
      CONFIG_ENV: dev
      CONFIG_VER: v1.0.0 # optional
      AWS_REGION: us-east-1
```
6. In the same folder as the `Dockerfile`, use cStore to push the `.env` files to an AWS S3 bucket with a `dev` tag. Check the resulting `cstore.yml` file into the repo.
7. Set up the [S3 Bucket](S3.md) policy to allow AWS container role access.

8. Set up the AWS container role policy to allow S3 bucket access.
```yml
data "aws_iam_policy_document" "app_policy" {
  statement {
    effect = "Allow"

    actions = [
      "s3:Get*",
    ]

    resources = [
      "${var.aws_s3_bucket_arn}/*",
    ]
  }

  # Only required, if injecting secrets from Secrets Manager.
  statement {
    effect = "Allow"

    actions = [
      "secretsmanager:GetSecretValue",
    ]

    resources = [
      "arn:aws:secretsmanager:us-east-1:${var.account_id}:secret:${var.secrets_prefix}/*",
    ]
  }
}
```
9. Deploy the conainer.