package token

import (
	"fmt"
)

const (
	notFound = "[NOT_FOUND]"

	envRegexStr        = `^([\w\d]+)=`
	tokenRegexStr      = `{{(([\w\d\/-]+))}}`
	tokenValueRegexStr = `{{(([\w\d\/-]+)[:]{2}(.*?))}}`

	envFileExt  = "env"
	jsonFileExt = "json"
)

// Token ...
type Token struct {
	Env    string
	EnvVar string
	Prop   string
	Value  string
}

// Formatted ...
func (t Token) Formatted() string {
	return fmt.Sprintf("{{%s/%s}}", t.Env, t.Prop)
}

// GetValue ...
func (t Token) GetValue(formatted bool) string {
	if formatted {
		return fmt.Sprintf("{{%s/%s::%s}}", t.Env, t.Prop, t.Value)
	}

	return t.Value
}

// Secret ...
func (t Token) Secret() string {
	if len(t.Env) > 0 {
		return fmt.Sprintf("%s/%s", t.Env, formatEnv(t.EnvVar))
	}

	return formatEnv(t.EnvVar)
}

// String ...
func (t Token) String() string {
	if len(t.Env) > 0 {
		return fmt.Sprintf("%s/%s/%s", t.Env, formatEnv(t.EnvVar), t.Prop)
	}

	return fmt.Sprintf("%s/%s", formatEnv(t.EnvVar), t.Prop)
}

// Equals ...
func (t Token) Equals(t2 Token) bool {
	return t.String() == t2.String()
}

// Find ...
func Find(b []byte, ext string, withValues bool) (tokens map[string]Token, err error) {
	switch ext {
	case envFileExt:
		return searchENV(b, withValues)
	case jsonFileExt:
		return searchJSON(b, withValues)
	}

	return map[string]Token{}, fmt.Errorf("unsupported file extension %s", ext)
}

// Replace ...
func Replace(b []byte, ext string, tokens map[string]Token, formattedValue bool) ([]byte, error) {
	switch ext {
	case envFileExt:
		return replaceENV(b, tokens, formattedValue)
	case jsonFileExt:
		return replaceJSON(b, tokens, formattedValue)
	}

	return []byte{}, fmt.Errorf("unsupported file extension %s", ext)
}

func getTokenRegex(value bool) string {
	if value {
		return tokenValueRegexStr
	}
	return tokenRegexStr
}
