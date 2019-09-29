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
	"github.com/turnerlabs/cstore/components/cfg"
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

// SupportsFeature ...
func (s S3Store) SupportsFeature(feature string) bool {
	switch feature {
	case VersionFeature:
		return true
	default:
		return false
	}
}

// SupportsFileType ...
func (s S3Store) SupportsFileType(fileType string) bool {
	return true
}

// Description ...
func (s S3Store) Description() string {
	return `
	details: https://github.com/turnerlabs/cstore/blob/master/docs/S3.md
`
}

// Pre ...
func (s *S3Store) Pre(clog catalog.Catalog, file *catalog.File, access contract.IVault, uo cfg.UserOptions, io models.IO) error {
	s.settings = map[string]setting.Setting{}
	s.context = clog.Context
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

	//---------------------------------------------
	//- Store authentication and encryption options
	//---------------------------------------------
	if uo.Prompt {
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
	//- Store Configuration
	//------------------------------------------
	s.settings[awsBucketName] = setting.Setting{
		Description:  "S3 Bucket that will store the file.",
		Group:        "AWS",
		Prop:         "S3_BUCKET",
		Prompt:       uo.Prompt,
		Silent:       uo.Silent,
		AutoSave:     true,
		DefaultValue: clog.GetAnyDataBy(awsBucketName, fmt.Sprintf("cstore-%s", clog.Context)),
		Vault:        file,
	}

	//------------------------------------------
	//- Encryption
	//------------------------------------------
	s.settings[serverEncryptionToken] = setting.Setting{
		Description:  "KMS Key ID is used by S3 to encrypt and decrypt secrets. Any role or user accessing a secret must also have access to the KMS key. Leave blank to use the default bucket encryption settings.",
		Group:        "AWS",
		Prop:         "STORE_KMS_KEY_ID",
		Prompt:       uo.Prompt,
		Silent:       uo.Silent,
		DefaultValue: clog.GetAnyDataBy("AWS_STORE_KMS_KEY_ID", ""),
		AutoSave:     false,
		Vault:        file,
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

		if len(value) > 0 {
			etype := "aws:kms"

			input.SSEKMSKeyId = &value
			input.ServerSideEncryption = &etype
		}
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

	return b, contract.Attributes{}, nil
}

// Changed ...
func (s S3Store) Changed(file *catalog.File, fileData []byte, version string) (time.Time, error) {

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

	fileMetaData, err := s3svc.GetObject(&input)

	if aerr, ok := err.(awserr.Error); ok {
		if aerr.Code() == s3.ErrCodeNoSuchKey {
			return time.Time{}, nil
		}

		return time.Time{}, err
	}

	defer fileMetaData.Body.Close()

	return *fileMetaData.LastModified, nil
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
