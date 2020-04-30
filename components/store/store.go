package store

import (
	"fmt"

	"github.com/turnerlabs/cstore/v4/components/catalog"
	"github.com/turnerlabs/cstore/v4/components/cfg"
	"github.com/turnerlabs/cstore/v4/components/contract"
	"github.com/turnerlabs/cstore/v4/components/models"
	"github.com/turnerlabs/cstore/v4/components/prompt"
)

const (
	ceKeyName = "CSTORE_ENCRYPTION_KEY"

	// VersionFeature ...
	VersionFeature = "VERSIONING"

	// SourceControlFeature ...
	SourceControlFeature = "SOURCE_CONTROL"

	// EnvFeature ...
	EnvFeature = "env"

	// JSONFeature ...
	JSONFeature = "json"
)

var stores = map[string]contract.IStore{}

// Get ...
func Get() map[string]contract.IStore {
	return stores
}

// Select checks available stores and chooses a default prompting the user if necesary.
func Select(file *catalog.File, clog catalog.Catalog, v contract.IVault, uo cfg.UserOptions, io models.IO) (contract.IStore, error) {

	if len(file.Store) > 0 {
		if store, found := stores[file.Store]; found {
			return store, store.Pre(clog, file, v, uo, io)
		}

		return nil, contract.ErrStoreNotFound
	}

	supportedStores := ""
	for _, s := range Get() {
		if s.SupportsFileType(file.Type) {
			if len(supportedStores) == 0 {
				supportedStores = s.Name()
			} else {
				supportedStores = fmt.Sprintf("%s,%s", supportedStores, s.Name())
			}
		}
	}

	val := prompt.GetValFromUser("Remote Store", prompt.Options{
		Description:  fmt.Sprintf("The remote storage solution where %s data will be pushed. (%s)", file.ActualPath(), supportedStores),
		DefaultValue: GetDefaultStoreFor(file.Type),
	}, io)

	if store, found := stores[val]; found {
		return store, store.Pre(clog, file, v, uo, io)
	}

	return nil, contract.ErrStoreNotFound
}

// GetDefaultStoreFor ...
func GetDefaultStoreFor(fileType string) string {
	switch fileType {
	case "env":
		return AWSParameterStore{}.Name()
	default:
		return S3Store{}.Name()
	}
}
