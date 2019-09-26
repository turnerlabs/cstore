package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/subosito/gotenv"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
)

//"\xE2\x9C\x94" This is a checkmark on mac, but question mark on windows; so,
// will use (done) for now to support multiple platforms.
const (
	checkMark = "(done)"
	none      = ""
)

// If the user specifies options during file push, make sure the exiting
// file options are overridden with the desired user options.
func updateUserOptions(file catalog.File, fileUpdate bool, opt cfg.UserOptions) catalog.File {

	if len(opt.AlternateRestorePath) > 0 {
		file.AlternatePath = opt.AlternateRestorePath
	}

	if b, err := strconv.ParseBool(opt.DeleteLocalFiles); err == nil {
		file.DeleteAfterPush = b
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

func bufferExportScript(file []byte) (bytes.Buffer, error) {
	reader := bytes.NewReader(file)
	pairs := gotenv.Parse(reader)
	var b bytes.Buffer

	for key, value := range pairs {
		_, err := b.WriteString(fmt.Sprintf("export %s='%s'\n", key, value))
		if err != nil {
			return b, err
		}
	}

	return b, nil
}

func toTaskDefSecretFormat(file []byte) (bytes.Buffer, error) {
	reader := bytes.NewReader(file)
	pairs := gotenv.Parse(reader)

	var buff bytes.Buffer

	secrets := []JsonFormat{}

	for key, value := range pairs {
		p := JsonFormat{
			ValueFrom: value,
			Name:      key,
		}

		secrets = append(secrets, p)
	}

	b, err := json.MarshalIndent(secrets, "", "    ")
	if err != nil {
		return buff, err
	}

	_, err = buff.Write(b)
	if err != nil {
		return buff, err
	}

	return buff, nil
}

type JsonFormat struct {
	Name      string `json:"name"`
	ValueFrom string `json:"valueFrom"`
}

func toTaskDefEnvFormat(file []byte) (bytes.Buffer, error) {
	reader := bytes.NewReader(file)
	pairs := gotenv.Parse(reader)

	var buff bytes.Buffer

	env := []EnvFormat{}

	for key, value := range pairs {
		p := EnvFormat{
			Value: value,
			Name:  key,
		}

		env = append(env, p)
	}

	b, err := json.MarshalIndent(env, "", "    ")
	if err != nil {
		return buff, err
	}

	_, err = buff.Write(b)
	if err != nil {
		return buff, err
	}

	return buff, nil
}

type EnvFormat struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
