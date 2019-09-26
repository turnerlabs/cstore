package remote

import (
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
)

// OverrideFileSettings ...
func OverrideFileSettings(fileEntry catalog.File, opt cfg.UserOptions) catalog.File {

	if len(opt.SecretsVault) > 0 {
		fileEntry.Vaults.Secrets = opt.SecretsVault
	}

	if len(opt.AccessVault) > 0 {
		fileEntry.Vaults.Access = opt.AccessVault
	}

	if len(opt.AlternateRestorePath) > 0 {
		fileEntry.AlternatePath = opt.AlternateRestorePath
	}

	return fileEntry
}
