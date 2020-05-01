package token

import (
	"os"
	"regexp"
	"strings"
)

// Substitute ...
func Substitute(value string) string {

	re := regexp.MustCompile(`\${([A-Z0-9_]+)}`)

	matches := re.FindAllStringSubmatch(value, -1)

	for _, m := range matches {
		env, found := os.LookupEnv(m[1])

		if found {
			value = strings.Replace(value, m[0], env, -1)
		}
	}

	return value
}
