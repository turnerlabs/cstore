## How to Load Configuration in a Lambda function running in AWS ##

### Node.js Example ###

Managing configuration from the command line is not enough. Functions need a way to get environment specific configuration in order to execute. 

This example uses S3 for the configuration store, but will work with any supported storage solution.

1. Place the following files in the lambda function folder. 
  - [cStore](https://github.com/turnerlabs/cstore/releases/download/v3.3.1-alpha/cstore_linux_amd64) (needs execute permissions)
  - [cstore.js](../examples/cstore.js)

2. Create a configuration file `dev.json` in the lambda function folder.
```json
{
    "user": "user",
    "password": "************"
}
```
3. In the lambda function folder, use a local install of cStore to push the `dev.json` file to an AWS S3 bucket with a `dev` tag. The resulting `cstore.yml` file should be checked into the repo but not the `dev.json` file as it may contain secrets.

4. Add this line of code to the lambda function handler file to load configuration.
```javascript
var config = cstore.pull('cstore_linux_amd64', process.env.ENVIRONMENT)
```

5. Update the terraform lambda function environment variables to specify which environment config file should be retrieved when the lambda function executes.    
```yml
    resource "aws_lambda_function" "lambda" {
      function_name = "${var.app}-${var.environment}-ci-auto-rotate"

      filename         = "${data.archive_file.lambda_zip.output_path}"
      source_code_hash = "${data.archive_file.lambda_zip.output_base64sha256}"

      handler = "handler.handler"
      runtime = "nodejs8.10"

      timeout = 10

      role = "${aws_iam_role.lambda_exec.arn}"

      tags = "${var.tags}"

      environment {
        variables = {
          ENVIRONMENT = "${var.environment}"
        }
      }
    }
```
6. Set up the [S3 Bucket](S3.md) policy to allow access for the AWS lambda function's role.

7. Set up the AWS lambda role policy to allow S3 bucket access.
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
8. Deploy the lambda function.