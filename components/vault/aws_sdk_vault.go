package vault

import (
	"fmt"
)

// AWSSDKVault ...
type AWSSDKVault struct{}

// Name ...
func (v AWSSDKVault) Name() string {
	return "aws-sdk"
}

// Description ...
func (v AWSSDKVault) Description() string {
	return `This vault uses the standard AWS credential chain to authenticate with AWS before accessing an AWS store like an S3 bucket or Parameter Store.`
}

// Set ...
func (v AWSSDKVault) Set(contextID, key, value string) error {
	return fmt.Errorf("command not applicable to vault")
}

// Delete ...
func (v AWSSDKVault) Delete(contextID, key string) error {
	return fmt.Errorf("command not applicable to vault")
}

// Get ...
func (v AWSSDKVault) Get(contextID, key, defaultVal, description string, askUser bool) (string, error) {

	return "", fmt.Errorf("command not applicable to vault")
}

func init() {
	v := AWSSDKVault{}
	vaults[v.Name()] = v
}
