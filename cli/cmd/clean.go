package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/display"
	"github.com/turnerlabs/cstore/components/path"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Deletes local files to protect configuration.",
	Long:  `Deletes local files to protect configuration.`,
	Run: func(cmd *cobra.Command, userSpecifiedFilePaths []string) {
		setupUserOptions(userSpecifiedFilePaths)

		cleanCatalog(uo.Catalog)
	},
}

func cleanCatalog(catalogPath string) {
	//-------------------------------------------------
	//- Get or create the local catalog for push.
	//-------------------------------------------------
	clog, err := catalog.GetMake(catalogPath, ioStreams)
	if err != nil {
		display.Error(err, ioStreams.UserOutput)
		os.Exit(1)
	}

	root := path.RemoveFileName(catalogPath)

	files := clog.FilesBy(uo.GetPaths(clog.CWD), uo.TagList, uo.AllTags, "")

	if len(files) == 0 {
		display.Error(fmt.Errorf("%s is not aware of requested files. Use 'list' command to view available files.", uo.Catalog), ioStreams.UserOutput)
		os.Exit(1)
	}

	for _, f := range files {
		if f.IsRef {
			cleanCatalog(path.BuildPath(root, f.Path))
		}

		file := clog.GetFullPath(f.Path)
		if err := os.Remove(file); err != nil {
			if !os.IsNotExist(err) {
				display.Error(fmt.Errorf("failed to delete %s (%s)", file, err), ioStreams.UserOutput)
			}
		}

		alternateFile := clog.GetFullPath(f.AlternatePath)
		if err := os.Remove(alternateFile); err != nil {
			if !os.IsNotExist(err) {
				display.Error(fmt.Errorf("failed to delete %s (%s)", alternateFile, err), ioStreams.UserOutput)
			}
		}

		secretsFile := fmt.Sprintf("%s.secrets", clog.GetFullPath(f.Path))
		if err := os.Remove(secretsFile); err != nil {
			if !os.IsNotExist(err) {
				display.Error(fmt.Errorf("failed to delete %s (%s)", secretsFile, err), ioStreams.UserOutput)
			}

		}
	}
}

func init() {
	RootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().StringVarP(&uo.Tags, "tags", "t", "", "Specify a list of tags used to filter files.")
}
