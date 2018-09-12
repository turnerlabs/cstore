package vault

import (
	"errors"

	"github.com/turnerlabs/cstore/components/cfg"
)

// ErrSecretNotFound is returned by the vault when the
// requested key cannot be found in the vault.
var ErrSecretNotFound = errors.New("secret not found")

// IVault ...
type IVault interface {
	// Name should return an identifier used by the store to determine
	// which vault to use when checking for secrets.
	Name() string

	// Description provides details on how to use the vault.
	Description() string

	// Get should return the secret for the key or the error
	// ErrSecretNotFound when the secret is not in the vault.
	Get(contextID, key, defaultValue, description string, askUser bool) (string, error)

	// Get should set the secret under the current key or
	// return an error.
	Set(contextID, key, value string) error

	// Delete should remove the secret under the current key or
	// return an error.
	Delete(contextID, key string) error
}

var vaults = map[string]IVault{}

// Get ...
func Get() map[string]IVault {
	return vaults
}

// GetBy ...
func GetBy(name string) (IVault, error) {
	if len(name) == 0 {
		return vaults[cfg.DefaultVault], nil
	}

	if v, found := vaults[name]; found {
		return v, nil
	}
	return nil, errors.New("vault not found")
}
