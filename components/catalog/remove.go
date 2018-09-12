package catalog

import "os"

// Remove ...
func Remove(path string) error {
	return os.Remove(path)
}
