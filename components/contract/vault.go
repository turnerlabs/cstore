package contract

import (
	"errors"

	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/models"
)

// IVault ...
type IVault interface {
	// Name should return a unique vault identifier that can be used
	// as a command line flag.
	Name() string

	// Description provides details on how to use the vault. This is
	// displayed to command line users upon request.
	Description() string

	// Pre executed before other vault actions are performed. It should
	// do any prep work like authentication for the vault or getting
	// input from the user. It can persist any data or session in the
	// vault struct for other actions to use.
	//
	// "clog" contains details about the catalog if needed when dealing
	// with a vault.
	//
	// If the vault needs to persist information in the catalog, the
	// "file.Data" map can be used. Since the file is passed by ref,
	// there is no need to return the file for data to be saved.
	//
	// "fileEntry" represents the file this vault will operatate on.
	// If any data needs
	//
	// "uo" specifies if the user requested settings.
	//
	// "io" contains readers and writers that should be used when
	// displaying instructions to or reading data from the command
	// line.
	//
	// "error" should return nil if the operation was successful.
	Pre(clog catalog.Catalog, fileEntry *catalog.File, uo cfg.UserOptions, io models.IO) error

	// Get should return the requested secret or an error.
	//
	// "contextID" is a guid which represents the context of the
	// catalog. It can be used to guarantee uniqueness for secret values.
	//
	// "group" is a collection of props. The same group could be passed
	// with different "prop" name.
	//
	// "prop" is the name for the value being retrieved.
	//
	// Return ErrSecretNotFound when the secret is not in the vault.
	//
	// "error" should be nil if operation was successful.
	Get(contextID, group, prop string) (string, error)

	// Set should create or update secret value.
	//
	// "contextID" is a guid which represents the context of the
	// catalog. It can be used to guarantee uniqueness for secret values.
	//
	// "group" is a collection of props. The same group could be passed
	// with different props and values. This method should not always
	// overwrite the group, but should append/update the group to ensure
	// other props in the same group are not overwritten.
	//
	// "prop" is the name for the value being set.
	//
	// "value" is the secret being set.
	//
	// "error" should be nil if operation was successful.
	Set(contextID, group, prop, value string) error

	// Delete should remove the secret under the current key or
	// return an error.
	//
	// "contextID" is a guid which represents the context of the
	// catalog. It can be used to guarantee uniqueness for secret values.
	//
	// "group" is a collection of props. The same group could be passed
	// with different "prop" values. So the group should only be deleted
	// when all the props in the group have been deleted.
	//
	// "prop" is the name for the value being deleted.
	//
	// "error" should be nil if operation was successful.
	Delete(contextID, group, prop string) error

	// BuildKey should create the unique key used to store value in vault.
	// It is useful for calling the other operations to ensure the key
	// is built the same way every time.
	BuildKey(contextID, group, prop string) string
}

// ErrSecretNotFound is returned by the vault when the
// requested key cannot be found in the vault.
var ErrSecretNotFound = errors.New("not found")
