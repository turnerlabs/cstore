package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/display"
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
			display.Error(fmt.Sprintf("%s for %s\n", err, uo.Catalog), ioStreams.UserOutput)
			os.Exit(1)
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
	files := clog.FilesBy(opt.GetPaths(clog.CWD), opt.TagList, opt.AllTags, opt.Version)

	if len(files) == 0 {
		display.Error(fmt.Sprint("\nNo matching files stored remotely!"), ioStreams.UserOutput)
		os.Exit(0)
	}

	fileList := ""
	for _, f := range files {
		if len(opt.Version) > 0 {
			fileList = fmt.Sprintf("%sDelete [%s](%s) from [%s]\n", fileList, f.Path, opt.Version, f.Store)
		} else {
			fileList = fmt.Sprintf("%sDelete [%s] from [%s]\n", fileList, f.Path, f.Store)
		}
	}

	if !prompt.Confirm(fmt.Sprintf("Files will be permanently deleted from remote storage!\n\n%s \nContinue?", fileList), true, io) {
		color.New(color.Bold, color.FgRed).Fprint(ioStreams.UserOutput, "\nOperation Aborted!\n")
		os.Exit(0)
	}

	//-------------------------------------------------
	//- Processed each file being purged.
	//-------------------------------------------------
	for key, fileEntry := range files {

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
		remoteComp, err := getRemoteComponents(&fileEntryTemp, clog, opt, io)
		if err != nil {
			display.Error(fmt.Sprintf("Could not purge %s!", fileEntry.Path), ioStreams.UserOutput)
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

			purged++
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
					display.Error(fmt.Sprintf("Failed to purge %s!", fileEntry.Path), io.UserOutput)
					logger.L.Print(err)
					fmt.Fprintln(io.UserOutput)
					continue
				}
			}

			//----------------------------------------------------
			//- Delete the file.
			//----------------------------------------------------
			if err = remoteComp.store.Purge(&fileEntry, none); err != nil {
				display.Error(fmt.Sprintf("Failed to purge %s!", fileEntry.Path), io.UserOutput)
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
					display.Error(fmt.Sprintf(".cstore file could not be removed for %s!", fileEntry.Path), io.UserOutput)
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

	color.New(color.Bold).Fprintf(ioStreams.UserOutput, "\n%d of %d file(s) purged from remote storage.\n\n", purged, count)

	return nil
}

func init() {
	RootCmd.AddCommand(purgeCmd)

	purgeCmd.Flags().StringVarP(&uo.Tags, "tags", "t", "", "Specify a list of tags used to filter files.")
	purgeCmd.Flags().StringVarP(&uo.Version, "ver", "v", "", "Remove specific version.")
}
