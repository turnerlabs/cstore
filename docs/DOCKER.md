## How to Load Configuration in a Docker container running in AWS ##

Managing configuration from the command line is not enough. Applications need a way to pull environment specific configuration in order to run. 

This example uses S3 for the configuration store, but will work with any supported storage solution.

1. Add [docker-entrypoint.sh](../examples/docker-entrypoint-env.sh) script to the repo. 
2. Replace `./app` in the script with the correct application executable. 
```bash
exec ./my-application
```
3. When using secrets injection, add `-i` to the pull command in the script to inject secrets from Secrets Manager.
```bash
cstore pull -le -t $CONFIG_ENV -v $CONFIG_VER -i
```

4. Use the `ENTRYPOINT` command in place of the `CMD` command in Dockerfile to run the shell script. 
```docker
ENTRYPOINT ["./docker-entrypoint.sh"]
```
5. Update the `Dockerfile` to install [cStore](https://github.com/turnerlabs/cstore/releases/download/v3.5.0-alpha/cstore_linux_amd64) for Linux (or the appropriate os) adding execute permissions.
```docker
RUN curl -L -o  /usr/local/bin/cstore https://github.com/turnerlabs/cstore/releases/download/v3.5.0-alpha/cstore_linux_386 && chmod +x /usr/local/bin/cstore
```
6. Update the `docker-compose.yml` file to specify which environment config should be pulled by the `docker-entrypoint.sh` script.    
```bash
    environment:
      CONFIG_ENV: dev
      CONFIG_VER: v1.0.0 # optional
      AWS_REGION: us-east-1
```
7. In the same folder as the `Dockerfile`, use cStore to push the `.env` or `.json` files to an AWS S3 bucket with a `dev` tag. Check the resulting `cstore.yml` file into the repo.
8. Set up the [S3 Bucket](S3.md) policy to allow AWS container role access.

9. Set up the AWS container role policy to allow S3 bucket access.
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
10. Deploy the conainer.