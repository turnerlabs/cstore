package display

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// Error ...
func Error(text string, w io.Writer) {
	color.New(color.Bold, color.FgRed).Fprint(w, "\nERROR: ")
	fmt.Fprintln(w, text)
}
