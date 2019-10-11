package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/setting"
)

// AWSSecretManagerVault ...
type AWSSecretManagerVault struct {
	Session *session.Session

	clog      catalog.Catalog
	fileEntry *catalog.File

	uo cfg.UserOptions
	io models.IO
}

// Name ...
func (v AWSSecretManagerVault) Name() string {
	return "aws-secret-manager"
}

// Description ...
func (v AWSSecretManagerVault) Description() string {
	return `
Secrets are saved and retrieved from AWS Secrets Manager. 

Placing secret tokens in the file {{ENV/KEY::SECRET}} will remove and push secrets into Secrets Manager. 

Using '-i' cli flag during a pull, will inject secrets into a copy of the file created with a '.secrets' extension during the restore.

When saving secrets in Secrets Manager, a KMS Key ID can be provided. Leaving the prompt blank will default to the default Secrets Manager KMS key or a previously specified KMS Key ID. 

In order to access Secrets Manager, applicable Secrets Manager permissions need to be granted along with encrypt and decrypt permissions for the KMS key that Secrets Manager used when storing the secret.
`
}

// BuildKey ...
func (v AWSSecretManagerVault) BuildKey(contextID, group, prop string) string {
	return fmt.Sprintf("%s/%s", contextID, strings.Split(strings.ToLower(group), "/")[0])
}

// Pre ...
func (v *AWSSecretManagerVault) Pre(clog catalog.Catalog, fileEntry *catalog.File, access contract.IVault, uo cfg.UserOptions, io models.IO) error {
	v.uo = uo
	v.io = io

	v.fileEntry = fileEntry

	//------------------------------------------
	//- Get AWS Region
	//------------------------------------------
	region, err := setting.Setting{
		Description:  "Export as an environment variable to silence this prompt.",
		Group:        clog.Context,
		Prop:         awsRegion,
		Prompt:       uo.Prompt,
		Silent:       uo.Silent,
		AutoSave:     true,
		PromptOnce:   true,
		DefaultValue: awsDefaultRegion,
		Vault:        EnvVault{},
	}.Get(clog.Context, io)

	//------------------------------------------
	//- Get AWS Credentials from Environment
	//------------------------------------------
	if _, ok := access.(EnvVault); ok {
		v.Session, err = session.NewSession(&aws.Config{
			Region: aws.String(region),
		})

		return err
	}

	//------------------------------------------
	//- Get AWS Credentials from Vault
	//------------------------------------------
	id, err := setting.Setting{
		Description: fmt.Sprintf("Save credential in %s.", access.Name()),
		Group:       clog.Context,
		Prop:        awsAccessKeyID,
		Prompt:      uo.Prompt,
		Silent:      uo.Silent,
		AutoSave:    true,
		PromptOnce:  true,
		Vault:       access,
	}.Get(clog.Context, io)
	if err != nil {
		return err
	}

	secret, err := setting.Setting{
		Description: fmt.Sprintf("Save credential in %s.", access.Name()),
		Group:       clog.Context,
		Prop:        awsSecretAccessKey,
		Prompt:      uo.Prompt,
		Silent:      uo.Silent,
		AutoSave:    true,
		PromptOnce:  true,
		Vault:       access,
	}.Get(clog.Context, io)
	if err != nil {
		return err
	}

	token, err := setting.Setting{
		Description: fmt.Sprintf("Save credential in %s.", access.Name()),
		Group:       clog.Context,
		Prop:        awsSessionToken,
		Prompt:      uo.Prompt,
		Silent:      uo.Silent,
		AutoSave:    true,
		PromptOnce:  true,
		Vault:       access,
	}.Get(clog.Context, io)
	if err != nil {
		return err
	}

	v.Session, err = session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(id, secret, token),
	})

	return err
}

// Set ...
func (v AWSSecretManagerVault) Set(contextID, group, prop, value string) error {

	secretKey := v.BuildKey(contextID, group, prop)

	secretProp := buildProp(group, prop)

	svc := secretsmanager.New(v.Session)

	KMSKeyID, err := setting.Setting{
		Description:  "KMS Key ID is used by Secrets Manager to encrypt and decrypt secrets. Any role or user accessing a secret must also have access to the KMS key. The aws/secretsmanager is the default Secrets Manager KMS key.",
		Prop:         awsVaultKMSKeyID,
		Prompt:       v.uo.Prompt,
		Silent:       v.uo.Silent,
		AutoSave:     false,
		DefaultValue: v.clog.GetDataByVault(v.Name(), awsVaultKMSKeyID, defaultKMSKey),
		Vault:        v.fileEntry,
	}.Get(contextID, v.io)
	if err != nil {
		return err
	}

	storedProps, err := getSecret(secretKey, svc)

	if err != nil {
		if err.Error() == contract.ErrSecretNotFound.Error() {

			b, err := json.Marshal(map[string]string{secretProp: value})
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

	storedProps[secretProp] = value

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
func (v AWSSecretManagerVault) Delete(contextID, group, prop string) error {
	return errors.New("not implemented")
}

// Get ...
func (v AWSSecretManagerVault) Get(contextID, group, prop string) (string, error) {
	svc := secretsmanager.New(v.Session)

	storedProps, err := getSecret(v.BuildKey(contextID, group, prop), svc)
	if err != nil {
		return "", err
	}

	if value, found := storedProps[buildProp(group, prop)]; found {
		return value, nil
	}

	return "", contract.ErrSecretNotFound
}

func buildProp(group, prop string) string {

	dirs := append(strings.Split(group, "/"), prop)

	return strings.Join(dirs[1:len(dirs)], "/")
}

func init() {
	v := AWSSecretManagerVault{}
	vaults[v.Name()] = &v
}
