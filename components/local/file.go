package local

import (
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/turnerlabs/cstore/v4/components/cipher"
	"github.com/turnerlabs/cstore/v4/components/file"
)

// BuildPath ...
func BuildPath(name string) string {
	const path = ".cstore"

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s/%s/%s", home, path, name)
}

// Update ...
func Update(name, key string, data []byte) error {

	var err error
	if len(key) > 0 {
		data, err = cipher.Encrypt(key, data)
		if err != nil {
			return err
		}
	}

	return file.Save(BuildPath(name), data)

}

// Missing ...
func Missing(name string) bool {
	_, err := os.Stat(BuildPath(name))

	return os.IsNotExist(err)
}

// Get ...
func Get(name, key string) ([]byte, error) {

	data, err := file.GetBy(BuildPath(name))
	if err != nil {
		return nil, err
	}

	if len(key) > 0 {
		data, err = cipher.Decrypt(key, data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}
