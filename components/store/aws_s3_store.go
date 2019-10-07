package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/setting"
	"github.com/turnerlabs/cstore/components/vault"
)

// S3Store ...
type S3Store struct {
	Session *session.Session

	clog catalog.Catalog

	uo cfg.UserOptions
	io models.IO

	bucket setting.Setting
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

	s.clog = clog
	s.io = io
	s.uo = uo

	//------------------------------------------
	//- Store Configuration
	//------------------------------------------
	s.bucket = setting.Setting{
		Description:  "S3 Bucket that will store the file.",
		Prop:         awsBucketSetting,
		Prompt:       uo.Prompt,
		Silent:       uo.Silent,
		AutoSave:     true,
		DefaultValue: clog.GetDataByStore(s.Name(), awsBucketSetting, fmt.Sprintf("%s-configs", clog.Context)),
		Vault:        file,
	}

	//------------------------------------------
	//- Get AWS Region
	//------------------------------------------
	region, err := setting.Setting{
		Description:  fmt.Sprintf("Silence this %s store prompt by setting environment variable.", s.Name()),
		Group:        clog.Context,
		Prop:         "AWS_REGION",
		Prompt:       uo.Prompt,
		Silent:       uo.Silent,
		AutoSave:     true,
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
		Description: fmt.Sprintf("Store credential for %s in %s.", s.Name(), access.Name()),
		Group:       clog.Context,
		Prop:        "AWS_ACCESS_KEY_ID",
		Prompt:      uo.Prompt,
		Silent:      uo.Silent,
		AutoSave:    true,
		Vault:       access,
	}.Get(clog.Context, io)
	if err != nil {
		return err
	}

	secret, err := setting.Setting{
		Description: fmt.Sprintf("Store credential for %s in %s.", s.Name(), access.Name()),
		Group:       clog.Context,
		Prop:        "AWS_SECRET_ACCESS_KEY",
		Prompt:      uo.Prompt,
		Silent:      uo.Silent,
		AutoSave:    true,
		Vault:       access,
	}.Get(clog.Context, io)
	if err != nil {
		return err
	}

	token, err := setting.Setting{
		Description: fmt.Sprintf("Store credential for %s in %s.", s.Name(), access.Name()),
		Group:       clog.Context,
		Prop:        "AWS_SESSION_TOKEN",
		Prompt:      uo.Prompt,
		Silent:      uo.Silent,
		AutoSave:    true,
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

// Purge ...
func (s S3Store) Purge(file *catalog.File, version string) error {

	contextKey := s.key(file.Path, version)

	bucket, err := s.bucket.Get(s.clog.Context, s.io)
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

	bucket, err := s.bucket.Get(s.clog.Context, s.io)
	if err != nil {
		return err
	}

	file.AddData(map[string]string{
		awsBucketSetting: bucket,
	})

	input := &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &contextKey,
		Body:   bytes.NewReader(fileData),
	}

	//------------------------------------------
	//- Set server side KMS Key encryption
	//------------------------------------------
	value, err := setting.Setting{
		Description:  "KMS Key ID is used by S3 to encrypt and decrypt secrets. Any role or user accessing a secret must also have access to the KMS key. Leave blank to use the default bucket encryption settings.",
		Prop:         awsStoreKMSKeyID,
		Prompt:       s.uo.Prompt,
		Silent:       s.uo.Silent,
		DefaultValue: s.clog.GetDataByStore(s.Name(), awsStoreKMSKeyID, ""),
		AutoSave:     false,
		Vault:        file,
	}.Get(s.clog.Context, s.io)

	if len(value) > 0 {
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

	setting := s.bucket
	setting.Prompt = false

	bucket, err := setting.Get(s.clog.Context, s.io)
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

	setting := s.bucket
	setting.Prompt = false

	bucket, err := setting.Get(s.clog.Context, s.io)
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
		return fmt.Sprintf("%s/%s/%s", s.clog.Context, version, path)
	}

	return fmt.Sprintf("%s/%s", s.clog.Context, path)
}
