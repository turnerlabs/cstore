package vault

import (
	"os"
	"strings"

	"github.com/turnerlabs/cstore/v4/components/catalog"
	"github.com/turnerlabs/cstore/v4/components/cfg"
	"github.com/turnerlabs/cstore/v4/components/contract"
	"github.com/turnerlabs/cstore/v4/components/models"
)

// EnvVault ...
type EnvVault struct{}

// Name ...
func (v EnvVault) Name() string {
	return "env"
}

// Description ...
func (v EnvVault) Description() string {
	return `
Secrets are saved and retrieved from environment variables.
	
When using this vault, users are prompted for any required environment variables that are not found in the environment. Once the user enters the value at the prompt the environment variable will only last until the execution of the command is complete.
`
}

// BuildKey ...
func (v EnvVault) BuildKey(contextID, group, prop string) string {
	return strings.ToUpper(prop)
}

// Pre ...
func (v EnvVault) Pre(clog catalog.Catalog, fileEntry *catalog.File, access contract.IVault, uo cfg.UserOptions, io models.IO) error {
	return nil
}

// Set ...
func (v EnvVault) Set(contextID, group, prop, value string) error {
	return os.Setenv(v.BuildKey(contextID, group, prop), value)
}

// Delete ...
func (v EnvVault) Delete(contextID, group, prop string) error {
	return os.Unsetenv(v.BuildKey(contextID, group, prop))
}

// Get ...
func (v EnvVault) Get(contextID, group, prop string) (string, error) {

	if len(os.Getenv(v.BuildKey(contextID, group, prop))) > 0 {
		return os.Getenv(v.BuildKey(contextID, group, prop)), nil
	}

	return "", contract.ErrSecretNotFound
}

func init() {
	v := EnvVault{}
	vaults[v.Name()] = v
}
