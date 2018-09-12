package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	mrand "math/rand"
	"time"
)

// AESKeyName is the environment variable that holds the encryption key. It
// must be 16 or 32 characters log to meet key requirements for 128 or 256
// bit encryption respectively.
var AESKeyName = "CSTORE_AES_KEY"

// Decrypt ...
func Decrypt(k string, cipherData []byte) ([]byte, error) {
	key := []byte(k)
	//ciphertext, _ := hex.DecodeString("22277966616d9bc47177bd02603d08c9a67d5380d0fe8cf3b44438dff7b9")
	//cipherText := string(cipherData)

	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(cipherData) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := cipherData[:aes.BlockSize]
	cipherData = cipherData[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherData, cipherData)

	return cipherData, nil
}

// Encrypt ...
func Encrypt(k string, plainData []byte) ([]byte, error) {
	key := []byte(k)

	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	cipherData := make([]byte, aes.BlockSize+len(plainData))
	iv := cipherData[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return []byte{}, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherData[aes.BlockSize:], plainData)

	return cipherData, nil
}

// GenerateAES256Key ...
func GenerateAES256Key() string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
		"1234567890" +
		"!@#$%^&*()_+=-~<>?/.,:;"

	return stringWithCharset(32, charset)
}

var seededRand = mrand.New(mrand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
