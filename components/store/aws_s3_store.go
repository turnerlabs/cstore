package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cipher"
	"github.com/turnerlabs/cstore/components/prompt"
	"github.com/turnerlabs/cstore/components/vault"
)

const (
	awsBucketName    = "AWS_S3_BUCKET"
	awsDefaultBucket = "cstore"
)

// S3Store ...
type S3Store struct {
	Session *session.Session

	CSKey        string
	CSEncryption bool

	SSKMSKeyID   string
	SSEncryption bool

	EncryptionEnabled bool

	S3Bucket string
}

// Name ...
func (s S3Store) Name() string {
	return "aws-s3"
}

// CanHandleFile ...
func (s S3Store) CanHandleFile(f catalog.File) bool {
	return true
}

// Description ...
func (s S3Store) Description() string {
	return fmt.Sprintf(description, "Any files can be stored in an AWS S3 Bucket.", awsAccessKeyID, awsSecretAccessKey, awsAccessKeyID, awsSecretAccessKey, awsDefaultProfile)
}

// Pre ...
func (s *S3Store) Pre(contextID string, file catalog.File, cv vault.IVault, ev vault.IVault, promptUser bool) error {

	sess, clientKey, KMSKeyID, err := setupAWS(contextID, file, cv, ev, promptUser)

	s.Session = sess
	s.EncryptionEnabled = false

	if len(clientKey) > 0 {
		s.CSKey = clientKey
		s.CSEncryption = true
		s.EncryptionEnabled = true
	}

	if len(KMSKeyID) > 0 {
		s.SSKMSKeyID = KMSKeyID
		s.SSEncryption = true
		s.EncryptionEnabled = true
	}

	if bucket, found := file.Data[awsBucketName]; found {
		s.S3Bucket = bucket
	} else if len(os.Getenv(awsBucketName)) > 0 {
		s.S3Bucket = os.Getenv(awsBucketName)
	} else {
		s.S3Bucket = prompt.GetValFromUser(awsBucketName, awsDefaultBucket, "", false)
	}

	return err
}

// Purge ...
func (s S3Store) Purge(contextKey string, file catalog.File) error {

	input := s3.DeleteObjectInput{
		Bucket: &s.S3Bucket,
		Key:    &contextKey,
	}

	s3svc := s3.New(s.Session)

	if _, err := s3svc.DeleteObject(&input); err != nil {
		return err
	}

	return nil
}

// Push ...
func (s S3Store) Push(contextKey string, file catalog.File, fileData []byte) (map[string]string, bool, error) {

	if s.CSEncryption {
		var err error

		fileData, err = cipher.Encrypt(s.CSKey, fileData)
		if err != nil {
			return nil, s.EncryptionEnabled, err
		}
	}

	uploader := s3manager.NewUploader(s.Session)

	input := &s3manager.UploadInput{
		Bucket: &s.S3Bucket,
		Key:    &contextKey,
		Body:   bytes.NewReader(fileData),
	}

	if s.SSEncryption {
		etype := "aws:kms"

		input.SSEKMSKeyId = &s.SSKMSKeyID
		input.ServerSideEncryption = &etype
	}

	_, err := uploader.Upload(input)

	data := map[string]string{
		awsBucketName: s.S3Bucket,
	}

	return data, s.EncryptionEnabled, err
}

// Pull ...
func (s S3Store) Pull(contextKey string, file catalog.File) ([]byte, Attributes, error) {

	input := s3.GetObjectInput{
		Bucket: &s.S3Bucket,
		Key:    &contextKey,
	}

	s3svc := s3.New(s.Session)

	fileData, err := s3svc.GetObject(&input)
	if err != nil {
		return []byte{}, Attributes{}, err
	}
	defer fileData.Body.Close()

	attr := Attributes{
		LastModified: *fileData.LastModified,
	}

	b, err := ioutil.ReadAll(fileData.Body)
	if err != nil {
		return b, attr, err
	}

	if s.CSEncryption {
		b, err = cipher.Decrypt(s.CSKey, b)
		if err != nil {
			return b, attr, err
		}
	}

	return b, attr, nil
}

func init() {
	s := new(S3Store)
	stores[s.Name()] = s
}
