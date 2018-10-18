package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cipher"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/prompt"
	"github.com/turnerlabs/cstore/components/setting"
	"github.com/turnerlabs/cstore/components/vault"
)

const (
	awsBucketName         = "AWS_S3_BUCKET"
	clientEncryptionToken = "CLIENT_ENCRYPTION"
	serverEncryptionToken = "SERVER_ENCRYPTION"

	fileDataEncryptionKey = "ENCRYPTION"

	sep = "::"

	eTypeClient = "c"
	eTypeServer = "s"

	cTypeProfile = "p"
	cTypeUser    = "u"

	autoDetect = "d"
	none       = "n"
)

// S3Store ...
type S3Store struct {
	Session *session.Session

	context  string
	settings map[string]setting.Setting

	encryptionType string
	credentialType string

	io models.IO
}

// Name ...
func (s S3Store) Name() string {
	return "aws-s3"
}

// Supports ...
func (s S3Store) Supports(feature string) bool {
	switch feature {
	case VersionFeature:
		return true
	default:
		return false
	}
}

// Description ...
func (s S3Store) Description() string {
	const description = `
The files are stored in AWS S3 using a similar path to their current location with the exception of a context folder prefix and a version folder when the file is versioned.

To authenticate with AWS S3, credentials can be set in multiple ways.

1. Use '-p' cli flag during a push to be prompted for auth settings.
2. Set %s and %s environment variables.
3. Set %s environment variable to a profile specified in the '~/.aws/credentials' file.

If using an AWS KMS key on the S3 bucket, users will also need KMS key encrypt and decript permissions
`

	return fmt.Sprintf(description, awsAccessKeyID, awsSecretAccessKey, awsProfile)
}

// Pre ...
func (s *S3Store) Pre(clog catalog.Catalog, file *catalog.File, access contract.IVault, promptUser bool, io models.IO) error {
	s.settings = map[string]setting.Setting{}
	s.context = clog.Context
	s.io = io

	s.credentialType = autoDetect
	s.encryptionType = getEncryptionType(*file)

	(setting.Setting{
		Group:        "AWS",
		Prop:         "REGION",
		Prompt:       promptUser,
		Set:          true,
		DefaultValue: awsDefaultRegion,
		Vault:        vault.EnvVault{},
	}).Get(clog.Context, io)

	//---------------------------------------------
	//- Store authentication and encryption options
	//---------------------------------------------
	if promptUser {
		s.credentialType = strings.ToLower(prompt.GetValFromUser("Authentication", prompt.Options{
			Description:  "OPTIONS\n (P)rofile \n (U)ser",
			DefaultValue: "P"}, io))
	}

	//------------------------------------------
	//- Required auth creds
	//------------------------------------------
	switch s.credentialType {
	case cTypeProfile:
		os.Unsetenv(awsSecretAccessKey)
		os.Unsetenv(awsAccessKeyID)

		(setting.Setting{
			Group:        "AWS",
			Prop:         "PROFILE",
			DefaultValue: os.Getenv(awsProfile),
			Prompt:       promptUser,
			Set:          true,
			Vault:        vault.EnvVault{},
		}).Get(clog.Context, io)

	case cTypeUser:
		os.Unsetenv(awsProfile)

		(setting.Setting{
			Group:  "AWS",
			Prop:   "ACCESS_KEY_ID",
			Prompt: promptUser,
			Set:    true,
			Vault:  access,
			Stage:  vault.EnvVault{},
		}).Get(clog.Context, io)

		(setting.Setting{
			Group:  "AWS",
			Prop:   "SECRET_ACCESS_KEY",
			Prompt: promptUser,
			Set:    true,
			Vault:  access,
			Stage:  vault.EnvVault{},
		}).Get(clog.Context, io)
	}

	//------------------------------------------
	//- Store Configuration
	//------------------------------------------
	s.settings[awsBucketName] = setting.Setting{
		Description:  "Specify the S3 Bucket that will store the file.",
		Group:        "AWS",
		Prop:         "S3_BUCKET",
		Prompt:       promptUser,
		Set:          true,
		DefaultValue: clog.GetAnyDataBy(awsBucketName, fmt.Sprintf("cstore-%s", clog.Context)),
		Vault:        file,
	}

	//------------------------------------------
	//- Optional encryption
	//------------------------------------------
	if promptUser {
		s.encryptionType = strings.ToLower(prompt.GetValFromUser("Encryption", prompt.Options{
			DefaultValue: strings.ToUpper(s.encryptionType),
			Description:  "OPTIONS\n (C)lient - 16 or 32 character encryption key \n (S)erver - override S3 Bucket KMS Key ID \n (N)one"}, io))
	}

	switch s.encryptionType {
	case eTypeClient:

		s.settings[clientEncryptionToken] = setting.Setting{
			Description:  "Specify a 16 or 32 bit encryption key. Save the key somewhere secure to decrypt the files later.",
			Group:        fmt.Sprintf("CSTORE_%s", strings.ToUpper(s.context)),
			Prop:         fmt.Sprintf("ENCRYPTION_KEY_%s", strings.ToUpper(file.Key())),
			DefaultValue: cipher.GenKey(32),
			Prompt:       promptUser,
			Set:          true,
			Vault:        access,
		}

	case eTypeServer:

		s.settings[serverEncryptionToken] = setting.Setting{
			Description: "Specify the AWS KMS Key ID to use for server side encryption.",
			Group:       "AWS",
			Prop:        "KMS_KEY_ID",
			Prompt:      promptUser,
			Set:         true,
			Vault:       access,
		}

	}

	//------------------------------------------
	//- Open connection to store.
	//------------------------------------------
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	s.Session = sess

	return err
}

// Purge ...
func (s S3Store) Purge(file *catalog.File, version string) error {

	contextKey := s.key(file.Path, version)

	setting, _ := s.settings[awsBucketName]

	bucket, err := setting.Get(s.context, s.io)
	if err != nil {
		return err
	}

	input := s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &contextKey,
	}

	s3svc := s3.New(s.Session)

	if _, err := s3svc.DeleteObject(&input); err != nil {
		return err
	}

	return nil
}

// Push ...
func (s S3Store) Push(file *catalog.File, fileData []byte, version string) error {

	contextKey := s.key(file.Path, version)

	setting, _ := s.settings[awsBucketName]

	bucket, err := setting.Get(s.context, s.io)
	if err != nil {
		return err
	}

	file.AddData(map[string]string{
		awsBucketName: bucket,
	})

	//------------------------------------------
	//- Set client side encryption
	//------------------------------------------
	if key, found := s.settings[clientEncryptionToken]; found {

		value, err := key.Get(s.context, s.io)
		if err != nil {
			return err
		}

		fileData, err = cipher.Encrypt(value, fileData)
		if err != nil {
			return err
		}

		file.AddData(map[string]string{
			fileDataEncryptionKey: clientEncryptionToken,
		})
	} else {
		delete(file.Data, fileDataEncryptionKey)
	}

	input := &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &contextKey,
		Body:   bytes.NewReader(fileData),
	}

	//------------------------------------------
	//- Set server side KMS Key encryption
	//------------------------------------------
	if key, found := s.settings[serverEncryptionToken]; found {

		value, err := key.Get(s.context, s.io)
		if err != nil {
			return err
		}

		etype := "aws:kms"

		input.SSEKMSKeyId = &value
		input.ServerSideEncryption = &etype
	}

	uploader := s3manager.NewUploader(s.Session)

	_, err = uploader.Upload(input)

	return err
}

// Pull ...
func (s S3Store) Pull(file *catalog.File, version string) ([]byte, contract.Attributes, error) {

	contextKey := s.key(file.Path, version)

	setting, _ := s.settings[awsBucketName]
	setting.Prompt = false

	bucket, err := setting.Get(s.context, s.io)
	if err != nil {
		return []byte{}, contract.Attributes{}, err
	}

	input := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &contextKey,
	}

	s3svc := s3.New(s.Session)

	fileData, err := s3svc.GetObject(&input)
	if err != nil {
		return []byte{}, contract.Attributes{}, err
	}
	defer fileData.Body.Close()

	b, err := ioutil.ReadAll(fileData.Body)
	if err != nil {
		return b, contract.Attributes{}, err
	}

	//------------------------------------------
	//- Set client side encryption
	//------------------------------------------

	if key, found := s.settings[clientEncryptionToken]; found {
		value, err := key.Get(s.context, s.io)
		if err != nil {
			return b, contract.Attributes{}, fmt.Errorf("%s not found in the %s vault", clientEncryptionToken, key.Vault.Name())
		}

		b, err = cipher.Decrypt(value, b)
		if err != nil {
			return b, contract.Attributes{}, err
		}
	}

	return b, contract.Attributes{
		LastModified: *fileData.LastModified,
	}, nil
}

// Changed ...
func (s S3Store) Changed(file *catalog.File, version string) (time.Time, error) {

	contextKey := s.key(file.Path, version)

	setting, _ := s.settings[awsBucketName]
	setting.Prompt = false

	bucket, err := setting.Get(s.context, s.io)
	if err != nil {
		return time.Time{}, err
	}

	input := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &contextKey,
	}

	s3svc := s3.New(s.Session)

	fileData, err := s3svc.GetObject(&input)

	if aerr, ok := err.(awserr.Error); ok {
		if aerr.Code() == s3.ErrCodeNoSuchKey {
			return time.Time{}, nil
		}

		return time.Time{}, err
	}

	defer fileData.Body.Close()

	return *fileData.LastModified, nil
}

func init() {
	s := new(S3Store)
	stores[s.Name()] = s
}

//------------------------------------------
//- Create S3 bucket key.
//------------------------------------------
func (s S3Store) key(path, version string) string {

	if len(version) > 0 {
		return fmt.Sprintf("%s/%s/%s", s.context, version, path)
	}

	return fmt.Sprintf("%s/%s", s.context, path)
}

func getEncryptionType(file catalog.File) string {
	if _, found := file.Data[fileDataEncryptionKey]; found {
		return eTypeClient
	}
	return none
}
