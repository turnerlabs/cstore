package vault

import (
	"fmt"

	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/cipher"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/local"
	"github.com/turnerlabs/cstore/components/models"
	yaml "gopkg.in/yaml.v2"
)

const fileName = "file.vlt"
const fileKeyName = "file.vlt.key"

// FileVault ...
type FileVault struct{}

// Name ...
func (v FileVault) Name() string {
	return "file"
}

// Description ...
func (v FileVault) Description() string {
	return fmt.Sprintf(`
Secrets are stored in an encrypted file with the default file '~/.cstore/%s' using a default key '~/.cstore/%s'. 
	
The key can be shared by placing it into the same folder on another machine to allow access to the encrypted vault data.

`, local.BuildPath(fileName), local.BuildPath(fileKeyName))
}

// BuildKey ...
func (v FileVault) BuildKey(contextID, group, prop string) string {
	if len(prop) > 0 {
		return fmt.Sprintf("%s-%s", group, prop)
	}

	return group
}

// Pre ...
func (v FileVault) Pre(clog catalog.Catalog, fileEntry *catalog.File, access contract.IVault, uo cfg.UserOptions, io models.IO) error {
	return nil
}

// Set ...
func (v FileVault) Set(contextID, group, prop, value string) error {
	eKey, _ := getEncryptionKey()
	data, _ := get(fileName, eKey)

	data[v.BuildKey(contextID, group, prop)] = value

	return create(eKey, data)
}

func getEncryptionKey() (string, error) {

	b, err := local.Get(fileKeyName, "")
	if err != nil {
		return cipher.GenerateAES256Key(), err
	}

	return string(b), nil
}

func saveEncryptionKey(eKey string) error {
	return local.Update(fileKeyName, "", []byte(eKey))
}

func create(eKey string, data map[string]string) error {
	d, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	if err = local.Update(fileName, eKey, d); err != nil {
		return err
	}

	return saveEncryptionKey(eKey)
}

// Delete ...
func (v FileVault) Delete(contextID, group, prop string) error {

	if local.Missing(fileName) {
		return nil
	}

	eKey, err := getEncryptionKey()
	if err != nil {
		return err
	}

	data, err := get(fileName, eKey)
	if err != nil {
		return err
	}

	delete(data, v.BuildKey(contextID, group, prop))

	d, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return local.Update(fileName, eKey, d)
}

// Get ...
func (v FileVault) Get(contextID, group, prop string) (string, error) {

	if local.Missing(fileName) {
		return "", contract.ErrSecretNotFound
	}

	eKey, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	data, err := get(fileName, eKey)
	if err != nil {
		return "", err
	}

	if value, found := data[v.BuildKey(contextID, group, prop)]; found {
		if len(value) == 0 {
			return value, contract.ErrSecretNotFound
		}
		return value, nil
	}

	return "", contract.ErrSecretNotFound
}

func get(file, key string) (map[string]string, error) {

	data := map[string]string{}

	b, err := local.Get(file, key)
	if err != nil {
		return data, err
	}

	if err = yaml.Unmarshal(b, &data); err != nil {
		return data, err
	}

	return data, nil
}

func init() {
	v := FileVault{}
	vaults[v.Name()] = v
}
