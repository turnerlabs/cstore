package models

import "io"

// IO ...
type IO struct {
	UserOutput io.Writer
	UserInput  io.Reader
	Export     io.Writer
}
