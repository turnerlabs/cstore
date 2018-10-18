package prompt

import (
	"fmt"
	"strings"

	"github.com/turnerlabs/cstore/components/models"
)

// Confirm ...
func Confirm(description string, critical bool, io models.IO) bool {
	var s string

	if critical {
		fmt.Fprintf(io.UserOutput, "\n%s%s%s%s%s (y/N): ", redColor, bold, description, unbold, noColor)
	} else {
		fmt.Fprintf(io.UserOutput, "\n%s (y/N): ", description)
	}

	c, err := fmt.Fscanf(io.UserInput, "%s\n", &s)
	if c > 0 && err != nil {
		panic(err)
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return false
}
