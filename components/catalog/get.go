package catalog

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/turnerlabs/cstore/v4/components/cfg"
	"github.com/turnerlabs/cstore/v4/components/models"
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

	c := Catalog{}

	b, err := ioutil.ReadFile(fmt.Sprintf("%s%s", location(g.Location), catalogName))
	if err == nil {
		v := Version{}
		if err = yaml.Unmarshal(b, &v); err != nil {
			return c, err
		}

		if supported := supportedVersion(v); !supported {
			return c, fmt.Errorf("cStore %s is incompatible with %s catalog", cfg.Version, v.Version)
		}

		fc := FileCatalog{}
		if err = yaml.Unmarshal(b, &fc); err != nil {
			if !strings.Contains(err.Error(), "[]catalog.File") {
				return c, err
			}

			if err = yaml.Unmarshal(b, &c); err != nil {
				return c, err
			}
		} else {
			c = fc.ToBusiness(g.Location)
		}

		if c.Files == nil {
			c.Files = map[string]File{}
		}
	}

	return c, err
}

func supportedVersion(v Version) bool {
	for _, sv := range cfg.SupportedVersions {
		if strings.Contains(sv, v.Version) {
			return true
		}
	}

	return false
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
