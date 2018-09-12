package vault

import (
	"os"

	"github.com/turnerlabs/cstore/components/prompt"
)

// EnvVault ...
type EnvVault struct{}

// Name ...
func (v EnvVault) Name() string {
	return "env"
}

// Description ...
func (v EnvVault) Description() string {
	return `An env vault uses environment variables or credential files to get access and encryption information.`
}

// Set ...
func (v EnvVault) Set(contextID, key, value string) error {
	return os.Setenv(key, value)
}

// Delete ...
func (v EnvVault) Delete(contextID, key string) error {
	return os.Unsetenv(key)
}

// Get ...
func (v EnvVault) Get(contextID, key, defaultVal, description string, askUser bool) (string, error) {

	if len(os.Getenv(key)) == 0 && askUser {
		if val := prompt.GetValFromUser(key, defaultVal, description, true); len(val) > 0 {
			os.Setenv(key, val)

			return val, nil
		}

		return "", ErrSecretNotFound
	}

	return os.Getenv(key), nil
}

func init() {
	v := EnvVault{}
	vaults[v.Name()] = v
}
