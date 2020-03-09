package prompt

import (
	"fmt"
	"strings"

	"github.com/turnerlabs/cstore/v4/components/models"
)

const (
	Warn   = "WARN"
	Danger = "DANGER"
	Normal = "NORMAL"
)

// Confirm ...
func Confirm(description, level string, io models.IO) bool {
	var s string

	switch level {
	case Warn:
		fmt.Fprintf(io.UserOutput, "\n%s%s%s%s%s (y/N): ", yellowColor, bold, description, unbold, noColor)
	case Danger:
		fmt.Fprintf(io.UserOutput, "\n%s%s%s%s%s (y/N): ", redColor, bold, description, unbold, noColor)
	default:
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
