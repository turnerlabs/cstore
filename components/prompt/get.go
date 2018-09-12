package prompt

import (
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// GetValFromUser ...
func GetValFromUser(name, defaultValue, description string, hideInput bool) string {
	var s string

	if len(description) > 0 {
		fmt.Printf("\n%s\n", description)
	}

	if len(defaultValue) > 0 {
		fmt.Printf("%s (%s): ", name, defaultValue)
	} else {
		fmt.Printf("%s: ", name)
	}

	if hideInput {
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err == nil {
			s = string(password)
		}
		fmt.Println()
	} else {
		fmt.Scanf("%s\n", &s)
	}

	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return defaultValue
	}
	return s
}
