package file

import (
	"io/ioutil"
	"os"
	"strings"
)

// Save ...
func Save(path string, b []byte) error {
	if strings.Contains(path, "/") {
		os.MkdirAll(strings.TrimRightFunc(path, notForwardSlash), os.ModePerm)
	}

	if err := ioutil.WriteFile(path, b, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func notForwardSlash(char rune) bool {
	if char != '/' {
		return true
	}
	return false
}
