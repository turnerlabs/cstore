package vault

import (
	"errors"

	"github.com/turnerlabs/cstore/v4/components/cfg"

	"github.com/turnerlabs/cstore/v4/components/catalog"
	"github.com/turnerlabs/cstore/v4/components/contract"
	"github.com/turnerlabs/cstore/v4/components/models"
)

var vaults = map[string]contract.IVault{}

// Get ...
func Get() map[string]contract.IVault {
	return vaults
}

// GetBy ...
func GetBy(name, defaultVault string, clog catalog.Catalog, fileEntry *catalog.File, access contract.IVault, uo cfg.UserOptions, io models.IO) (contract.IVault, error) {
	if len(name) == 0 {
		v := vaults[defaultVault]
		return v, v.Pre(clog, fileEntry, access, uo, io)
	}

	if v, found := vaults[name]; found {
		return v, v.Pre(clog, fileEntry, access, uo, io)
	}
	return nil, errors.New("vault not found")
}
