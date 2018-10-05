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
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List file(s) stored remotely.",
	Long:  `List file(s) stored remotely.`,
	Run: func(cmd *cobra.Command, args []string) {

		tags := getTags(tagList)

		total, err := listFilesFor(viper.GetString(fileToken), viper.GetString(credsToken), viper.GetString(encryptToken), tags)
		if err != nil {
			logger.L.Fatal(err)
		}

		logger.L.Printf("%d file(s) stored remotely.\n", total)
	},
}

func listFilesFor(cRef, cVault, eVault string, tags []string) (int, error) {
	path := getPath(cRef)

	clog, err := catalog.Get(cRef)
	if err != nil {
		return 0, err
	}

	files := catalog.FilterByTag(clog.Files, tags, false)

	total := 0
	for _, fileInfo := range files {
		if fileInfo.IsRef {
			fullPath := fileInfo.Path

			if len(path) > 0 {
				fullPath = fmt.Sprintf("%s/%s", path, fileInfo.Path)
			}

			count, err := listFilesFor(fullPath, cVault, eVault, tags)
			if err != nil {
				return 0, err
			}

			total += count
		} else {
			if len(path) > 0 {
				logger.L.Printf(" - %s/%s %s\n", path, fileInfo.Path, formatEncryptionFlag(fileInfo.Encrypted))
			} else {
				logger.L.Printf(" - %s %s\n", fileInfo.Path, formatEncryptionFlag(fileInfo.Encrypted))
			}

			for _, ver := range fileInfo.Versions {
				logger.L.Printf("  + %s\n", ver)
			}

			total++
		}
	}

	return total, nil
}

func formatEncryptionFlag(encrypted bool) string {
	if encrypted {
		return "(encrypted)"
	}

	return ""
}

func init() {
	RootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	listCmd.Flags().StringVarP(&tagList, "tags", "t", "", "Specify a list of tags used to filter files.")
}
