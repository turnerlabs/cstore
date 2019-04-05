package env

import (
	"bytes"
	"fmt"
	"os"

	"github.com/subosito/gotenv"
)

// DiffCurrent ...
func DiffCurrent(file []byte) []byte {
	newFile := []byte{}

	environment := gotenv.Parse(bytes.NewReader(file))

	for key, value := range environment {
		if _, exists := os.LookupEnv(key); !exists {
			newFile = append(newFile, []byte(fmt.Sprintf("%s=%s\n", key, value))...)
		}
	}

	return newFile
}
