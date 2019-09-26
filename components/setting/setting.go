package setting

import (
	"fmt"
	"os"

	"github.com/turnerlabs/cstore/components/contract"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/prompt"
)

// Setting ...
type Setting struct {
	Group string
	Prop  string

	DefaultValue string
	Description  string

	Prompt    bool
	Silent    bool
	HideInput bool
	AutoSave  bool

	Vault contract.IVault
}

// Key ...
func (s Setting) Key(context string) string {
	return s.Vault.BuildKey(context, s.Group, s.Prop)
}

// Get ...
func (s Setting) Get(context string, io models.IO) (string, error) {
	value, err := s.Vault.Get(context, s.Group, s.Prop)
	if err != nil {
		if err.Error() == contract.ErrSecretNotFound.Error() {
			s.Prompt = true
		} else {
			return value, err
		}
	}

	if s.Prompt && !s.Silent {
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

		value = prompt.GetValFromUser(formattedKey, opt, io)

		if s.AutoSave || prompt.Confirm(fmt.Sprintf("Save %s preference in %s?", formattedKey, s.Vault.Name()), prompt.Warn, io) {
			if err := s.Vault.Set(context, s.Group, s.Prop, value); err != nil {
				return value, err
			}
		}

	}

	return value, nil
}
