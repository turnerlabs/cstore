package catalog

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/models"
	yaml "gopkg.in/yaml.v2"
)

// GetMake loads data from existing catalog when
// available or returns a new empty catalog object.
func GetMake(catalogName string, io models.IO) (Catalog, error) {

	c, err := Get(catalogName)
	if err != nil {
		c = create(io)
	}

	return c, nil
}

// Get ...
func Get(catalogName string) (Catalog, error) {
	g, _ := GetGhost()

	c := Catalog{
		CWD: g.Location,
	}

	b, err := ioutil.ReadFile(fmt.Sprintf("%s%s", c.Location(), catalogName))
	if err == nil {
		if err = yaml.Unmarshal(b, &c); err != nil {
			return c, err
		}

		if !strings.Contains(cfg.Version, c.Version) {
			return c, fmt.Errorf("cStore %s is incompatible with a %s catalog", cfg.Version, c.Version)
		}

		if c.Files == nil {
			c.Files = map[string]File{}
		}
	}

	return c, err
}

func getContext() string {
	wd, err := os.Getwd()
	if err != nil {
		return uuid.NewV4().String()
	}

	directories := strings.Split(wd, "/")

	if len(directories) < 1 {
		return uuid.NewV4().String()
	}

	return directories[len(directories)-1]
}
