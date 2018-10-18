package token

import "regexp"

// RemoveSecrets ...
func RemoveSecrets(b []byte) []byte {
	return regexp.MustCompile(`[:]{2}(.*?)}}`).ReplaceAll(b, []byte("}}"))
}
