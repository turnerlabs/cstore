package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/logger"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/path"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List cataloged files.",
	Long:  `List cataloged files.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupUserOptions(args)

		fmt.Fprintf(ioStreams.UserOutput, "\nThe files listed are stored remotely. Use CLI flags -g and -v to display tags and versions for each file.\n\n")

		total, err := listFilesFor(uo.Catalog, uo, ioStreams)
		if err != nil {
			fmt.Fprintf(ioStreams.UserOutput, "%sERROR:%s ", uo.Format.Red, uo.Format.NoColor)
			logger.L.Fatalf("%s\n\n", err)
		}

		fmt.Fprintf(ioStreams.UserOutput, "\n%s%d file(s) stored remotely.%s\n\n", uo.Format.Bold, total, uo.Format.UnBold)
	},
}

func listFilesFor(catalogPath string, opt cfg.UserOptions, io models.IO) (int, error) {
	basePath := path.RemoveFileName(catalogPath)

	//-------------------------------------------------
	//- Get catalog containing files to list.
	//-------------------------------------------------
	clog, err := catalog.Get(catalogPath)
	if err != nil {
		return 0, err
	}

	//-------------------------------------------------
	//- Print catalog entries.
	//-------------------------------------------------
	total := 0

	for _, fileEntry := range clog.FilesBy(opt.GetPaths(clog.CWD), opt.TagList, opt.AllTags, opt.Version) {
		fullPath := path.BuildPath(basePath, fileEntry.Path)

		//-------------------------------------------------
		//- If entry is catalog, print child entries.
		//-------------------------------------------------
		if fileEntry.IsRef {
			count, err := listFilesFor(fullPath, opt, io)
			if err != nil {
				return 0, err
			}

			total += count

			continue
		}

		//-------------------------------------------------
		//- Print file entry and versions.
		//-------------------------------------------------
		fmt.Fprintf(io.UserOutput, "|-%s%s%s [%s%s%s] \n", opt.Format.Blue, fullPath, opt.Format.NoColor, opt.Format.Bold, fileEntry.Store, opt.Format.UnBold)

		if opt.ViewTags && len(fileEntry.Tags) > 0 {
			fmt.Fprintf(io.UserOutput, "|   %stags%s\n", opt.Format.Bold, opt.Format.UnBold)
			for _, tag := range fileEntry.Tags {
				fmt.Fprintf(io.UserOutput, "|    %s|- %s%s\n", opt.Format.Bold, opt.Format.UnBold, tag)
			}

			if !opt.ViewVersions {
				fmt.Fprintln(io.UserOutput, "|")
			}
		}

		if opt.ViewVersions && len(fileEntry.Versions) > 0 {
			fmt.Fprintf(io.UserOutput, "|   %sversions%s\n", opt.Format.Bold, opt.Format.UnBold)
			for _, ver := range fileEntry.Versions {
				fmt.Fprintf(io.UserOutput, "|    %s|- %s%s\n", opt.Format.Bold, opt.Format.UnBold, ver)
			}
			fmt.Fprintln(io.UserOutput, "|")
		}

		total++
	}

	return total, nil
}

func init() {
	RootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&uo.Tags, "tags", "t", "", "Specify a list of tags used to filter files.")
	listCmd.Flags().BoolVarP(&uo.ViewTags, "view-tags", "g", false, "Display a list of tags for each file.")
	listCmd.Flags().BoolVarP(&uo.ViewVersions, "view-version", "v", false, "Display a list of versions for each file.")
}
