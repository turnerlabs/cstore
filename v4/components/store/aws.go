package store

import "time"

const (
	awsRegion          = "AWS_REGION"
	awsProfile         = "AWS_PROFILE"
	awsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	awsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	awsSessionToken    = "AWS_SESSION_TOKEN"

	awsBucketSetting = "AWS_S3_BUCKET"

	awsStoreKMSKeyID = "AWS_STORE_KMS_KEY_ID"

	awsDefaultRegion  = "us-east-1"
	awsDefaultProfile = "default"

	defaultSMKMSKey = "aws/secretsmanager"
)

type kmsKeyID struct {
	value         string
	awsInputValue string
}

type secret struct {
	name  string
	value string

	keyID        string
	lastModified time.Time
}
