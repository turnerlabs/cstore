package cmd

import (
	"bytes"
	"fmt"

	"github.com/subosito/gotenv"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/store"
	"github.com/turnerlabs/cstore/components/vault"
)

//"\xE2\x9C\x94" This is a checkmark on mac, but question mark on windows; so,
// will use (done) for now to support multiple platforms.
const (
	checkMark = "(done)"
	none      = ""
)

// If the user specifies options during file push, make sure the exiting
// file options are overridden with the desired user options.
func updateUserOptions(file catalog.File, opt cfg.UserOptions) catalog.File {

	if len(opt.AlternateRestorePath) > 0 {
		file.AternatePath = opt.AlternateRestorePath
	}

	if len(opt.Paths) > 0 && len(opt.Tags) > 0 {
		file.Tags = opt.TagList
	} else if len(file.Tags) == 0 {
		file.Tags = opt.TagsFrom(file.Path)
	}

	if len(opt.Store) > 0 {
		file.Store = opt.Store
	}

	if len(opt.SecretsVault) > 0 {
		file.Vaults.Secrets = opt.SecretsVault
	}

	if len(opt.AccessVault) > 0 {
		file.Vaults.Access = opt.AccessVault
	}

	return file
}

type remoteComponents struct {
	store   contract.IStore
	access  contract.IVault
	secrets contract.IVault
}

func getRemoteComponents(fileEntry *catalog.File, clog catalog.Catalog, uo cfg.UserOptions, io models.IO) (remoteComponents, error) {
	remote := remoteComponents{}

	v, err := vault.GetBy(fileEntry.Vaults.Secrets, cfg.DefaultSecretsVault, clog, fileEntry, uo.Prompt, io)
	if err != nil {
		return remote, err
	}
	remote.secrets = v
	fileEntry.Vaults.Secrets = v.Name()

	v, err = vault.GetBy(fileEntry.Vaults.Access, cfg.DefaultAccessVault, clog, fileEntry, uo.Prompt, io)
	if err != nil {
		return remote, err
	}
	remote.access = v
	fileEntry.Vaults.Access = v.Name()

	st, err := store.Select(fileEntry, clog, remote.access, uo, io)
	if err != nil {
		return remote, err
	}
	remote.store = st
	fileEntry.Store = st.Name()

	return remote, nil
}

func getFilePathsToPush(clog catalog.Catalog, opt cfg.UserOptions) []string {
	paths := opt.GetPaths(clog.CWD)

	if len(paths) == 0 {
		if len(opt.TagList) > 0 {
			paths = clog.GetPathsBy(opt.TagList, opt.AllTags)
		} else {
			paths = clog.GetPaths()
		}
	}

	return removeDups(paths)
}

func removeDups(elements []string) []string {
	unique := map[string]string{}

	for v := range elements {
		unique[elements[v]] = elements[v]
	}

	result := []string{}
	for key := range unique {
		result = append(result, key)
	}

	return result
}

func overrideFileSettings(fileEntry catalog.File, opt cfg.UserOptions) catalog.File {

	if len(opt.SecretsVault) > 0 {
		fileEntry.Vaults.Secrets = opt.SecretsVault
	}

	if len(opt.AccessVault) > 0 {
		fileEntry.Vaults.Access = opt.AccessVault
	}

	if len(opt.AlternateRestorePath) > 0 {
		fileEntry.AternatePath = opt.AlternateRestorePath
	}

	return fileEntry
}

func bufferExportScript(file []byte) bytes.Buffer {
	reader := bytes.NewReader(file)
	pairs := gotenv.Parse(reader)
	var b bytes.Buffer

	for key, value := range pairs {
		b.WriteString(fmt.Sprintf("export %s='%s'\n", key, value))
	}

	return b
}
