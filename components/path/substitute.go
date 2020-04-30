package path

import (
	"os"
	"strings"
	"regexp"
)

// SubstituteTokens ...
func SubstituteTokens(path string) string {

	re := regexp.MustCompile(`\${([A-Z_]+)}`)

	matches := re.FindAllStringSubmatch(path, -1)

	for _, m := range matches {
		env, found := os.LookupEnv(m[1])

		if found {
			path = strings.Replace(path, m[0], env, -1)
		}
	}

	return path
} 