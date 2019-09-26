package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/vault"
)

// vaultsCmd represents the vaults command
var vaultsCmd = &cobra.Command{
	Use:   "vaults",
	Short: "List available vaults or vault details.",
	Long:  `List available vaults or vault details.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 1 {
			if v, found := vault.Get()[args[0]]; found {
				fmt.Fprintf(os.Stderr, "%s\n", v.Description())
			} else {
				fmt.Fprintln(os.Stderr, "Vault not found.")
			}
		} else {
			fmt.Fprintf(ioStreams.UserOutput, "\nVaults are used to store and retrieve secrets that are needed for store access or in file contents. During a push or pull, any vault can override the default vaults by using the '-c' and '-x' cli flags for store access and secret injection respectively.\n\n")

			fmt.Fprintf(ioStreams.UserOutput, "Use 'cstore vaults VAULT_NAME' cmd for details.\n")
			for _, v := range vault.Get() {
				fmt.Print("|-")
				color.New(color.FgBlue).Fprintf(ioStreams.UserOutput, "%s\n", v.Name())
			}

			fmt.Fprintln(ioStreams.UserOutput)
		}
	},
}

func init() {
	RootCmd.AddCommand(vaultsCmd)
}
