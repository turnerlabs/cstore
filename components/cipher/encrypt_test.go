package cipher

import (
	"fmt"
	"testing"
)

func TestCipher(t *testing.T) {
	key := "AES256Key-32Characters1234567890"

	data, err := Encrypt(key, []byte("my data"))
	if err != nil {
		panic(err)
	}

	data, err = Decrypt(key, data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result: %s", string(data))
}
