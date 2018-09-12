package store

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/logger"
	"github.com/turnerlabs/cstore/components/prompt"
	"github.com/turnerlabs/cstore/components/vault"
)

const (
	awsRegion          = "AWS_REGION"
	awsProfile         = "AWS_PROFILE"
	awsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	awsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	awsKMSKeyID        = "AWS_KMS_KEY_ID"

	awsDefaultRegion  = "us-east-1"
	awsDefaultProfile = "default"

	eTypeBoth   = "b"
	eTypeClient = "c"
	eTypeServer = "s"

	eTargetNone = "s"
	eTargetFile = "f"
	eTargetBoth = "b"

	description = `%s To authenticate with AWS, credentials must be provided in one of three ways.

1. Add keys %s and %s in OSX Keychain.
2. Add environment variables %s and %s.
3. Add profile '%s' to AWS credentials file.

The OSX Keychain option requires the vault flag to set up the keychain vault when performing the first push. After the first push, future actions will use the catalog to determine the vault.

Example: $ cstore push {file} -c osx-keychain
`
)

func turnOnAWSProfileLookup() {
	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
}

func setupSessionCreds(contextID string, cv vault.IVault, promptUser bool) error {

	desc := fmt.Sprintf("Use a profile, vault, or export %s and %s env variables to automate this step.", awsAccessKeyID, awsSecretAccessKey)

	if value, err := cv.Get(contextID, awsAccessKeyID, "", desc, promptUser); err == nil {
		os.Setenv(awsAccessKeyID, value)
	}

	if value, err := cv.Get(contextID, awsSecretAccessKey, "", "", promptUser); err == nil {
		os.Setenv(awsSecretAccessKey, value)
	}

	if err := sessionCredsExist(); err != nil {
		return err
	}

	return nil
}

func setupProfileCreds(promptUser bool) error {
	turnOnAWSProfileLookup()

	desc := fmt.Sprintf("Export %s env variable to automate this step.", awsProfile)

	if len(os.Getenv(awsProfile)) == 0 && promptUser {
		os.Setenv(awsProfile, prompt.GetValFromUser(awsProfile, awsDefaultProfile, desc, false))
	}

	return nil
}

func setupAWS(contextID string, file catalog.File, cv vault.IVault, ev vault.IVault, promptUser bool) (sess *session.Session, encryptionKey string, KMSKeyID string, err error) {

	if len(os.Getenv(awsRegion)) == 0 {
		if promptUser {
			os.Setenv(awsRegion, prompt.GetValFromUser(awsRegion, awsDefaultRegion, "", false))
		} else {
			os.Setenv(awsRegion, awsDefaultRegion)
		}
	}

	credsProvided := false

	if cv.Name() == "env" {
		setupProfileCreds(true)

		if err := sharedCredsExist(os.Getenv(awsProfile)); err != nil {
			logger.L.Print(err)
		} else {
			credsProvided = true
		}

		if !credsProvided {

			if err = setupSessionCreds(contextID, cv, true); err != nil {
				return sess, encryptionKey, KMSKeyID, err
			}

			if err := sessionCredsExist(); err != nil {
				return sess, encryptionKey, KMSKeyID, err
			}

			credsProvided = true
		}
	}

	sess, err = session.NewSession()
	if err != nil {
		return sess, encryptionKey, KMSKeyID, err
	}

	encryptionKey, KMSKeyID, err = configureEncryption(ev, contextID, file, promptUser)
	if err != nil {
		return sess, encryptionKey, KMSKeyID, err
	}

	return sess, encryptionKey, KMSKeyID, err
}

func encryptionConfigMissing(encryptionKey, KMSKeyID string) bool {
	return len(KMSKeyID) == 0 && len(encryptionKey) == 0
}

func getKeys(contextID, key string) []string {
	return []string{
		buildEncryptionToken(ceKeyName, contextID, key),
		buildEncryptionToken(ceKeyName, contextID, ""),
		buildEncryptionToken(awsKMSKeyID, contextID, key),
		buildEncryptionToken(awsKMSKeyID, contextID, ""),
	}
}

func removeKeys(ev vault.IVault, contextID string, file catalog.File) error {
	keys := getKeys(contextID, file.Key())

	for _, key := range keys {
		if err := ev.Delete(contextID, key); err != nil {
			if err != vault.ErrSecretNotFound {
				return err
			}
		}
	}

	return nil
}

func configureEncryption(ev vault.IVault, contextID string, file catalog.File, promptUser bool) (encryptionKey string, KMSKeyID string, err error) {

	encryptionPrompts := promptUser

	if promptUser {
		if protect := prompt.GetValFromUser(fmt.Sprintf("Encrypt/Decrypt '%s'", file.Path), "y/N", "", false); strings.ToLower(protect) == "y" {
			encryptionPrompts = true
		} else {
			encryptionPrompts = false
		}
	}

	eType := eTypeBoth
	if encryptionPrompts {
		eType = prompt.GetValFromUser("Encryption Type", eTypeBoth, "OPTIONS\n (c)lient - Requires a 16 or 32 character client encryption key. \n (s)erver - Requires an AWS KMS Key ID. \n (b)oth \n", false)

		if err := removeKeys(ev, contextID, file); err != nil {
			return "", "", err
		}
	}

	if eType == eTypeClient || eType == eTypeBoth {
		fileToken := buildEncryptionToken(ceKeyName, contextID, file.Key())
		catalogToken := buildEncryptionToken(ceKeyName, contextID, "")

		if value, err := ev.Get(contextID, fileToken, "", "Set file specific client encryption key.", encryptionPrompts); err == nil {
			encryptionKey = value
		} else if value, err := ev.Get(contextID, catalogToken, "", "Set default client encryption key.", encryptionPrompts); err == nil {
			encryptionKey = value
		}
	}

	if eType == eTypeServer || eType == eTypeBoth {
		fileToken := buildEncryptionToken(awsKMSKeyID, contextID, file.Key())
		catalogToken := buildEncryptionToken(awsKMSKeyID, contextID, "")

		if value, err := ev.Get(contextID, fileToken, "", "Set file specific KMS Key ID.", encryptionPrompts); err == nil {
			KMSKeyID = value
		} else if value, err := ev.Get(contextID, catalogToken, "", "Set default KMS Key ID.", encryptionPrompts); err == nil {
			KMSKeyID = value
		}
	}

	return
}

func buildEncryptionToken(baseToken, contextID, fileKey string) string {
	if len(fileKey) > 0 {
		return fmt.Sprintf("%s-%s-%s", baseToken, contextID[0:strings.Index(contextID, "-")], fileKey)
	}

	return fmt.Sprintf("%s-%s", baseToken, contextID[0:strings.Index(contextID, "-")])
}

func sessionCredsExist() error {
	creds := credentials.NewEnvCredentials()
	_, err := creds.Get()

	return err
}

func sharedCredsExist(profile string) error {
	creds := credentials.NewCredentials(&credentials.SharedCredentialsProvider{
		Profile: profile,
	})
	_, err := creds.Get()

	return err
}
