package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/display"
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
			display.Error(fmt.Sprintf("Failed to list files for %s.\n", uo.Catalog), ioStreams.UserOutput)
			logger.L.Fatalf("%s\n\n", err)
		}

		color.New(color.Bold).Fprintf(ioStreams.UserOutput, "\n%d file(s) stored remotely.\n\n", total)
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

		fmt.Fprintf(io.UserOutput, "|-")
		color.New(color.FgBlue).Fprintf(io.UserOutput, "%s", fullPath)
		color.New(color.Bold).Fprintf(io.UserOutput, " [%s]", fileEntry.Store)
		fmt.Fprintf(io.UserOutput, "\n")

		if opt.ViewTags && len(fileEntry.Tags) > 0 {
			fmt.Fprintf(io.UserOutput, "|")
			color.New(color.Bold).Fprintln(io.UserOutput, "   tags")
			for _, tag := range fileEntry.Tags {
				fmt.Fprintf(io.UserOutput, "|    |- %s\n", tag)
			}

			if !opt.ViewVersions {
				fmt.Fprintln(io.UserOutput, "|")
			}
		}

		if opt.ViewVersions && len(fileEntry.Versions) > 0 {
			fmt.Fprintf(io.UserOutput, "|")
			color.New(color.Bold).Fprintln(io.UserOutput, "   versions")
			for _, ver := range fileEntry.Versions {
				fmt.Fprintf(io.UserOutput, "|    |- %s\n", ver)
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
