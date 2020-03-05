package main // github.com/turnerlabs/cstore/v4

import (
	"fmt"
	"os"

	"github.com/turnerlabs/cstore/v4/cli/cmd"
	"github.com/turnerlabs/cstore/v4/components/cfg"
)

var version = "v3.0.0-rc"

func main() {
	cfg.Version = version

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
