package vault

import (
	"os/user"

	keychain "github.com/keybase/go-keychain"
	"github.com/turnerlabs/cstore/components/prompt"
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

// Set ...
func (v KeychainVault) Set(contextID, key, value string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	if err = setValueInKeychain(u.Username, key, value); err != nil {

		if err == keychain.ErrorDuplicateItem {
			item := keychain.NewItem()
			item.SetSecClass(keychain.SecClassGenericPassword)
			item.SetService(key)
			item.SetAccount(u.Username)

			if err = keychain.DeleteItem(item); err != nil {
				return err
			}

			if err = setValueInKeychain(u.Username, key, value); err != nil {
				return err
			}

			return nil
		}

		return err
	}

	return nil
}

// Get ...
func (v KeychainVault) Get(contextID, key, defaultVal, description string, askUser bool) (string, error) {

	u, err := user.Current()
	if err != nil {
		return "", err
	}

	val, err := getFromKeychain(u.Username, key)
	if err != nil {
		if err == ErrSecretNotFound && askUser {
			val = prompt.GetValFromUser(key, defaultVal, description, true)

			if len(val) == 0 {
				return "", err
			}

			if err := setValueInKeychain(u.Username, key, val); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return val, nil
}

// Delete ...
func (v KeychainVault) Delete(contextID, key string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	if err := deleteKeyInKeychain(u.Username, key); err != nil {
		if err == keychain.ErrorItemNotFound {
			return ErrSecretNotFound
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
		return "", ErrSecretNotFound
	} else {
		return string(results[0].Data), nil
	}
}

func init() {
	v := KeychainVault{}
	vaults[v.Name()] = v
}
