package token

import (
	"fmt"
	"regexp"
	"strings"
)

const notFound = "[NOT_FOUND]"
const sep = "::"

// Find ...
func Find(b []byte) map[string]string {
	tokens := map[string]string{}

	var regex = regexp.MustCompile(`{{(.*)}}`)

	byteTokens := regex.FindAllSubmatch(b, -1)
	if byteTokens == nil {
		return map[string]string{}
	}

	for _, bt := range byteTokens {
		tokens[string(bt[0])] = notFound
	}

	return tokens
}

// Build ...
func Build(secret string, tokens map[string]string) string {
	props := map[string]string{}

	for k, v := range tokens {
		secretProp := strings.Split(strings.Trim(k, "{}"), sep)
		if secretProp[0] == secret {
			if len(secretProp) == 2 {
				if v == notFound {
					props[secretProp[1]] = ""
				} else {
					props[secretProp[1]] = v
				}

			}
		}
	}

	ex := "{"
	for k, v := range props {
		ex = fmt.Sprintf(`%s"%s":"%s",`, ex, k, v)
	}

	ex = strings.TrimRight(ex, ",")

	return fmt.Sprintf("%s}", ex)
}
