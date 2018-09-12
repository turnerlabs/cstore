package store

import (
	"errors"
	"fmt"
	"time"

	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/vault"
)

const (
	ceKeyName = "CSTORE_ENCRYPTION_KEY"

	envVarPrefix = "ENV_"
	envVarType   = "envvar"
)

// ErrStoreNotFound is returned when the requested store
// does not exist.
var ErrStoreNotFound = errors.New("store not found")

// IStore is a persistence abstraction facilitating the
// implementation of different file stores capable of saving,
// getting, and deleting the contents of a file using a key.
//
// To add a new store, create a struct implementing this interface
// in a separate file using the naming convention `{type}_store.go`.
type IStore interface {

	// Name is unique value that can be an argument pass to a command.
	Name() string

	// CanHandleFile describes what files can be stored in this store.
	CanHandleFile(catalog.File) bool

	// Description provides details on how to use the store.
	Description() string

	// Pre ensures required data for the store to operate exists
	// and is valid before performing any store actions. It should
	// return nil to indicate the store is ready to be used.
	//
	// The vaults are used to get secrets required by the store. It
	// is common to set the vaults on the struct to allow methods
	// access to them.
	//
	// Often store authentication, validation, and defaulting are
	// performed in Pre.
	//
	// ContextID is the unique identifier for the yml file created
	// when content is pushed to the store. It can be used for naming
	// of things that should be unique to this yml files context.
	Pre(contextID string, file catalog.File, creds vault.IVault, encrypt vault.IVault, prompt bool) error

	// Push is called when a file needs to be remotely stored.
	// The file should be stored using the key parameter.
	//
	// The interface slice returned is data stored in the catalog
	// It can be anything the store needs.
	//
	// The boolean returned should be true if content was encrypted.
	Push(contextKey string, file catalog.File, fileData []byte) (map[string]string, bool, error)

	// Push is called when a file needs to be retrieved from the remote store.
	// The file should be retrieved using the key parameter.
	Pull(contextKey string, file catalog.File) ([]byte, Attributes, error)

	// Purge is called when a file needs to be deleted from the remote store.
	// The file should be deleted using the key parameter.
	Purge(contextKey string, file catalog.File) error
}

// Attributes ...
type Attributes struct {
	LastModified time.Time
}

var stores = map[string]IStore{}

// Get ...
func Get() map[string]IStore {
	return stores
}

// Select checks available stores and chooses a default.
func Select(file catalog.File, contextID string, cv vault.IVault, ev vault.IVault, prompt bool) (IStore, error) {

	if len(file.Store) > 0 {
		if store, found := stores[file.Store]; found {
			return store, store.Pre(contextID, file, cv, ev, prompt)
		}

		return nil, ErrStoreNotFound
	}

	if store, found := stores[cfg.DefaultStore]; found {
		return store, store.Pre(contextID, file, cv, ev, prompt)
	}

	return nil, ErrStoreNotFound
}

func addEnvVarPrefix(envvar string) string {
	return fmt.Sprintf("%s%s", envVarPrefix, envvar)
}

func removeEnvVarPrefix(envvar string) string {
	return envvar[len(envVarPrefix):len(envvar)]
}
