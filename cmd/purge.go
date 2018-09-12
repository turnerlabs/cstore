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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/logger"
	"github.com/turnerlabs/cstore/components/store"
	"github.com/turnerlabs/cstore/components/vault"
)

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge file(s) remotely.",
	Long: `Purge file(s) remotely.

If all files were removed successfully, the local catalog is deleted.`,
	Run: func(cmd *cobra.Command, args []string) {
		clog, err := catalog.Get(viper.GetString(fileToken))
		if err != nil {
			logger.L.Fatal(err)
		}

		tags := getTags(tagList)

		files := clog.FilesBy(args, tags)

		count := 0
		purged := 0

		for key, value := range files {
			contextKey := clog.ContextKey(key)

			if value.IsRef {
				delete(clog.Files, key)
				continue
			}

			count++

			credsVaultName := value.Vaults.Credentials
			if len(viper.GetString(credsToken)) > 0 {
				credsVaultName = viper.GetString(credsToken)
			}

			cv, err := vault.GetBy(credsVaultName)
			if err != nil {
				logger.L.Printf("\nCould not purge %s!\n", value.Path)
				logger.L.Print(err)
				continue
			}

			encryptVaultName := value.Vaults.Encryption
			if len(viper.GetString(encryptToken)) > 0 {
				encryptVaultName = viper.GetString(encryptToken)
			}

			ev, err := vault.GetBy(encryptVaultName)
			if err != nil {
				logger.L.Printf("\nCould not purge %s!\n", value.Path)
				logger.L.Print(err)
				continue
			}

			st, err := store.Select(value, clog.Context, cv, ev, viper.GetBool(promptToken))
			if err != nil {
				logger.L.Printf("\nCould not purge %s!\n", value.Path)
				logger.L.Print(err)
				continue
			}

			if len(version) > 0 {

				versionedKey := fmt.Sprintf("%s/%s", contextKey, version)

				if err = st.Purge(versionedKey, value); err != nil {
					logger.L.Print(err)
				}

				for i, ver := range value.Versions {
					if ver == version {
						value.Versions = append(value.Versions[:i], value.Versions[i+1:]...)
					}
				}

				clog.Files[key] = value
			} else {

				for _, version := range value.Versions {
					versionedKey := fmt.Sprintf("%s/%s", contextKey, version)

					if err = st.Purge(versionedKey, value); err != nil {
						logger.L.Print(err)
					}
				}

				if err = st.Purge(contextKey, value); err == nil {
					delete(clog.Files, key)
					if err := clog.FilePurged(key); err != nil {
						logger.L.Print(err)
					}
					purged++
				} else {
					logger.L.Printf("\nCould not purge %s!\n", value.Path)
					logger.L.Print(err)
					continue
				}
			}
		}

		if len(clog.Files) == 0 {
			if err = catalog.Remove(viper.GetString(fileToken)); err != nil {
				logger.L.Fatal(err)
			}
		} else {
			if err = catalog.Write(viper.GetString(fileToken), clog); err != nil {
				logger.L.Fatal(err)
			}
		}

		logger.L.Printf("%d of %d file(s) purged from remote storage. \n", purged, count)
	},
}

func init() {
	RootCmd.AddCommand(purgeCmd)

	purgeCmd.Flags().StringVarP(&tagList, "tags", "t", "", "Specify a list of tags used to filter files.")
	purgeCmd.Flags().StringVarP(&version, "ver", "v", "", "Remove specific version.")
}
