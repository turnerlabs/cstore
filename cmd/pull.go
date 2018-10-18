package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	localFile "github.com/turnerlabs/cstore/components/file"
	"github.com/turnerlabs/cstore/components/logger"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/path"
	"github.com/turnerlabs/cstore/components/token"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Retrieve file(s).",
	Long:  `Retrieve file(s).`,
	Run: func(cmd *cobra.Command, userSpecifiedFilePaths []string) {
		setupUserOptions(userSpecifiedFilePaths)

		if count, total, err := Pull(uo.Catalog, uo, ioStreams); err != nil {
			fmt.Fprintf(ioStreams.UserOutput, "%sERROR:%s ", uo.Format.Red, uo.Format.NoColor)
			logger.L.Fatalf("%s\n\n", err)
		} else {
			fmt.Fprintf(ioStreams.UserOutput, "\n%s%d of %d requested file(s) retrieved.%s\n\n", uo.Format.Bold, count, total, uo.Format.UnBold)
		}
	},
}

// Pull ...
func Pull(catalogPath string, opt cfg.UserOptions, io models.IO) (int, int, error) {
	restoredCount := 0
	fileCount := 0
	errorOccured := false

	//-------------------------------------------------
	//- Get the local catalog for reference.
	//-------------------------------------------------
	clog, err := catalog.Get(catalogPath)
	if err != nil {
		return 0, 0, err
	}

	root := path.RemoveFileName(catalogPath)

	//----------------------------------------------------------
	//- Attempt to restore requested files.
	//-
	//- Files can be retrieved with a specified version. If a
	//- versioned file entry is not found in the catalog, cStore
	//- will attempt to restore that version of all file entries
	//- matching the remaining criteria. This provides the
	//- ability to get only versioned files when the catalog
	//- aware of the version or to store and retrieve versions
	//- without the catalog being aware of the version. This is
	//- useful, when a version needs to be pushed and pulled,
	//- but the catalog file cannot be updated easily.
	//----------------------------------------------------------
	fmt.Fprintln(io.UserOutput)

	files := clog.FilesBy(opt.GetPaths(clog.CWD), opt.TagList, opt.AllTags, opt.Version)

	if len(opt.Version) > 0 && len(files) == 0 {
		files = clog.FilesBy(opt.GetPaths(clog.CWD), opt.TagList, opt.AllTags, "")
	}

	if len(files) == 0 {
		return 0, 0, fmt.Errorf("%s is not aware of requested files. Use 'list' command to view available files.", opt.Catalog)
	}

	for _, fileEntry := range files {

		//-----------------------------------------------------
		//- Override saved file settings with user preferences.
		//-----------------------------------------------------
		fileEntry = overrideFileSettings(fileEntry, opt)

		//----------------------------------------------------
		//- Check for a linked catalog with child files.
		//----------------------------------------------------
		if fileEntry.IsRef {
			c, t, err := Pull(path.BuildPath(root, fileEntry.Path), opt, io)
			if err != nil {
				return 0, 0, err
			}

			restoredCount += c
			fileCount += t
			continue
		}

		fileCount++

		//----------------------------------------------------
		//- Get the remote store and vaults components ready.
		//----------------------------------------------------
		fileEntryTemp := fileEntry
		remoteComp, err := getRemoteComponents(&fileEntryTemp, clog, opt.Prompt, io)
		if err != nil {
			fmt.Fprintf(io.UserOutput, "%sERROR:%s Could not retrieve %s!\n", opt.Format.Red, opt.Format.NoColor, path.BuildPath(root, fileEntry.Path))
			logger.L.Print(err)
			fmt.Fprintln(io.UserOutput)
			errorOccured = true
			continue
		}

		//----------------------------------------------------
		//- Pull remote file from store.
		//----------------------------------------------------
		file, attr, err := remoteComp.store.Pull(&fileEntry, opt.Version)
		if err != nil {
			fmt.Fprintf(io.UserOutput, "%sERROR:%s Could not retrieve %s!\n", opt.Format.Red, opt.Format.NoColor, path.BuildPath(root, fileEntry.Path))
			logger.L.Print(err)
			fmt.Fprintln(io.UserOutput)
			errorOccured = true
			continue
		}

		//-------------------------------------------------
		//- If user specifies, inject secrets into file.
		//-------------------------------------------------
		fileWithSecrets := file
		if opt.InjectSecrets {
			if !fileEntry.SupportsSecrets() {
				fmt.Fprintf(io.UserOutput, "%sERROR:%s Secrets not supported for %s due to incompatible file type.\n\n", opt.Format.Red, opt.Format.NoColor, fileEntry.Path)
				errorOccured = true
				continue
			}

			tokens, err := token.Find(fileWithSecrets, fileEntry.Type, false)
			if err != nil {
				fmt.Fprintf(io.UserOutput, "%sERROR:%s Failed to find tokens in file %s.\n\n", opt.Format.Red, opt.Format.NoColor, fileEntry.Path)
				logger.L.Print(err)
			}

			for k, t := range tokens {

				value, err := remoteComp.secrets.Get(clog.Context, t.Secret(), t.Prop)
				if err != nil {
					fmt.Fprintf(io.UserOutput, "%sERROR:%s Failed to get value for %s/%s for %s!\n (%s) ", opt.Format.Red, opt.Format.NoColor, t.Secret(), t.Prop, path.BuildPath(root, fileEntry.Path), t.Secret())
					logger.L.Print(err)
					fmt.Fprintln(io.UserOutput)
					errorOccured = true
					continue
				}

				t.Value = value
				tokens[k] = t
			}

			fileWithSecrets, err = token.Replace(fileWithSecrets, fileEntry.Type, tokens)
			if err != nil {
				fmt.Fprintf(io.UserOutput, "%sERROR:%s Failed to replace tokens in file %s.\n\n", opt.Format.Red, opt.Format.NoColor, fileEntry.Path)
				logger.L.Print(err)
			}
		}

		//----------------------------------------------------
		//- If user specifies, send export commands to stdout.
		//----------------------------------------------------
		if opt.ExportEnv && fileEntry.Type == "env" {
			script := bufferExportScript(fileWithSecrets)

			if _, err := script.WriteTo(io.Export); err != nil {
				return 0, 0, err
			}

			fmt.Fprintf(io.UserOutput, "\nExport commands sent to stdout.\n")

			restoredCount++
			continue
		}

		//-----------------------------------------------------
		//- Save editable, secret, and alternate files locally.
		//-----------------------------------------------------
		fullPath := clog.GetFullPath(path.BuildPath(root, fileEntry.Path))

		if len(opt.AlternateRestorePath) == 0 {
			if err = localFile.Save(fullPath, file); err != nil {
				return 0, 0, err
			}
		}

		if opt.InjectSecrets {
			if err = localFile.Save(fmt.Sprintf("%s.secrets", fullPath), fileWithSecrets); err != nil {
				return 0, 0, err
			}
		}

		if len(fileEntry.AternatePath) > 0 || len(opt.AlternateRestorePath) > 0 {
			fullAternatePath := clog.GetFullPath(path.BuildPath(root, fileEntry.AternatePath))

			if err = localFile.Save(fullAternatePath, fileWithSecrets); err != nil {
				return 0, 0, err
			}
		}

		fmt.Fprintf(io.UserOutput, "Retrieve [%s%s%s] -> [%s%s%s]\n", opt.Format.Bold, remoteComp.store.Name(), opt.Format.UnBold, opt.Format.Blue, path.BuildPath(root, fileEntry.Path), opt.Format.NoColor)

		restoredCount++

		//-------------------------------------------------
		//- Save the time the user last pulled file.
		//-------------------------------------------------
		if err := clog.RecordPull(fileEntry.Key(), attr.LastModified); err != nil {
			logger.L.Print(err)
			errorOccured = true
			continue
		}
	}

	if errorOccured {
		return restoredCount, fileCount, errors.New("issues were encountered for some files")
	}

	return restoredCount, fileCount, nil
}

func init() {
	RootCmd.AddCommand(pullCmd)

	pullCmd.Flags().BoolVarP(&uo.ExportEnv, "export", "e", false, "Set environment variables from files.")
	pullCmd.Flags().StringVarP(&uo.Tags, "tags", "t", "", "Specify a list of tags used to filter files.")
	pullCmd.Flags().StringVarP(&uo.Version, "ver", "v", "", "Set a version to identify a file specific state.")
	pullCmd.Flags().BoolVarP(&uo.InjectSecrets, "inject-secrets", "i", false, "Generate *.secrets file containing configuration including secrets.")
	pullCmd.Flags().StringVarP(&uo.AlternateRestorePath, "alt", "a", "", "Set an alternate path to clone the file to during a restore.")
}
