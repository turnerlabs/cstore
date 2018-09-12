package file

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// Get ...
func Get(path string) ([]byte, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return []byte{}, fmt.Errorf("Cannot find %s", path)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	return buf.Bytes(), err
}

// IsEnv ...
func IsEnv(path string) bool {
	if strings.HasSuffix(path, ".env") {
		return true
	}
	return false
}
