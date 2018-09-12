// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/logger"
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
				logger.L.Printf("%s\n", v.Description())
			} else {
				logger.L.Println("Vault not found.")
			}
		} else {
			for _, v := range vault.Get() {
				logger.L.Printf(" - %s\n", v.Name())
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(vaultsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// vaultsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// vaultsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
