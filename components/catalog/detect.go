package catalog

import (
	"github.com/turnerlabs/cstore/components/cfg"
	yaml "gopkg.in/yaml.v2"
)

// IsOne ...
func IsOne(file []byte) bool {

	c := new(Catalog)

	if err := yaml.Unmarshal(file, c); err != nil {
		return false
	}

	if c.Version == cfg.Version[0:2] && len(c.Context) > 0 {
		return true
	}

	return false
}
