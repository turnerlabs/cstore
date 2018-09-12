package catalog

import (
	"io/ioutil"

	uuid "github.com/satori/go.uuid"
	yaml "gopkg.in/yaml.v2"
)

const v1 = "v1"

// GetMake loads data from existing catalog when
// available or returns a new empty catalog object.
func GetMake(catalogName string) (Catalog, error) {

	c, err := Get(catalogName)
	if err != nil {
		c = create()
	}

	return c, nil
}

// Get ...
func Get(catalogName string) (Catalog, error) {
	c := Catalog{}

	b, err := ioutil.ReadFile(catalogName)
	if err == nil {
		if err = yaml.Unmarshal(b, &c); err != nil {
			return c, err
		}

		if c.Files == nil {
			c.Files = map[string]File{}
		}
	}

	return c, err
}

func getContext() string {
	return uuid.NewV4().String()
}

func create() Catalog {

	return Catalog{
		Version: v1,
		Context: getContext(),
		Files:   map[string]File{},
	}
}
