package vault

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/setting"
	"github.com/turnerlabs/cstore/components/token"
)

const defaultKMSKey = "aws/secretsmanager"

type vaultSettings struct {
	KMSKeyID setting.Setting
}

// AWSSecretsManagerVault ...
type AWSSecretsManagerVault struct {
	Session *session.Session

	settings vaultSettings
	io       models.IO
}

// Name ...
func (v AWSSecretsManagerVault) Name() string {
	return "aws-secrets-manager"
}

// Description ...
func (v AWSSecretsManagerVault) Description() string {
	return `
Secrets are saved and retrieved from AWS Secrets Manager. 

Placing secret tokens in the file {{ENV/KEY::SECRET}} will remove and push secrets into Secrets Manager. 

Using '-i' cli flag during a pull, will inject secrets into a copy of the file created with a '.secrets' extension during the restore.

When saving secrets in Secrets Manager, a KMS Key ID can be provided. Leaving the prompt blank will default to the default Secrets Manager KMS key or a previously specified KMS Key ID. 

In order to access Secrets Manager, applicable Secrets Manager permissions need to be granted along with encrypt and decrypt permissions for the KMS key that Secrets Manager used when storing the secret.
`
}

// BuildKey ...
func (v AWSSecretsManagerVault) BuildKey(contextID, group, prop string) string {
	return fmt.Sprintf("%s/%s", contextID, group)
}

// Pre ...
func (v *AWSSecretsManagerVault) Pre(clog catalog.Catalog, fileEntry *catalog.File, uo cfg.UserOptions, io models.IO) error {
	v.io = io

	v.settings = vaultSettings{
		KMSKeyID: setting.Setting{
			Description:  "KMS Key ID is used by Secrets Manager to encrypt and decrypt secrets. Any role or user accessing a secret must also have access to the KMS key. The aws/secretsmanager is the default Secrets Manager KMS key.",
			Group:        "AWS",
			Prop:         "VAULT_KMS_KEY_ID",
			Prompt:       uo.Prompt,
			Silent:       uo.Silent,
			AutoSave:     false,
			DefaultValue: clog.GetDataByVault(v.Name(), "AWS_VAULT_KMS_KEY_ID", defaultKMSKey),
			Vault:        fileEntry,
		},
	}

	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	v.Session = sess

	return nil
}

// Set ...
func (v AWSSecretsManagerVault) Set(contextID, group, prop, value string) error {

	secretKey := v.BuildKey(contextID, group, prop)

	svc := secretsmanager.New(v.Session)

	KMSKeyID, err := v.settings.KMSKeyID.Get(contextID, v.io)
	if err != nil {
		return err
	}

	storedProps, err := getSecret(secretKey, svc)

	if err != nil {
		if err.Error() == contract.ErrSecretNotFound.Error() {

			b, err := json.Marshal(map[string]string{prop: value})
			if err != nil {
				return err
			}

			input := &secretsmanager.CreateSecretInput{
				Name:         aws.String(v.BuildKey(contextID, group, prop)),
				SecretString: aws.String(string(b)),
				Description:  aws.String("cStore"),
			}

			if KMSKeyID != defaultKMSKey {
				input.KmsKeyId = &KMSKeyID
			}

			if _, err = svc.CreateSecret(input); err != nil {
				return err
			}

			return nil
		}

		return err
	}

	storedProps[prop] = value

	b, err := json.Marshal(storedProps)
	if err != nil {
		return err
	}

	input := &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(v.BuildKey(contextID, group, prop)),
		SecretString: aws.String(string(b)),
		Description:  aws.String("cStore"),
	}

	if KMSKeyID != defaultKMSKey {
		input.KmsKeyId = &KMSKeyID
	}

	if _, err = svc.UpdateSecret(input); err != nil {
		return err
	}

	return nil
}

// Delete ...
func (v AWSSecretsManagerVault) Delete(contextID, group, prop string) error {
	return errors.New("not implemented")
}

// Get ...
func (v AWSSecretsManagerVault) Get(contextID, group, prop string) (string, error) {
	svc := secretsmanager.New(v.Session)

	storedProps, err := getSecret(v.BuildKey(contextID, group, prop), svc)
	if err != nil {
		return "", err
	}

	if value, found := storedProps[prop]; found {
		return value, nil
	}

	return "", contract.ErrSecretNotFound
}

func getSecret(key string, svc *secretsmanager.SecretsManager) (map[string]string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(key),
		VersionStage: aws.String("AWSCURRENT"),
	}

	output, err := svc.GetSecretValue(input)
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case secretsmanager.ErrCodeResourceNotFoundException:
			return map[string]string{}, contract.ErrSecretNotFound
		default:
			return map[string]string{}, err
		}
	}

	storedProps := map[string]string{}
	err = json.Unmarshal([]byte(*output.SecretString), &storedProps)
	if err != nil {
		return map[string]string{}, err
	}

	return storedProps, nil
}

func extractSecrets(tokens map[string]token.Token) map[string]string {

	secrets := map[string]string{}

	for _, v := range tokens {
		secrets[v.Secret()] = ""
	}

	return secrets
}

func init() {
	v := AWSSecretsManagerVault{}
	vaults[v.Name()] = &v
}
