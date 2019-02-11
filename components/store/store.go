package store

import (
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/models"
)

const (
	ceKeyName = "CSTORE_ENCRYPTION_KEY"

	// VersionFeature ...
	VersionFeature = "versioning"
)

var stores = map[string]contract.IStore{}

// Get ...
func Get() map[string]contract.IStore {
	return stores
}

// Select checks available stores and chooses a default.
func Select(file *catalog.File, clog catalog.Catalog, v contract.IVault, uo cfg.UserOptions, io models.IO) (contract.IStore, error) {

	if len(file.Store) > 0 {
		if store, found := stores[file.Store]; found {
			return store, store.Pre(clog, file, v, uo, io)
		}

		return nil, contract.ErrStoreNotFound
	}

	if store, found := stores[cfg.DefaultStore]; found {
		return store, store.Pre(clog, file, v, uo, io)
	}

	return nil, contract.ErrStoreNotFound
}
