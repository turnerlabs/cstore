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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/file"
	"github.com/turnerlabs/cstore/components/logger"
	"github.com/turnerlabs/cstore/components/store"
	"github.com/turnerlabs/cstore/components/token"
	"github.com/turnerlabs/cstore/components/vault"
)

const (
	//"\xE2\x9C\x94" This is a checkmark on mac, but question mark on windows; so,
	// will use (done) for now to support multiple platforms.
	checkMark = "(done)"

	storeToken = "store"
)

var (
	storeName         string
	tagList           string
	version           string
	alternateFilePath string
	modifySecrets     bool
	deleteFile        bool
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Store file(s) remotely.",
	Long:  `Store file(s) remotely.`,
	Run: func(cmd *cobra.Command, paths []string) {

		clog, err := catalog.GetMake(viper.GetString(fileToken))
		if err != nil {
			logger.L.Fatal(err)
		}

		pathsSpecified := len(paths) > 0

		tags := getTags(tagList)

		if len(paths) == 0 {
			if len(tags) > 0 {
				paths = append(paths, clog.GetTaggedPaths(tags, true)...)
			} else {
				paths = clog.GetFileNames()
			}
		}

		paths = removeDups(paths)

		filesAdded := []string{}
		pendingFiles := 0

		for _, path := range paths {

			targetFile, err := file.Get(path)

			if err != nil {
				logger.L.Print(err)
				continue
			}

			newFile := catalog.File{
				Path:         path,
				AternatePath: alternateFilePath,
				Store:        storeName,
				Vaults:       catalog.Vault{},
				IsRef:        catalog.IsOne(targetFile),
				IsEnv:        file.IsEnv(path),
				Tags:         clog.GetTagsBy(path),
			}

			if pathsSpecified {
				newFile.Tags = tags
			}

			if !clog.Exists(newFile) && !newFile.IsRef {
				if len(newFile.Store) == 0 {
					newFile.Store = viper.GetString(storeToken)
					newFile.Vaults.Credentials = viper.GetString(credsToken)
					newFile.Vaults.Encryption = viper.GetString(encryptToken)
				}
			}

			fileData, err := clog.Update(newFile)
			if err != nil {
				logger.L.Print(err)
				continue
			}

			if fileData.IsRef {
				logger.L.Printf("Linking %s   %s \n", fileData.Path, checkMark)
				continue
			}

			pendingFiles++

			cv, err := vault.GetBy(fileData.Vaults.Credentials)
			if err != nil {
				logger.L.Print(err)
				continue
			}

			ev, err := vault.GetBy(fileData.Vaults.Encryption)
			if err != nil {
				logger.L.Print(err)
				continue
			}

			st, err := store.Select(fileData, clog.Context, cv, ev, viper.GetBool(promptToken))
			if err != nil {
				logger.L.Print(err)
				continue
			}

			if !st.CanHandleFile(fileData) {
				logger.L.Printf("'%s' cannot store file '%s'.", st.Name(), fileData.Path)
				continue
			}

			contextKey := clog.ContextKey(fileData.Key())

			push := true
			if _, attr, err := st.Pull(contextKey, fileData); err == nil {
				if clog.FilePulledBefore(fileData.Key(), "", attr.LastModified) {
					fmt.Fprintf(os.Stderr, "Remote file '%s' has changed since your last pull. Overwrite?", path)
					push = confirm()
				}
			}

			fPath := formatPath(fileData.Path)

			if push {
				if modifySecrets {
					tokens := token.Find(targetFile, clog.Context, true)

					if _, err := st.SetTokens(tokens, clog.Context); err != nil {
						logger.L.Fatal(err)
					}

					targetFile = token.Clean(targetFile)

					if err = file.Save(fileData.Path, targetFile); err != nil {
						logger.L.Print(err)
					}
				}

				fmt.Fprintf(os.Stderr, "Pushing %s ", fPath)
				data, encrypted, err := st.Push(contextKey, fileData, targetFile)
				if err != nil {
					logger.L.Print(err)
					continue
				}

				if len(version) > 0 {
					versionedKey := fmt.Sprintf("%s/%s", contextKey, version)

					_, _, err := st.Push(versionedKey, fileData, targetFile)
					if err != nil {
						logger.L.Print(err)
						continue
					}

					if len(fileData.Versions) == 0 {
						fileData.Versions = []string{}
					}

					if !fileData.VersionExists(version) {
						fileData.Versions = append(fileData.Versions, version)
					}
				}

				fileData.Encrypted = encrypted
				fileData.Data = data
				clog.Files[fileData.Key()] = fileData

				_, attr, err := st.Pull(contextKey, fileData)
				if err != nil {
					logger.L.Print(err)
					continue
				}

				if err := clog.FilePulled(fileData.Key(), "", attr.LastModified); err != nil {
					logger.L.Print(err)
					continue
				}

				logger.L.Printf(" %s \n", checkMark)
				filesAdded = append(filesAdded, fileData.Path)

			} else {
				logger.L.Printf("Skipping %s x \n", fPath)
			}
		}

		if err := catalog.Write(viper.GetString(fileToken), clog); err != nil {
			logger.L.Print(err)
			os.Exit(1)
		}

		if deleteFile {
			for _, path := range filesAdded {
				os.Remove(path)
				os.Remove(fmt.Sprintf("%s.secrets", path))
			}
		}

		logger.L.Printf("%d of %d file(s) pushed to remote store. \n", len(filesAdded), pendingFiles)
	},
}

func removeDups(elements []string) []string {
	uniquePaths := map[string]string{}

	for v := range elements {
		uniquePaths[elements[v]] = elements[v]
	}

	result := []string{}
	for key := range uniquePaths {
		result = append(result, key)
	}

	return result
}

func getPaths(args []string) []string {
	return args
}

func getTags(tagList string) []string {
	tags := strings.Split(tagList, "|")

	if len(tags) == 1 && tags[0] == "" {
		return []string{}
	}

	return tags
}

func formatPath(path string) string {
	const totalLength int = 30
	const ellipsis string = "..."

	pathLength := len(path)

	if pathLength > totalLength {
		trimLength := pathLength - (totalLength - len(ellipsis))

		path = fmt.Sprintf("%s%s", ellipsis, path[trimLength:pathLength])
	} else {
		for i := 0; i < (totalLength - pathLength); i++ {
			path = fmt.Sprintf("%s ", path)
		}
	}

	return path
}

func confirm() bool {
	var s string

	fmt.Printf(" (y/N): ")
	fmt.Scanf("%s", &s)

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return false
}

func init() {
	RootCmd.AddCommand(pushCmd)

	pushCmd.Flags().StringVarP(&storeName, "store", "s", "", "Set the context store used to store files. The 'stores' command lists options.")
	pushCmd.Flags().BoolVarP(&deleteFile, "delete", "d", false, "Delete the local file after pushing.")
	pushCmd.Flags().StringVarP(&tagList, "tags", "t", "", "Set a list of tags used to identify the file.")
	pushCmd.Flags().StringVarP(&version, "ver", "v", "", "Set a version to identify the file current state.")
	pushCmd.Flags().StringVarP(&alternateFilePath, "alt", "a", "", "Set an alternate path to clone the file to during a restore.")
	pushCmd.Flags().BoolVarP(&modifySecrets, "modify-secrets", "m", false, "Store secrets for tokens in file.")
}
