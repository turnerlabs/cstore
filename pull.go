package cstore

import (
	"bytes"
	"fmt"

	"github.com/subosito/gotenv"
	"github.com/turnerlabs/cstore/v4/components/models"

	"github.com/turnerlabs/cstore/v4/components/remote"

	"github.com/turnerlabs/cstore/v4/components/catalog"
	"github.com/turnerlabs/cstore/v4/components/cfg"
	"github.com/turnerlabs/cstore/v4/components/path"
	"github.com/turnerlabs/cstore/v4/components/token"
)

// Pull retrieves configuration fron a remote store using cstore.yml
func Pull(catalogPath string, o Options) ([]byte, error) {

	opt := o.ToUserOptions()

	data := []byte{}

	//-------------------------------------------------
	//- Get the local catalog for reference.
	//-------------------------------------------------
	clog, err := catalog.Get(catalogPath)
	if err != nil {
		return data, err
	}

	root := path.RemoveFileName(catalogPath)

	files := clog.FilesBy(opt.GetPaths(clog.CWD), opt.TagList, opt.AllTags, opt.Version)

	if len(opt.Version) > 0 && len(files) == 0 {
		files = clog.FilesBy(opt.GetPaths(clog.CWD), opt.TagList, opt.AllTags, "")
	}

	if len(files) == 0 {
		return data, fmt.Errorf("FileNotFoundError: file not found in %s", catalogPath)
	}

	for _, fileEntry := range files {

		// //-----------------------------------------------------
		// //- Override saved file settings with user preferences.
		// //-----------------------------------------------------
		fileEntry = remote.OverrideFileSettings(fileEntry, opt)

		//----------------------------------------------------
		//- Check for a linked catalog with child files.
		//----------------------------------------------------
		if fileEntry.IsRef {
			tempConfig, err := Pull(path.BuildPath(root, fileEntry.Path), o)
			if err != nil {
				return tempConfig, err
			}

			data = append(data, tempConfig...)

			continue
		}

		//----------------------------------------------------
		//- Get the remote store and vaults components ready.
		//----------------------------------------------------
		fileEntryTemp := fileEntry
		remoteComp, err := remote.InitComponents(&fileEntryTemp, clog, opt, models.IO{})
		if err != nil {

			p := path.BuildPath(root, fileEntry.Path)
			if len(opt.Version) > 0 {
				p = fmt.Sprintf("%s (%s)", p, opt.Version)
			}

			return data, fmt.Errorf("PullFailedError1: %s (%s)", p, err)
		}

		//----------------------------------------------------
		//- Pull remote file from store.
		//----------------------------------------------------
		file, _, err := remoteComp.Store.Pull(&fileEntry, opt.Version)
		if err != nil {

			p := path.BuildPath(root, fileEntry.Path)
			if len(opt.Version) > 0 {
				p = fmt.Sprintf("%s (%s)", p, opt.Version)
			}

			return data, fmt.Errorf("PullFailedError2: %s (%s)", p, err)
		}

		//-------------------------------------------------
		//- If user specifies, inject secrets into file.
		//-------------------------------------------------
		fileWithSecrets := file

		if opt.InjectSecrets {
			if !fileEntry.SupportsSecrets() {
				return data, fmt.Errorf("IncompatibleFileError: %s secrets not supported", fileEntry.Path)
			}

			tokens, err := token.Find(fileWithSecrets, fileEntry.Type, false)
			if err != nil {
				return data, fmt.Errorf("MissingTokensError: failed to find tokens in file %s (%s)", fileEntry.Path, err)
			}

			for k, t := range tokens {

				value, err := remoteComp.Secrets.Get(clog.Context, t.Secret(), t.Prop)
				if err != nil {
					return data, fmt.Errorf("GetSecretValueError: failed to get value for %s/%s for %s (%s)", t.Secret(), t.Prop, path.BuildPath(root, fileEntry.Path), err)
				}

				t.Value = value
				tokens[k] = t
			}

			fileWithSecrets, err = token.Replace(fileWithSecrets, fileEntry.Type, tokens, false)
			if err != nil {
				return data, fmt.Errorf("TokenReplacementError: failed to replace tokens in file %s (%s)", fileEntry.Path, err)
			}
		}

		data = fileWithSecrets
	}

	return data, nil
}

// PullEnv retrieves configuration stored in .env format as a map
func PullEnv(catalogPath string, o Options) (map[string]string, error) {

	config := map[string]string{}

	b, err := Pull(catalogPath, o)
	if err != nil {
		return config, err
	}

	envvars := gotenv.Parse(bytes.NewReader(b))

	for k, v := range envvars {
		config[k] = v
	}

	return config, nil
}

type Options struct {
	AllTags       bool
	Tags          []string
	Paths         []string
	Version       string
	InjectSecrets bool
}

func (o Options) ToUserOptions() cfg.UserOptions {
	return cfg.UserOptions{
		Paths:         o.Paths,
		Version:       o.Version,
		TagList:       o.Tags,
		AllTags:       o.AllTags,
		InjectSecrets: o.InjectSecrets,
		Silent:        true,
	}
}
