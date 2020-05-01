package setting

import (
	"fmt"
	"os"

	"github.com/turnerlabs/cstore/v4/components/contract"
	"github.com/turnerlabs/cstore/v4/components/models"
	"github.com/turnerlabs/cstore/v4/components/prompt"
	"github.com/turnerlabs/cstore/v4/components/token"
)

type IKeyValueStore interface {
	Name() string

	Get(contextID, group, prop string) (string, error)
	Set(contextID, group, prop, value string) error
	Delete(contextID, group, prop string) error

	BuildKey(contextID, group, prop string) string
}

var didPrompt = map[string]bool{}

// Setting ...
type Setting struct {
	Group string
	Prop  string

	DefaultValue string
	Description  string

	Prompt     bool
	Silent     bool
	HideInput  bool
	AutoSave   bool
	PromptOnce bool

	Vault IKeyValueStore
}

// Key ...
func (s Setting) Key(context string) string {
	return s.Vault.BuildKey(context, s.Group, s.Prop)
}

// Value ...
type Value struct {
	Value  string
	Actual string
}

// Set ...
func (v *Value) Set(value string) {
	v.Value = value
	v.Actual = token.Substitute(value)
}

// Get ...
func (s Setting) Get(context string, io models.IO) (Value, error) {
	v := Value{}

	value, err := s.Vault.Get(context, s.Group, s.Prop)
	if err != nil {
		if err.Error() == contract.ErrSecretNotFound.Error() {
			s.Prompt = true
		} else {
			return v, err
		}
	}

	v.Set(value)

	if s.Prompt && !s.Silent && !didPrompt[s.Prop] {

		if s.PromptOnce {
			didPrompt[s.Prop] = true
		}

		formattedKey := s.Vault.BuildKey(context, s.Group, s.Prop)

		opt := prompt.Options{
			DefaultValue: s.DefaultValue,
			Description:  s.Description,
			HideInput:    s.HideInput,
		}

		if env := os.Getenv(formattedKey); len(env) > 0 {
			opt.DefaultValue = env
		} else if len(value) > 0 {
			opt.DefaultValue = value
		}

		v.Set(prompt.GetValFromUser(formattedKey, opt, io))

		if s.AutoSave || prompt.Confirm(fmt.Sprintf("Save %s preference in %s?", formattedKey, s.Vault.Name()), prompt.Warn, io) {
			if err := s.Vault.Set(context, s.Group, s.Prop, v.Value); err != nil {
				return v, err
			}
		}

	}

	return v, nil
}
