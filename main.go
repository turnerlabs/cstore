package main

import (
	"fmt"
	"os"

	"github.com/turnerlabs/cstore/cmd"
	"github.com/turnerlabs/cstore/components/cfg"
)

<<<<<<< HEAD
var version = "v1.0.0-rc"
=======
var version = "v0.5.0-beta"
>>>>>>> updating readme

func main() {
	cfg.Version = version

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
