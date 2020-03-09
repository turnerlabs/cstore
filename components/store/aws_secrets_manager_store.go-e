package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/subosito/gotenv"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/display"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/setting"
	"github.com/turnerlabs/cstore/components/vault"
)

// AWSSecretsManagerStore ...
type AWSSecretsManagerStore struct {
	Session *session.Session

	clog catalog.Catalog

	uo cfg.UserOptions
	io models.IO
}

// Name ...
func (s AWSSecretsManagerStore) Name() string {
	return "aws-secrets"
}

// SupportsFeature ...
func (s AWSSecretsManagerStore) SupportsFeature(feature string) bool {
	switch feature {
	case VersionFeature:
		return false
	default:
		return false
	}
}

// SupportsFileType ...
func (s AWSSecretsManagerStore) SupportsFileType(fileType string) bool {
	switch fileType {
	case EnvFeature, JSONFeature:
		return true
	default:
		return false
	}
}

// Description ...
func (s AWSSecretsManagerStore) Description() string {
	return `
	detail: https://github.com/turnerlabs/cstore/blob/master/docs/SECRETS_MANAGER.md
`
}

// Pre ...
func (s *AWSSecretsManagerStore) Pre(clog catalog.Catalog, file *catalog.File, access contract.IVault, uo cfg.UserOptions, io models.IO) error {
	s.clog = clog
	s.uo = uo
	s.io = io

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
		Vault:        vault.EnvVault{},
	}.Get(clog.Context, io)

	//------------------------------------------
	//- Get AWS Credentials from Environment
	//------------------------------------------
	if _, ok := access.(vault.EnvVault); ok {
		s.Session, err = session.NewSession(&aws.Config{
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

	s.Session, err = session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(id, secret, token),
	})

	return err
}

// Push ...
func (s AWSSecretsManagerStore) Push(file *catalog.File, fileData []byte, version string) error {

	if len(fileData) == 0 {
		return errors.New("empty file")
	}

	//------------------------------------------
	//- Get encryption key
	//------------------------------------------
	value, err := setting.Setting{
		Description:  "KMS Key ID is used by Secrets Manager to encrypt and decrypt secrets. Any role or user accessing a secret must also have access to the KMS key. When pushing updates, the default setting will preserve existing KMS keys. The aws/ssm key is the default Systems Manager KMS key.",
		Prop:         awsStoreKMSKeyID,
		DefaultValue: s.clog.GetDataByStore(s.Name(), awsStoreKMSKeyID, defaultSMKMSKey),
		Prompt:       s.uo.Prompt,
		Silent:       s.uo.Silent,
		AutoSave:     false,
		Vault:        file,
	}.Get(s.clog.Context, s.io)
	if err != nil {
		return err
	}

	KMSKeyID := kmsKeyID{
		value:         value,
		awsInputValue: value,
	}

	if value == defaultSMKMSKey {
		KMSKeyID.awsInputValue = ""
	}

	//------------------------------------------
	//- Push configuration
	//------------------------------------------
	return s.pushFile(file, fileData, KMSKeyID)
}

func (s AWSSecretsManagerStore) pushFile(file *catalog.File, fileData []byte, KMSKeyID kmsKeyID) error {
	params := map[string]string{}

	switch file.Type {
	case "json":
		temp := map[string]json.RawMessage{}

		if err := json.Unmarshal(fileData, &temp); err != nil {
			return err
		}

		for k, v := range temp {
			var str string

			if err := json.Unmarshal(v, &str); err != nil {

				b, err := json.MarshalIndent(v, "", "   ")
				if err != nil {
					return err
				}

				params[k] = string(b)
			} else {
				params[k] = fmt.Sprintf(`{"%s":"%s"}`, k, str)
			}
		}

	case "env":
		envParams := gotenv.Parse(bytes.NewReader(fileData))
		if len(envParams) == 0 {
			return errors.New("failed to parse environment variables")
		}

		for k, v := range envParams {
			params[k] = fmt.Sprintf(`{"%s":"%s"}`, k, v)
		}
	default:
		return fmt.Errorf("store does not support file type: %s", file.Type)
	}

	svc := secretsmanager.New(s.Session)

	//------------------------------------------
	//- Delete removed params
	//------------------------------------------
	for name, dataType := range file.Data {
		if dataType != "SECRET" {
			continue
		}

		key := formatSecretToken(s.clog.Context, file.Path, name)

		removed := true
		for k := range params {
			if k == name {
				removed = false
			}
		}

		if removed {
			if _, err := svc.DeleteSecret(&secretsmanager.DeleteSecretInput{
				SecretId: aws.String(key),
			}); err != nil {
				fmt.Fprintf(s.io.UserOutput, "secret: %s", key)
				return err
			}

			delete(file.Data, name)
		}
	}

	for name, value := range params {
		newSecret := secret{
			name:  name,
			value: value,
			keyID: KMSKeyID.value,
		}

		key := formatSecretToken(s.clog.Context, file.Path, name)

		storedProps, err := getSecret(key, svc)

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == secretsmanager.ErrCodeInvalidRequestException {
					_, err = svc.RestoreSecret(&secretsmanager.RestoreSecretInput{
						SecretId: aws.String(key),
					})
					if err != nil {
						return err
					}

					display.Warn(fmt.Errorf("Requested restoration for deleted secret %s. A second push will update the secret.", key), s.io.UserOutput)

					file.AddData(map[string]string{
						name: "SECRET",
					})

					continue
				}
			}

			if err.Error() == contract.ErrSecretNotFound.Error() {

				input := &secretsmanager.CreateSecretInput{
					Name:         aws.String(key),
					SecretString: aws.String(value),
					Description:  aws.String("cStore"),
					KmsKeyId:     aws.String(KMSKeyID.awsInputValue),
				}

				if _, err = svc.CreateSecret(input); err != nil {
					return err
				}

				file.AddData(map[string]string{
					name: "SECRET",
				})

				continue
			}

			return err
		}

		storedSecret, err := describeSecret(key, svc)
		if err != nil {
			return err
		}

		storedSecret.value = storedProps

		if hasSecretChanged(storedSecret, newSecret) {

			input := &secretsmanager.UpdateSecretInput{
				SecretId:     aws.String(key),
				SecretString: aws.String(value),
				Description:  aws.String("cStore"),
				KmsKeyId:     aws.String(KMSKeyID.awsInputValue),
			}

			if _, err = svc.UpdateSecret(input); err != nil {
				return err
			}
		}

		file.AddData(map[string]string{
			name: "SECRET",
		})
	}

	return nil
}

func hasSecretChanged(existing secret, new secret) bool {

	if new.keyID != "" && existing.keyID != new.keyID {
		return true
	}

	if existing.value != new.value {
		return true
	}

	return false
}

func getSecrets(context, path string, data map[string]string, svc *secretsmanager.SecretsManager) (map[string]string, error) {
	secrets := map[string]string{}

	secretReferences := 0

	for secret, dataType := range data {
		if dataType != "SECRET" {
			continue
		}

		secretReferences++

		path := formatSecretToken(context, path, secret)

		value, err := getSecret(path, svc)
		if err != nil {
			return secrets, err
		}

		secrets[secret] = value
	}

	if secretReferences == 0 {
		return secrets, errors.New("SecretReferencesNotFound: catalog data missing secret references")
	}

	return secrets, nil
}

func describeSecret(key string, svc *secretsmanager.SecretsManager) (secret, error) {
	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(key),
	}

	o, err := svc.DescribeSecret(input)
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case secretsmanager.ErrCodeResourceNotFoundException:
			return secret{}, contract.ErrSecretNotFound
		default:
			return secret{}, err
		}
	}

	if o.DeletedDate != nil {
		return secret{}, contract.ErrSecretNotFound
	}

	s := secret{
		name:         *o.Name,
		lastModified: *o.LastChangedDate,
		keyID:        defaultSMKMSKey,
	}

	if o.KmsKeyId != nil {
		s.keyID = *o.KmsKeyId
	}

	return s, nil
}

func getSecret(key string, svc *secretsmanager.SecretsManager) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(key),
		VersionStage: aws.String("AWSCURRENT"),
	}

	output, err := svc.GetSecretValue(input)
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case secretsmanager.ErrCodeResourceNotFoundException:
			return "", contract.ErrSecretNotFound
		default:
			return "", err
		}
	}

	return *output.SecretString, nil
}

// Pull ...
func (s AWSSecretsManagerStore) Pull(file *catalog.File, version string) ([]byte, contract.Attributes, error) {

	svc := secretsmanager.New(s.Session)

	storedSecrets, err := getSecrets(s.clog.Context, file.Path, file.Data, svc)
	if err != nil {
		return []byte{}, contract.Attributes{}, err
	}

	if len(storedSecrets) == 0 {
		return []byte{}, contract.Attributes{}, errors.New("SecretsNotFound: verify correct AWS account and credentials")
	}

	switch file.Type {
	case "json":
		props := map[string]json.RawMessage{}

		for key, value := range storedSecrets {
			temp := map[string]json.RawMessage{}

			if err := json.Unmarshal([]byte(value), &temp); err == nil {
				if tempValue, found := temp[key]; found {
					value = string(tempValue)
				}
			}

			props[key] = []byte(value)
		}

		b, err := json.MarshalIndent(props, "", "   ")
		if err != nil {
			return []byte{}, contract.Attributes{}, err
		}

		return b, contract.Attributes{}, nil
	case "env":
		buffer := bytes.Buffer{}
		for key, value := range storedSecrets {
			temp := map[string]string{}

			if err := json.Unmarshal([]byte(value), &temp); err != nil {
				return []byte{}, contract.Attributes{}, err
			}

			buffer.WriteString(fmt.Sprintf("%s=%s\n", key, temp[key]))
		}

		return buffer.Bytes(), contract.Attributes{}, nil
	default:
		return []byte{}, contract.Attributes{}, fmt.Errorf("store does not support file type: %s", file.Type)
	}
}

// Purge ...
func (s AWSSecretsManagerStore) Purge(file *catalog.File, version string) error {

	svc := secretsmanager.New(s.Session)

	for name, dataType := range file.Data {
		if dataType != "SECRET" {
			continue
		}

		key := formatSecretToken(s.clog.Context, file.Path, name)

		if _, err := svc.DeleteSecret(&secretsmanager.DeleteSecretInput{
			SecretId: aws.String(key),
		}); err != nil {
			return err
		}

		delete(file.Data, name)
	}

	return nil
}

// Changed ...
func (s AWSSecretsManagerStore) Changed(file *catalog.File, fileData []byte, version string) (time.Time, error) {
	svc := secretsmanager.New(s.Session)

	storedSecretMetaData := []secret{}
	for name, value := range file.Data {
		if value != "SECRET" {
			continue
		}

		secret, err := describeSecret(formatSecretToken(s.clog.Context, file.Path, name), svc)
		if err != nil {
			if err.Error() == contract.ErrSecretNotFound.Error() {
				return time.Time{}, nil
			}

			return time.Time{}, err
		}

		storedSecretMetaData = append(storedSecretMetaData, secret)
	}

	return lastModifiedSecret(storedSecretMetaData), nil
}

func listSecrets(svc *secretsmanager.SecretsManager, startsWith string, nextToken string, secrets []*secretsmanager.SecretListEntry) ([]*secretsmanager.SecretListEntry, error) {

	input := &secretsmanager.ListSecretsInput{}

	if len(nextToken) > 0 {
		input.SetNextToken(nextToken)
	}

	output, err := svc.ListSecrets(input)
	if err != nil {
		return nil, err
	}

	secrets = append(secrets, output.SecretList...)

	if output.NextToken == nil || len(*output.NextToken) == 0 {
		return secrets, nil
	}

	return listSecrets(svc, startsWith, *output.NextToken, secrets)
}

func lastModifiedSecret(params []secret) time.Time {
	mostRecentlyModified := time.Time{}
	for _, sp := range params {
		if mostRecentlyModified.Before(sp.lastModified) {
			mostRecentlyModified = sp.lastModified
		}
	}

	return mostRecentlyModified
}

func init() {
	s := new(AWSSecretsManagerStore)
	stores[s.Name()] = s
}
