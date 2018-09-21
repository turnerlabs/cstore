package token

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"regexp"
)

const notFound = "[NOT_FOUND]"
const envRegexStr = `^([\w\d]+)=`

// Token ...
type Token struct {
	ContextID string
	Env       string
	EnvVar    string
	Prop      string
	Value     string
}

// Formatted ...
func (t Token) Formatted() string {
	return fmt.Sprintf("{{%s/%s}}", t.Env, t.Prop)
}

// Secret ...
func (t Token) Secret() string {
	if len(t.Env) > 0 {
		return fmt.Sprintf("%s/%s/%s", t.ContextID, t.Env, formatEnv(t.EnvVar))
	}

	return fmt.Sprintf("%s/%s", t.ContextID, formatEnv(t.EnvVar))
}

// String ...
func (t Token) String() string {
	if len(t.Env) > 0 {
		return fmt.Sprintf("%s/%s/%s/%s", t.ContextID, t.Env, formatEnv(t.EnvVar), t.Prop)
	}

	return fmt.Sprintf("%s/%s/%s", t.ContextID, formatEnv(t.EnvVar), t.Prop)
}

func formatEnv(env string) string {
	return strings.Replace(env, "_", "-", -1)
}

// Equals ...
func (t Token) Equals(t2 Token) bool {
	return t.String() == t2.String()
}

// Clean ...
func Clean(b []byte) []byte {
	return regexp.MustCompile(`[:]{2}(.*?)}}`).ReplaceAll(b, []byte("}}"))
}

// Replace ...
func Replace(b []byte, tokens map[string]Token) []byte {
	var envRegex = regexp.MustCompile(envRegexStr)

	nb := []byte{}
	for _, l := range bytes.Split(b, []byte{'\n'}) {

		byteEnv := envRegex.FindSubmatch(l)
		for _, v := range tokens {
			if byteEnv != nil && len(byteEnv) == 2 && strings.ToLower(string(byteEnv[1])) == v.EnvVar {
				l = bytes.Replace(l, []byte(v.Formatted()), []byte(v.Value), -1)
			}
		}

		nb = append(nb, l...)
		nb = append(nb, []byte("\n")...)
	}

	return nb
}

// Find ...
func Find(b []byte, contextID string, withValues bool) (tokens map[string]Token) {
	tokens = map[string]Token{}

	lines := bytes.Split(b, []byte{'\n'})

	var tokenRegexStr = `{{(([\w\d\/-]+))}}`
	if withValues {
		tokenRegexStr = `{{(([\w\d\/-]+)[:]{2}(.*?))}}`
	}

	var tokenRegex = regexp.MustCompile(tokenRegexStr)
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
				ContextID: contextID,
				EnvVar:    strings.ToLower(string(byteEnv[1])),
				Env:       strings.Join(nss, "/"),
				Prop:      strings.ToLower(ss[len(ss)-1]),
				Value:     notFound,
			}

			if len(bt) == 4 {
				t.Value = string(bt[3])
			}

			tokens[t.String()] = t
		}
	}

	//os.Exit(0)

	return tokens
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
