package display

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// Error ...
func Error(err error, w io.Writer) {
	ErrorText(err.Error(), w)
}

// ErrorText ...
func ErrorText(text string, w io.Writer) {
	color.New(color.Bold, color.FgRed).Fprint(w, "\nERROR: ")
	fmt.Fprintln(w, text)
	fmt.Fprintln(w)
}
