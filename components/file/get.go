package file

import (
	"bytes"
	"fmt"
	"os"
)

// GetBy ...
func GetBy(path string) ([]byte, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return []byte{}, fmt.Errorf("Cannot find %s", path)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	return buf.Bytes(), err
}
