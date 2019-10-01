package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/subosito/gotenv"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/display"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/prompt"
	"github.com/turnerlabs/cstore/components/setting"
	"github.com/turnerlabs/cstore/components/vault"
)

const defaultSMKMSKey = "aws/secretsmanager"

// AWSSecretsManagerStore ...
type AWSSecretsManagerStore struct {
	Session *session.Session

	context  string
	settings map[string]setting.Setting

	encryptionType string
	credentialType string

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
	case EnvFeature:
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
	s.settings = map[string]setting.Setting{}
	s.context = clog.Context
	s.uo = uo
	s.io = io

	s.credentialType = autoDetect
	s.encryptionType = getEncryptionType(*file)

	(setting.Setting{
		Group:        "AWS",
		Prop:         "REGION",
		Prompt:       uo.Prompt,
		Silent:       uo.Silent,
		AutoSave:     true,
		DefaultValue: awsDefaultRegion,
		Vault:        vault.EnvVault{},
	}).Get(clog.Context, io)

	//------------------------------------------
	//- Auth Credentials
	//------------------------------------------
	if uo.Prompt {
		s.credentialType = strings.ToLower(prompt.GetValFromUser("Authentication", prompt.Options{
			Description:  "OPTIONS\n (P)rofile \n (U)ser",
			DefaultValue: "P"}, io))
	}

	switch s.credentialType {
	case cTypeProfile:
		os.Unsetenv(awsSecretAccessKey)
		os.Unsetenv(awsAccessKeyID)

		(setting.Setting{
			Group:        "AWS",
			Prop:         "PROFILE",
			DefaultValue: os.Getenv(awsProfile),
			Prompt:       uo.Prompt,
			Silent:       uo.Silent,
			AutoSave:     true,
			Vault:        vault.EnvVault{},
		}).Get(clog.Context, io)

	case cTypeUser:
		os.Unsetenv(awsProfile)

		(setting.Setting{
			Group:    "AWS",
			Prop:     "ACCESS_KEY_ID",
			Prompt:   uo.Prompt,
			Silent:   uo.Silent,
			AutoSave: true,
			Vault:    access,
		}).Get(clog.Context, io)

		(setting.Setting{
			Group:    "AWS",
			Prop:     "SECRET_ACCESS_KEY",
			Prompt:   uo.Prompt,
			Silent:   uo.Silent,
			AutoSave: true,
			Vault:    access,
		}).Get(clog.Context, io)
	}

	//------------------------------------------
	//- Encryption
	//------------------------------------------
	s.settings[serverEncryptionToken] = setting.Setting{
		Description:  "KMS Key ID is used by Parameter Store to encrypt and decrypt secrets. Any role or user accessing a secret must also have access to the KMS key. When pushing updates, the default setting will preserve existing KMS keys. The aws/ssm key is the default Systems Manager KMS key.",
		Group:        "AWS",
		Prop:         "STORE_KMS_KEY_ID",
		DefaultValue: clog.GetDataByStore(s.Name(), "AWS_STORE_KMS_KEY_ID", defaultSMKMSKey),
		Prompt:       uo.Prompt,
		Silent:       uo.Silent,
		AutoSave:     false,
		Vault:        file,
	}

	//------------------------------------------
	//- Open Connection
	//------------------------------------------
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	s.Session = sess

	return err
}

// Push ...
func (s AWSSecretsManagerStore) Push(file *catalog.File, fileData []byte, version string) error {

	if !file.SupportsConfig() {
		return fmt.Errorf("store does not support file type: %s", file.Type)
	}

	if len(fileData) == 0 {
		return errors.New("empty file")
	}

	//------------------------------------------
	//- Get encryption keys
	//------------------------------------------
	KMSKeyID := ""

	key, serverEncryption := s.settings[serverEncryptionToken]
	if serverEncryption {
		value, err := key.Get(s.context, s.io)
		if err != nil {
			return err
		}

		if value != defaultSMKMSKey {
			KMSKeyID = value
		}
	}

	//------------------------------------------
	//- Push configuration
	//------------------------------------------
	params := gotenv.Parse(bytes.NewReader(fileData))
	if len(params) == 0 {
		return errors.New("failed to parse environment variables")
	}

	svc := secretsmanager.New(s.Session)

	//------------------------------------------
	//- Delete removed params
	//------------------------------------------
	for name, dataType := range file.Data {
		if dataType != "SECRET" {
			continue
		}

		key := formatSecretToken(s.context, file.Path, name)

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

		key := formatSecretToken(s.context, file.Path, name)

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

					display.Warn(fmt.Errorf("After %s was marked for deletion, restoration requested. Push again to update value.", key), s.io.UserOutput)
					continue
				}
			}

			if err.Error() == contract.ErrSecretNotFound.Error() {

				b, err := json.Marshal(map[string]string{name: value})
				if err != nil {
					return err
				}

				input := &secretsmanager.CreateSecretInput{
					Name:         aws.String(key),
					SecretString: aws.String(string(b)),
					Description:  aws.String("cStore"),
					KmsKeyId:     aws.String(KMSKeyID),
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

		secret, err := describeSecret(key, svc)
		secret.values = storedProps

		fmt.Println(hasSecretChanged(secret, name, value, file.Data))

		if hasSecretChanged(secret, name, value, file.Data) {

			b, err := json.Marshal(map[string]string{name: value})
			if err != nil {
				return err
			}

			input := &secretsmanager.UpdateSecretInput{
				SecretId:     aws.String(key),
				SecretString: aws.String(string(b)),
				Description:  aws.String("cStore"),
				KmsKeyId:     aws.String(KMSKeyID),
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

func hasSecretChanged(existing secret, name, value string, data map[string]string) bool {

	v, found := existing.values[name]

	if !found {
		return true
	}

	keyID, _ := data["AWS_STORE_KMS_KEY_ID"]

	return v != value || (keyID != "" && existing.keyID != keyID)
}

func getSecrets(context, path string, data map[string]string, svc *secretsmanager.SecretsManager) (map[string]string, error) {
	secrets := map[string]string{}

	for secret, dataType := range data {
		if dataType != "SECRET" {
			continue
		}

		path := formatSecretToken(context, path, secret)

		pairs, err := getSecret(path, svc)
		if err != nil {
			return secrets, err
		}

		for k, v := range pairs {
			secrets[k] = v
		}
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

// Pull ...
func (s AWSSecretsManagerStore) Pull(file *catalog.File, version string) ([]byte, contract.Attributes, error) {

	svc := secretsmanager.New(s.Session)

	storedSecrets, err := getSecrets(s.context, file.Path, file.Data, svc)
	if err != nil {
		return []byte{}, contract.Attributes{}, err
	}

	if len(storedSecrets) == 0 {
		return []byte{}, contract.Attributes{}, errors.New("secrets not found, verify AWS account and credentials")
	}

	var buffer bytes.Buffer

	for key, value := range storedSecrets {
		buffer.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	return buffer.Bytes(), contract.Attributes{}, nil
}

// Purge ...
func (s AWSSecretsManagerStore) Purge(file *catalog.File, version string) error {

	svc := secretsmanager.New(s.Session)

	for name, dataType := range file.Data {
		if dataType != "SECRET" {
			continue
		}

		key := formatSecretToken(s.context, file.Path, name)

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

		secret, err := describeSecret(formatSecretToken(s.context, file.Path, name), svc)
		if err != nil {
			return time.Time{}, err
		}

		storedSecretMetaData = append(storedSecretMetaData, secret)
	}

	return lastModifiedSecret(storedSecretMetaData), nil
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

type secret struct {
	name   string
	values map[string]string

	keyID        string
	lastModified time.Time
}

func init() {
	s := new(AWSSecretsManagerStore)
	stores[s.Name()] = s
}
