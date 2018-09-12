package catalog

import yaml "gopkg.in/yaml.v2"

// IsOne ...
func IsOne(file []byte) bool {

	c := new(Catalog)

	if err := yaml.Unmarshal(file, c); err != nil {
		return false
	}

	if c.Version == v1 && len(c.Context) > 0 {
		return true
	}

	return false
}
