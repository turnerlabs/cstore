package cmd

import (
	"os"

	"github.com/turnerlabs/cstore/components/remote"

	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/display"
	localFile "github.com/turnerlabs/cstore/components/file"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a local catalog file.",
	Long:  `Generate a catalog file without pushing to remote storage.`,
	Run: func(cmd *cobra.Command, userSpecifiedFilePaths []string) {
		setupUserOptions(userSpecifiedFilePaths)

		//-------------------------------------------------
		//- Get or create the local catalog for push.
		//-------------------------------------------------
		clog, err := catalog.GetMake(uo.Catalog, ioStreams)
		if err != nil {
			display.Error(err, ioStreams.UserOutput)
			os.Exit(1)
		}

		for _, filePath := range getFilePathsToPush(clog, uo) {

			file, err := localFile.GetBy(clog.GetFullPath(filePath))
			if err != nil {
				display.Error(err, ioStreams.UserOutput)
				continue
			}

			fileEntry, update := clog.LookupEntry(filePath, file)

			//-------------------------------------------------
			//- Set file options based on command line flags
			//-------------------------------------------------
			fileEntry = updateUserOptions(fileEntry, update, uo)

			//--------------------------------------------------
			//- Get the remote store and vault components ready.
			//--------------------------------------------------
			_, err = remote.InitComponents(&fileEntry, clog, uo, ioStreams)
			if err != nil {
				display.Error(err, ioStreams.UserOutput)
				continue
			}

			if err := clog.UpdateEntry(fileEntry); err != nil {
				display.Error(err, ioStreams.UserOutput)
				continue
			}
		}

		if err := catalog.Write(clog.GetFullPath(uo.Catalog), clog); err != nil {
			display.Error(err, ioStreams.UserOutput)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
