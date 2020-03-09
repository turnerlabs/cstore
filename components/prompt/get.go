package prompt

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/turnerlabs/cstore/v4/components/models"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	bold   = "\033[1m"
	unbold = "\033[0m"

	redColor    = "\033[0;31m"
	yellowColor = "\033[0;33m"
	blueColor   = "\033[0;34m"
	noColor     = "\033[0m"
)

// Options ...
type Options struct {
	Description  string
	DefaultValue string
	HideInput    bool
}

// GetValFromUser ...
func GetValFromUser(name string, v Options, io models.IO) string {
	var s string

	fmt.Fprintln(io.UserOutput)

	if len(v.Description) > 0 {
		fmt.Fprintf(io.UserOutput, "%s\n\n", v.Description)
	}

	if len(v.DefaultValue) > 0 {
		fmt.Fprintf(io.UserOutput, "%sDefault:%s %s\n", bold, unbold, v.DefaultValue)
	}

	fmt.Fprintf(io.UserOutput, "%s%s:%s ", bold, name, unbold)

	if v.HideInput {
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err == nil {
			s = string(password)
		}
	} else {
		c, err := fmt.Fscanf(io.UserInput, "%s\n", &s)
		if c > 0 && err != nil {
			panic(err)
		}
	}

	fmt.Fprintln(io.UserOutput)

	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return v.DefaultValue
	}

	return s
}
