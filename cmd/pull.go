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
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/file"
	"github.com/turnerlabs/cstore/components/logger"
	"github.com/turnerlabs/cstore/components/store"
	"github.com/turnerlabs/cstore/components/token"
	"github.com/turnerlabs/cstore/components/vault"
)

var exportEnv bool
var secrets bool

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Restore file(s) on file system.",
	Long:  `Restore file(s) on file system.`,
	Run: func(cmd *cobra.Command, args []string) {

		tags := getTags(tagList)

		count, total, err := pull(viper.GetString(fileToken), viper.GetString(credsToken), viper.GetString(encryptToken), args, tags)
		if err != nil {
			logger.L.Fatal(err)
		}

		logger.L.Printf("%d of %d file(s) restored on file system.\n", count, total)
	},
}

func pull(catalogPath, cVault, eVault string, args []string, tags []string) (int, int, error) {

	root := getPath(catalogPath)

	clog, err := catalog.Get(catalogPath)
	if err != nil {
		return 0, 0, err
	}

	files := clog.FilesBy(args, tags)

	restoredCount := 0
	totalCount := 0

	for fileKey, fileInfo := range files {
		credsVaultName := fileInfo.Vaults.Credentials
		if len(viper.GetString(credsToken)) > 0 {
			credsVaultName = viper.GetString(credsToken)
		}

		cv, err := vault.GetBy(credsVaultName)
		if err != nil {
			logger.L.Printf("\nCould not restore %s!\n", fileInfo.Path)
			logger.L.Print(err)
			continue
		}

		encryptVaultName := fileInfo.Vaults.Encryption
		if len(viper.GetString(encryptToken)) > 0 {
			encryptVaultName = viper.GetString(encryptToken)
		}

		ev, err := vault.GetBy(encryptVaultName)
		if err != nil {
			logger.L.Printf("\nCould not restore %s!\n", fileInfo.Path)
			logger.L.Print(err)
			continue
		}

		contextKey := clog.ContextKey(fileKey)

		fullPath := buildPath(root, fileInfo.Path)

		if fileInfo.IsRef {
			c, t, err := pull(fullPath, cVault, eVault, args, tags)
			if err != nil {
				return 0, 0, err
			}

			restoredCount += c
			totalCount += t
			continue
		}

		totalCount++

		st, err := store.Select(fileInfo, clog.Context, cv, ev, viper.GetBool(promptToken))
		if err != nil {
			return 0, 0, err
		}

		if len(version) > 0 {
			contextKey = fmt.Sprintf("%s/%s", contextKey, version)
		}

		b, attr, err := st.Pull(contextKey, fileInfo)
		if err != nil {
			logger.L.Printf("\nCould not restore %s!\n", fileInfo.Path)
			logger.L.Print(err)
			continue
		}

		bt := b
		if secrets {
			tokens := token.Find(b)

			tokens, err = st.GetTokens(tokens)
			if err != nil {
				logger.L.Printf("\nCould not get tokens for %s!\n", fileInfo.Path)
				logger.L.Print(err)
				continue
			}

			for t, v := range tokens {
				bt = bytes.Replace(bt, []byte(t), []byte(v), -1)
			}
		}

		clog.FilePulled(fileKey, version, attr.LastModified)

		if exportEnv && fileInfo.IsEnv {
			script := bufferExportScript(bt)

			if _, err := script.WriteTo(os.Stdout); err != nil {
				return 0, 0, err
			}

			logger.L.Printf("Export commands sent to stdout.\n")
		} else {
			if err = file.Save(fullPath, b); err != nil {
				return 0, 0, err
			}

			if secrets {
				if err = file.Save(fmt.Sprintf("%s.secrets", fullPath), bt); err != nil {
					return 0, 0, err
				}
			}

			if len(fileInfo.AternatePath) > 0 {
				fullAternatePath := buildPath(root, fileInfo.AternatePath)

				if err = file.Save(fullAternatePath, bt); err != nil {
					return 0, 0, err
				}
			}

			restoredCount++
		}
	}

	if err = catalog.Write(catalogPath, clog); err != nil {
		return 0, 0, err
	}

	return restoredCount, totalCount, nil
}

func buildPath(root, path string) string {
	if len(root) > 0 {
		return fmt.Sprintf("%s/%s", root, path)
	}
	return path
}

func getPath(path string) string {
	pathEnd := strings.LastIndex(path, "/")
	if pathEnd == -1 {
		return ""
	}
	return path[:pathEnd]
}

func bufferExportScript(file []byte) bytes.Buffer {
	reader := bytes.NewReader(file)
	pairs := gotenv.Parse(reader)
	var b bytes.Buffer

	for key, value := range pairs {
		b.WriteString(fmt.Sprintf("export %s='%s'\n", key, value))
	}

	return b
}

func init() {
	RootCmd.AddCommand(pullCmd)

	pullCmd.Flags().BoolVarP(&exportEnv, "export", "e", false, "Set environment variables from files.")
	pullCmd.Flags().StringVarP(&tagList, "tags", "t", "", "Specify a list of tags used to filter files.")
	pullCmd.Flags().StringVarP(&version, "ver", "v", "", "Set a version to identify a file specific state.")
	pullCmd.Flags().BoolVarP(&secrets, "inject-secrets", "i", false, "Generate *.secrets file containing configuration including secrets.")
}
