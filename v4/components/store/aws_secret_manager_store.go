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
	"github.com/turnerlabs/cstore/v4/components/catalog"
	"github.com/turnerlabs/cstore/v4/components/cfg"
	"github.com/turnerlabs/cstore/v4/components/contract"
	"github.com/turnerlabs/cstore/v4/components/convert"
	"github.com/turnerlabs/cstore/v4/components/display"
	"github.com/turnerlabs/cstore/v4/components/models"
	"github.com/turnerlabs/cstore/v4/components/setting"
	"github.com/turnerlabs/cstore/v4/components/vault"
)

// AWSSecretManagerStore ...
type AWSSecretManagerStore struct {
	Session *session.Session

	clog catalog.Catalog

	uo cfg.UserOptions
	io models.IO
}

// Name ...
func (s AWSSecretManagerStore) Name() string {
	return "aws-secret"
}

// SupportsFeature ...
func (s AWSSecretManagerStore) SupportsFeature(feature string) bool {
	switch feature {
	case VersionFeature:
		return false
	default:
		return false
	}
}

// SupportsFileType ...
func (s AWSSecretManagerStore) SupportsFileType(fileType string) bool {
	switch fileType {
	case EnvFeature, JSONFeature:
		return true
	default:
		return false
	}
}

// Description ...
func (s AWSSecretManagerStore) Description() string {
	return `
	detail: https://github.com/turnerlabs/cstore/v4/blob/master/docs/SECRETS_MANAGER.md
`
}

// Pre ...
func (s *AWSSecretManagerStore) Pre(clog catalog.Catalog, file *catalog.File, access contract.IVault, uo cfg.UserOptions, io models.IO) error {
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
func (s AWSSecretManagerStore) Push(file *catalog.File, fileData []byte, version string) error {

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
	return s.pushBlob(file, fileData, KMSKeyID)
}

func (s AWSSecretManagerStore) pushBlob(file *catalog.File, fileData []byte, KMSKeyID kmsKeyID) error {

	switch file.Type {
	case "json":
		if !json.Valid(fileData) {
			return errors.New("failed to parse JSON")
		}
	case "env":
		result, err := convert.ToJSONObjectFormat(fileData)
		if err != nil {
			return err
		}

		fileData = result.Bytes()
	}

	svc := secretsmanager.New(s.Session)

	key := fmt.Sprintf("%s/%s", s.clog.Context, file.Path)

	sv, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == secretsmanager.ErrCodeInvalidRequestException {
				if _, err = svc.RestoreSecret(&secretsmanager.RestoreSecretInput{
					SecretId: aws.String(key),
				}); err != nil {
					return err
				}

				display.Warn(fmt.Errorf("After %s was marked for deletion, restoration has been requested. Push again to update value.", key), s.io.UserOutput)
				return nil
			} else if aerr.Code() == secretsmanager.ErrCodeResourceNotFoundException {

				if _, err = svc.CreateSecret(&secretsmanager.CreateSecretInput{
					Name:         aws.String(key),
					SecretString: aws.String(string(fileData)),
					Description:  aws.String("cStore"),
					KmsKeyId:     aws.String(KMSKeyID.awsInputValue),
				}); err != nil {
					return err
				}

				return nil
			}
		}

		return err
	}

	sd, err := describeSecret(key, svc)
	if err != nil {
		return err
	}

	if !bytes.Equal([]byte(*sv.SecretString), fileData) || (KMSKeyID.value != "" && sd.keyID != KMSKeyID.value) {

		if _, err := svc.UpdateSecret(&secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(key),
			SecretString: aws.String(string(fileData)),
			Description:  aws.String("cStore"),
			KmsKeyId:     aws.String(KMSKeyID.awsInputValue),
		}); err != nil {
			return err
		}

		return nil
	}

	return err
}

// Pull ...
func (s AWSSecretManagerStore) Pull(file *catalog.File, version string) ([]byte, contract.Attributes, error) {

	svc := secretsmanager.New(s.Session)

	key := fmt.Sprintf("%s/%s", s.clog.Context, file.Path)

	sv, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	})

	if err != nil {
		return []byte{}, contract.Attributes{}, err
	}

	switch file.Type {
	case "json":
		return []byte(*sv.SecretString), contract.Attributes{}, err
	case "env":
		envFormat, err := convert.ToENVFileFormat([]byte(*sv.SecretString))

		return envFormat.Bytes(), contract.Attributes{}, err
	default:
		return []byte(*sv.SecretString), contract.Attributes{}, err
	}
}

// Purge ...
func (s AWSSecretManagerStore) Purge(file *catalog.File, version string) error {

	svc := secretsmanager.New(s.Session)

	key := fmt.Sprintf("%s/%s", s.clog.Context, file.Path)

	if _, err := svc.DeleteSecret(&secretsmanager.DeleteSecretInput{
		SecretId: aws.String(key),
	}); err != nil {
		return err
	}

	return nil
}

// Changed ...
func (s AWSSecretManagerStore) Changed(file *catalog.File, fileData []byte, version string) (time.Time, error) {
	svc := secretsmanager.New(s.Session)

	key := fmt.Sprintf("%s/%s", s.clog.Context, file.Path)

	secret, err := describeSecret(key, svc)
	if err != nil {
		return time.Time{}, err
	}

	return secret.lastModified, nil
}

func init() {
	s := new(AWSSecretManagerStore)
	stores[s.Name()] = s
}
