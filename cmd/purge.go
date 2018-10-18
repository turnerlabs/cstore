package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/logger"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/path"
	"github.com/turnerlabs/cstore/components/prompt"
)

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge file(s) remotely.",
	Long: `Purge file(s) remotely.

When all files were removed successfully, the local catalog is deleted.

Purge does not delete linked catalogs or their files.`,
	Run: func(cmd *cobra.Command, userSpecifiedFilePaths []string) {
		setupUserOptions(userSpecifiedFilePaths)

		if err := Purge(uo, ioStreams); err != nil {
			fmt.Fprintf(ioStreams.UserOutput, "%sERROR:%s ", uo.Format.Red, uo.Format.NoColor)
			logger.L.Fatalf("%s\n\n", err)
		}
	},
}

// Purge ...
func Purge(opt cfg.UserOptions, io models.IO) error {
	count := 0
	purged := 0

	//-------------------------------------------------
	//- Get the local catalog for reference.
	//-------------------------------------------------
	clog, err := catalog.Get(opt.Catalog)
	if err != nil {
		return err
	}

	//-------------------------------------------------
	//- Confirm file deletes with user.
	//-------------------------------------------------
	if !prompt.Confirm("Remote files will be permanently deleted! Continue?", true, io) {
		fmt.Fprintf(io.UserOutput, "\n%s%sOperation Aborted!%s%s\n", opt.Format.Red, opt.Format.Bold, opt.Format.UnBold, opt.Format.NoColor)
		os.Exit(0)
	}

	//-------------------------------------------------
	//- Processed each file being purged.
	//-------------------------------------------------
	for key, fileEntry := range clog.FilesBy(opt.GetPaths(clog.CWD), opt.TagList, opt.AllTags, opt.Version) {

		//-------------------------------------------------
		//- Delete links but not the link catalogs files.
		//-------------------------------------------------
		if fileEntry.IsRef {
			delete(clog.Files, key)
			continue
		}

		count++

		//-----------------------------------------------------
		//- Override saved file settings with user preferences.
		//-----------------------------------------------------
		fileEntryTemp := overrideFileSettings(fileEntry, opt)

		//----------------------------------------------------
		//- Get the remote store and vaults components ready.
		//----------------------------------------------------
		remoteComp, err := getRemoteComponents(&fileEntryTemp, clog, opt.Prompt, io)
		if err != nil {
			fmt.Fprintf(io.UserOutput, "%sERROR:%s Could not purge %s!\n", opt.Format.Red, opt.Format.NoColor, fileEntry.Path)
			logger.L.Print(err)
			fmt.Fprintln(io.UserOutput)
			continue
		}

		//----------------------------------------------------
		//- If version specified, delete it.
		//----------------------------------------------------
		if len(opt.Version) > 0 {
			if err = remoteComp.store.Purge(&fileEntry, opt.Version); err != nil {
				logger.L.Print(err)
			}

			for i, ver := range fileEntry.Versions {
				if ver == opt.Version {
					fileEntry.Versions = append(fileEntry.Versions[:i], fileEntry.Versions[i+1:]...)
				}
			}

			clog.Files[key] = fileEntry
		}

		//----------------------------------------------------
		//- If no version specified, delete all versions.
		//----------------------------------------------------
		if len(opt.Version) == 0 {
			//----------------------------------------------------
			//- Delete all file versions.
			//----------------------------------------------------
			for _, version := range fileEntry.Versions {
				if err = remoteComp.store.Purge(&fileEntry, version); err != nil {
					fmt.Fprintf(io.UserOutput, "%sError:%s Failed to purge %s!\n", opt.Format.Red, opt.Format.NoColor, fileEntry.Path)
					logger.L.Print(err)
					fmt.Fprintln(io.UserOutput)
					continue
				}
			}

			//----------------------------------------------------
			//- Delete the file.
			//----------------------------------------------------
			if err = remoteComp.store.Purge(&fileEntry, none); err != nil {
				fmt.Fprintf(io.UserOutput, "%sERROR:%s Failed to purge %s!\n", opt.Format.Red, opt.Format.NoColor, fileEntry.Path)
				logger.L.Print(err)
				fmt.Fprintln(io.UserOutput)
				continue
			}

			delete(clog.Files, key)
			purged++

			//----------------------------------------------------
			//- Delete the ghost .cstore reference file.
			//----------------------------------------------------
			fullPath := clog.GetFullPath(path.RemoveFileName(fileEntry.Path))
			if len(fullPath) > 0 && !clog.AnyFilesIn(path.RemoveFileName(fileEntry.Path)) {
				if err := os.Remove(fmt.Sprintf("%s%s", fullPath, catalog.GhostFile)); err != nil {
					fmt.Fprintf(io.UserOutput, "%sERROR:%s .cstore file could not be removed for %s!\n", opt.Format.Red, opt.Format.NoColor, fileEntry.Path)
					logger.L.Print(err)
					fmt.Fprintln(io.UserOutput)
				}
			}

			//----------------------------------------------------
			//- Remove file audit records.
			//----------------------------------------------------
			if err := clog.RemoveRecords(key); err != nil {
				logger.L.Print(err)
			}

		}
	}

	//----------------------------------------------------
	//- Delete or update the local catalog.
	//----------------------------------------------------
	if len(clog.Files) == 0 {
		if err = catalog.Remove(clog.GetFullPath(opt.Catalog)); err != nil {
			logger.L.Fatal(err)
		}
	} else {
		if err = catalog.Write(clog.GetFullPath(opt.Catalog), clog); err != nil {
			logger.L.Fatal(err)
		}
	}

	fmt.Fprintf(io.UserOutput, "\n%s%d of %d file(s) purged from remote storage.%s\n\n", opt.Format.Bold, purged, count, opt.Format.UnBold)

	return nil
}

func init() {
	RootCmd.AddCommand(purgeCmd)

	purgeCmd.Flags().StringVarP(&uo.Tags, "tags", "t", "", "Specify a list of tags used to filter files.")
	purgeCmd.Flags().StringVarP(&uo.Version, "ver", "v", "", "Remove specific version.")
}
