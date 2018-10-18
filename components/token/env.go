package token

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
)

func formatEnv(env string) string {
	return strings.Replace(env, "_", "-", -1)
}

func replaceENV(b []byte, tokens map[string]Token) ([]byte, error) {
	var envRegex = regexp.MustCompile(envRegexStr)

	nb := []byte{}
	for _, line := range bytes.Split(b, []byte{'\n'}) {

		lineMatches := envRegex.FindSubmatch(line)
		for _, token := range tokens {
			if lineMatches != nil && len(lineMatches) == 2 && strings.ToLower(string(lineMatches[1])) == token.EnvVar {
				line = bytes.Replace(line, []byte(token.Formatted()), []byte(token.Value), -1)
			}
		}

		nb = append(nb, line...)
		nb = append(nb, []byte("\n")...)
	}

	return nb, nil
}

func searchENV(b []byte, forValues bool) (tokens map[string]Token, err error) {
	tokens = map[string]Token{}

	lines := bytes.Split(b, []byte{'\n'})

	var tokenRegex = regexp.MustCompile(getTokenRegex(forValues))
	var envRegex = regexp.MustCompile(envRegexStr)

	for _, l := range lines {

		byteTokens := tokenRegex.FindAllSubmatch(l, -1)
		byteEnv := envRegex.FindSubmatch(l)

		if byteTokens == nil || byteEnv == nil {
			continue
		}

		for _, bt := range byteTokens {

			ss := strings.Split(string(bt[2]), "/")
			nss := ss[:len(ss)-1]

			t := Token{
				EnvVar: strings.ToLower(string(byteEnv[1])),
				Env:    strings.Join(nss, "/"),
				Prop:   strings.ToLower(ss[len(ss)-1]),
				Value:  notFound,
			}

			if len(bt) == 4 {
				t.Value = string(bt[3])
			}

			tokens[t.String()] = t
		}
	}

	return tokens, nil
}

// Build ...
func Build(secret string, tokens map[string]Token) ([]byte, error) {
	props := map[string]string{}

	for _, v := range tokens {

		if v.Value == notFound {
			props[v.Prop] = ""
		} else {
			props[v.Prop] = v.Value
		}
	}

	return json.Marshal(props)
}
