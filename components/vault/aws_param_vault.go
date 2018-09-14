package vault

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/turnerlabs/cstore/components/prompt"
)

// AWSParameterStoreVault ...
type AWSParameterStoreVault struct{}

// Name ...
func (v AWSParameterStoreVault) Name() string {
	return "aws-parameter"
}

// Description ...
func (v AWSParameterStoreVault) Description() string {
	return `This vault retrieves secrets from the AWS Parameter Store. This allows a store to retrieve securely stored values like encryption keys and passwords.`
}

// Set ...
func (v AWSParameterStoreVault) Set(contextID, key, value string) error {
	compositeKey := buildParamKey(contextID, key)

	return setValueInParamStore(compositeKey, value)
}

// Delete ...
func (v AWSParameterStoreVault) Delete(contextID, key string) error {
	compositeKey := buildParamKey(contextID, key)

	return deleteKeyInParamStore(compositeKey)
}

// Get ...
func (v AWSParameterStoreVault) Get(contextID, key, defaultVal, description string, askUser bool) (string, error) {
	compositeKey := buildParamKey(contextID, key)

	val, err := getFromParamStore(compositeKey)

	if err != nil {
		if err == ErrSecretNotFound && askUser {
			val = prompt.GetValFromUser(key, defaultVal, description, true)

			if len(val) > 0 {
				if err := setValueInParamStore(compositeKey, val); err != nil {
					return val, err
				}
			}
		} else {
			return val, err
		}
	}

	return val, nil
}

func deleteKeyInParamStore(key string) error {

	input := ssm.DeleteParameterInput{
		Name: &key,
	}

	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	svc := ssm.New(sess)

	_, err = svc.DeleteParameter(&input)
	if err != nil {
		if err.Error() != ssm.ErrCodeParameterNotFound {
			return ErrSecretNotFound
		}
	}

	return nil
}

func setValueInParamStore(key, value string) error {

	pType := "String"
	overwrite := true

	input := ssm.PutParameterInput{
		Name:      &key,
		Type:      &pType,
		Value:     &value,
		Overwrite: &overwrite,
	}

	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	svc := ssm.New(sess)

	_, err = svc.PutParameter(&input)
	if err != nil {
		return err
	}

	return nil
}

func getFromParamStore(key string) (string, error) {
	sess, err := session.NewSession()
	if err != nil {
		return "", err
	}

	svc := ssm.New(sess)

	names := []*string{&key}
	input2 := ssm.GetParametersInput{
		Names: names,
	}

	output, err := svc.GetParameters(&input2)
	if err != nil {
		return "", err
	}

	if len(output.Parameters) == 0 {
		return "", ErrSecretNotFound
	}

	return *output.Parameters[0].Value, nil
}

func buildParamKey(contextID, name string) string {
	return fmt.Sprintf("/cstore/%s/%s", contextID, name)
}

func init() {
	v := AWSParameterStoreVault{}
	vaults[v.Name()] = v
}
