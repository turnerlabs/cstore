package vault

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/turnerlabs/cstore/components/cipher"
	"github.com/turnerlabs/cstore/components/configure"
	"github.com/turnerlabs/cstore/components/prompt"
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
	return fmt.Sprintf(`This vault uses an encrypted file to store and get credentails and/or encryption keys. 
	
defaults

 -location: %s
 -key:      %s

The key can be shared by placing it into the same folder on another machine to allow access to the vault data.

`, configure.BuildPath(fileName), configure.BuildPath(fileKeyName))
}

// Set ...
func (v FileVault) Set(contextID, key, value string) error {
	return set(contextID, key, value)
}

func getEncryptionKey() (string, error) {

	b, err := configure.Get(fileKeyName, "")
	if err != nil {
		return cipher.GenerateAES256Key(), err
	}
	return string(b), nil
}

func saveEncryptionKey(eKey string) error {
	return configure.Update(fileKeyName, "", []byte(eKey))
}

func set(contextID, key, value string) error {

	eKey, _ := getEncryptionKey()
	data, _ := get(fileName, eKey)

	data[key] = value

	return create(eKey, data)
}

func create(eKey string, data map[string]string) error {
	d, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	if err = configure.Update(fileName, eKey, d); err != nil {
		return err
	}

	return saveEncryptionKey(eKey)
}

// Delete ...
func (v FileVault) Delete(contextID, key string) error {

	if configure.Missing(fileName) {
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

	delete(data, key)

	d, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return configure.Update(fileName, eKey, d)
}

// Get ...
func (v FileVault) Get(contextID, key, defaultVal, description string, askUser bool) (string, error) {

	found := false
	value := ""

	if !configure.Missing(fileName) {
		eKey, _ := getEncryptionKey()

		data, err := get(fileName, eKey)
		if err != nil {
			return "", err
		}

		value, found = data[key]
	}

	if !found && askUser {
		if val := prompt.GetValFromUser(key, defaultVal, description, true); len(val) > 0 {
			set(contextID, key, val)

			return val, nil
		}

		return "", ErrSecretNotFound
	}

	return value, nil
}

func get(file, key string) (map[string]string, error) {

	data := map[string]string{}

	b, err := configure.Get(file, key)
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
