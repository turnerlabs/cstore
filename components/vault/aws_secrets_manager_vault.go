package vault

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/turnerlabs/cstore/components/prompt"
)

// AWSSecretsManagerVault ...
type AWSSecretsManagerVault struct {
	Session *session.Session
}

// Name ...
func (v AWSSecretsManagerVault) Name() string {
	return "aws-secrets-manager"
}

// Description ...
func (v AWSSecretsManagerVault) Description() string {
	return `This vault retrieves secrets from the AWS Secrets Manager. This allows a cstore pull command on an AWS S3 Bucket store to retrieve secret values like database passwords and api keys from AWS Secrets Manager when AWS permission allow.`
}

// Set ...
func (v AWSSecretsManagerVault) Set(contextID, key, value string) error {

	svc := secretsmanager.New(v.Session)
	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(contextID),
		SecretString: aws.String(value),
		Description:  aws.String("cStore"),
	}

	_, err := svc.CreateSecret(input)
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case secretsmanager.ErrCodeResourceNotFoundException:
			return ErrSecretNotFound
		case secretsmanager.ErrCodeResourceExistsException:
			val := prompt.GetValFromUser(fmt.Sprintf("Secret %s exists! Overwrite? (y/N)", contextID),
				"",
				"",
				false)

			if strings.ToLower(val) == "y" || strings.ToLower(val) == "yes" {
				input := &secretsmanager.UpdateSecretInput{
					SecretId:     aws.String(contextID),
					SecretString: aws.String(value),
					Description:  aws.String("cStore"),
				}

				_, err := svc.UpdateSecret(input)
				if _, ok := err.(awserr.Error); ok {
					return err
				}
			}
		default:
			return err
		}
	}

	return nil
}

// Delete ...
func (v AWSSecretsManagerVault) Delete(contextID, key string) error {
	return errors.New("not implemented")
}

// Get ...
func (v AWSSecretsManagerVault) Get(contextID, key, defaultVal, description string, askUser bool) (string, error) {

	svc := secretsmanager.New(v.Session)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(contextID),
		VersionStage: aws.String("AWSCURRENT"),
	}

	output, err := svc.GetSecretValue(input)
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case secretsmanager.ErrCodeResourceNotFoundException:
			return "", ErrSecretNotFound
		default:
			return "", err
		}
	}

	return *output.SecretString, nil
}

func init() {
	v := AWSSecretsManagerVault{}
	vaults[v.Name()] = v
}
