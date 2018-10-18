package vault

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/turnerlabs/cstore/components/contract"
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

// BuildKey ...
func (v AWSParameterStoreVault) BuildKey(contextID, group, prop string) string {
	return fmt.Sprintf("/cstore/%s/%s/%s", contextID, group, prop)
}

// Pre ...
func (v AWSParameterStoreVault) Pre(contextID string) error {
	return nil
}

// Set ...
func (v AWSParameterStoreVault) Set(contextID, group, prop, value string) error {
	return setValueInParamStore(v.BuildKey(contextID, group, prop), value)
}

// Delete ...
func (v AWSParameterStoreVault) Delete(contextID, group, prop string) error {
	return deleteKeyInParamStore(v.BuildKey(contextID, group, prop))
}

// Get ...
func (v AWSParameterStoreVault) Get(contextID, group, prop string) (string, error) {
	return getFromParamStore(v.BuildKey(contextID, group, prop))
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
			return contract.ErrSecretNotFound
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
		return "", contract.ErrSecretNotFound
	}

	return *output.Parameters[0].Value, nil
}

func init() {
	//--------------------------------
	//- Disabled until converted to v2
	//--------------------------------
	// v := AWSParameterStoreVault{}
	// vaults[v.Name()] = v
}
