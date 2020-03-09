package catalog

import (
	"errors"

	"github.com/turnerlabs/cstore/v4/components/cfg"
	yaml "gopkg.in/yaml.v2"
)

// IsOne ...
func IsOne(file []byte) (bool, error) {

	c := new(Version)

	if err := yaml.Unmarshal(file, c); err != nil {
		return false, nil
	}

	if len(c.Context) > 0 {
		if c.Version == cfg.Version[0:2] {
			return true, nil
		}

		return false, errors.New("version mismatch")
	}

	return false, nil
}
