package catalog

import (
	"io/ioutil"

	uuid "github.com/satori/go.uuid"
	"github.com/turnerlabs/cstore/components/prompt"
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

	val := prompt.GetValFromUser("Context", getContext(), "The folder or context for the configuration files.", false)

	return Catalog{
		Version: v1,
		Context: val,
		Files:   map[string]File{},
	}
}
