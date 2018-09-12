package main

import (
	"fmt"
	"os"

	"github.com/turnerlabs/cstore/cmd"
	"github.com/turnerlabs/cstore/components/cfg"
)

var version = "v0.4.0-beta"

func main() {
	cfg.Version = version

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
