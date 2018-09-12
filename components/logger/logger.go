package logger

import (
	"log"
	"os"
)

// L ...
var L = log.New(os.Stderr, "", 0)
