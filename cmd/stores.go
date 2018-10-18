package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/store"
)

// storesCmd represents the stores command
var storesCmd = &cobra.Command{
	Use:   "stores",
	Short: "List available stores and details.",
	Long:  `List available stores and details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			if s, found := store.Get()[args[0]]; found {
				fmt.Fprintf(os.Stderr, "%s\n", s.Description())
			} else {
				fmt.Fprintln(os.Stderr, "Store not found.")
			}
		} else {
			fmt.Fprintf(ioStreams.UserOutput, "\nStores are used to remotely store configuration or other files. During a push, any store can override the default store by using the '-s' cli flag.\n\n")

			fmt.Fprintf(ioStreams.UserOutput, "Use 'cstore stores STORE_NAME' cmd for details.\n")

			for _, store := range store.Get() {
				fmt.Fprintf(os.Stderr, "|-%s%s%s\n", uo.Format.Blue, store.Name(), uo.Format.NoColor)
			}

			fmt.Fprintln(ioStreams.UserOutput)
		}
	},
}

func init() {
	RootCmd.AddCommand(storesCmd)
}
