package vault

import (
	"fmt"
	"os/user"

	keychain "github.com/keybase/go-keychain"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/models"
)

const accessGroup = "cstore"

// KeychainVault ...
type KeychainVault struct{}

// Name ...
func (v KeychainVault) Name() string {
	return "osx-keychain"
}

// Description ...
func (v KeychainVault) Description() string {
	return `This vault retrieves secrets stored as passwords from the OSX Keychain app. This allows a store to retrieve securely stored values like encryption keys and passwords. This vault is only accessible on OSX.`
}

// BuildKey ...
func (v KeychainVault) BuildKey(contextID, group, prop string) string {
	if len(prop) > 0 {
		return fmt.Sprintf("%s: %s", group, prop)
	}

	return group
}

// Pre ...
func (v KeychainVault) Pre(clog catalog.Catalog, fileEntry *catalog.File, access contract.IVault, uo cfg.UserOptions, io models.IO) error {
	return nil
}

// Set ...
func (v KeychainVault) Set(contextID, group, prop, value string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	if err = setValueInKeychain(u.Username, v.BuildKey(contextID, group, prop), value); err != nil {

		if err == keychain.ErrorDuplicateItem {
			item := keychain.NewItem()
			item.SetSecClass(keychain.SecClassGenericPassword)
			item.SetService(v.BuildKey(contextID, group, prop))
			item.SetAccount(u.Username)

			if err = keychain.DeleteItem(item); err != nil {
				return err
			}

			return setValueInKeychain(u.Username, v.BuildKey(contextID, group, prop), value)
		}

		return err
	}

	return nil
}

// Get ...
func (v KeychainVault) Get(contextID, group, prop string) (string, error) {

	u, err := user.Current()
	if err != nil {
		return "", err
	}

	return getFromKeychain(u.Username, v.BuildKey(contextID, group, prop))
}

// Delete ...
func (v KeychainVault) Delete(contextID, group, prop string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	if err := deleteKeyInKeychain(u.Username, v.BuildKey(contextID, group, prop)); err != nil {
		if err == keychain.ErrorItemNotFound {
			return contract.ErrSecretNotFound
		}

		return err
	}

	return nil
}

func setValueInKeychain(user, key, value string) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(key)
	item.SetAccount(user)
	item.SetAccessGroup(accessGroup)
	item.SetData([]byte(value))
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	return keychain.AddItem(item)
}

func deleteKeyInKeychain(user, key string) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(key)
	item.SetAccount(user)
	item.SetAccessGroup(accessGroup)
	return keychain.DeleteItem(item)
}

func getFromKeychain(username, key string) (string, error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(key)
	query.SetAccount(username)
	query.SetAccessGroup(accessGroup)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)
	results, err := keychain.QueryItem(query)

	if err != nil {
		return "", err
	} else if len(results) != 1 {
		return "", contract.ErrSecretNotFound
	} else {
		return string(results[0].Data), nil
	}
}

func init() {
	v := KeychainVault{}
	vaults[v.Name()] = v
}
