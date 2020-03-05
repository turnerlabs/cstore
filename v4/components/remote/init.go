package remote

import (
	"github.com/turnerlabs/cstore/v4/components/catalog"
	"github.com/turnerlabs/cstore/v4/components/cfg"
	"github.com/turnerlabs/cstore/v4/components/contract"
	"github.com/turnerlabs/cstore/v4/components/models"
	"github.com/turnerlabs/cstore/v4/components/store"
	"github.com/turnerlabs/cstore/v4/components/vault"
)

// Components ...
type Components struct {
	Store      contract.IStore
	Access     contract.IVault
	Secrets    contract.IVault
	Encryption contract.IVault
}

// InitComponents ...
func InitComponents(fileEntry *catalog.File, clog catalog.Catalog, uo cfg.UserOptions, io models.IO) (Components, error) {
	remote := Components{}

	v, err := vault.GetBy(fileEntry.Vaults.Access, cfg.DefaultAccessVault, clog, fileEntry, nil, uo, io)
	if err != nil {
		return remote, err
	}
	remote.Access = v
	fileEntry.Vaults.Access = v.Name()

	v, err = vault.GetBy(fileEntry.Vaults.Secrets, cfg.DefaultSecretsVault, clog, fileEntry, remote.Access, uo, io)
	if err != nil {
		return remote, err
	}
	remote.Secrets = v
	fileEntry.Vaults.Secrets = v.Name()

	st, err := store.Select(fileEntry, clog, remote.Access, uo, io)
	if err != nil {
		return remote, err
	}
	remote.Store = st
	fileEntry.Store = st.Name()

	return remote, nil
}
