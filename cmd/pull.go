package cmd

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/display"
	"github.com/turnerlabs/cstore/components/env"
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
			display.Error(fmt.Errorf("%s for %s", err, uo.Catalog), ioStreams.UserOutput)
			os.Exit(1)
		} else {
			color.New(color.Bold).Fprintf(ioStreams.UserOutput, "\n%d of %d requested file(s) retrieved.\n\n", count, total)
		}
	},
}

// Pull ...
func Pull(catalogPath string, opt cfg.UserOptions, io models.IO) (int, int, error) {
	restoredCount := 0
	fileCount := 0

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
		remoteComp, err := getRemoteComponents(&fileEntryTemp, clog, opt, io)
		if err != nil {
			display.Error(fmt.Errorf("Could not retrieve %s! (%s)", path.BuildPath(root, fileEntry.Path), err), io.UserOutput)
			continue
		}

		//----------------------------------------------------
		//- Pull remote file from store.
		//----------------------------------------------------
		file, _, err := remoteComp.store.Pull(&fileEntry, opt.Version)
		if err != nil {
			display.Error(fmt.Errorf("Could not retrieve %s! (%s)", path.BuildPath(root, fileEntry.Path), err), io.UserOutput)
			continue
		}

		//----------------------------------------------------
		//- Remove environment variables already exported
		//----------------------------------------------------
		if opt.NoOverwrite {
			file = env.DiffCurrent(file)
		}

		//-------------------------------------------------
		//- If user specifies, inject secrets into file.
		//-------------------------------------------------
		fileWithSecrets := file

		if opt.InjectSecrets {
			if !fileEntry.SupportsSecrets() {
				display.Error(fmt.Errorf("Secrets not supported for %s due to incompatible file type.", fileEntry.Path), io.UserOutput)
				continue
			}

			tokens, err := token.Find(fileWithSecrets, fileEntry.Type, false)
			if err != nil {
				display.Error(fmt.Errorf("Failed to find tokens in file %s. (%s)", fileEntry.Path, err), io.UserOutput)
			}

			for k, t := range tokens {

				value, err := remoteComp.secrets.Get(clog.Context, t.Secret(), t.Prop)
				if err != nil {
					display.Error(fmt.Errorf("Failed to get value for %s/%s for %s! (%s)", t.Secret(), t.Prop, path.BuildPath(root, fileEntry.Path), t.Secret()), io.UserOutput)
					continue
				}

				t.Value = value
				tokens[k] = t
			}

			fileWithSecrets, err = token.Replace(fileWithSecrets, fileEntry.Type, tokens)
			if err != nil {
				display.Error(fmt.Errorf("Failed to replace tokens in file %s. (%s)", fileEntry.Path, err), io.UserOutput)
			}
		}

		//----------------------------------------------------
		//- If user specifies, send export commands to stdout.
		//----------------------------------------------------
		if opt.ExportEnv || len(opt.ExportFormat) > 0 {

			script := bytes.Buffer{}
			msg := "\n%s sent to stdout.\n"

			switch fileEntry.Type {
			case "env":
				switch opt.ExportFormat {
				case "task-def-secrets":
					msg = fmt.Sprintf(msg, "AWS task definition secrets")
					script, err = toTaskDefSecretFormat(fileWithSecrets)
					if err != nil {
						logger.L.Print(err)
					}
				case "task-def-env":
					msg = fmt.Sprintf(msg, "AWS task definition environment")
					script, err = toTaskDefEnvFormat(fileWithSecrets)
					if err != nil {
						logger.L.Print(err)
					}
				default:
					msg = fmt.Sprintf(msg, "Terminal export commands")
					script, err = bufferExportScript(fileWithSecrets)
					if err != nil {
						logger.L.Print(err)
					}
				}
			case "json":
				script.Write(fileWithSecrets)
				msg = fmt.Sprintf(msg, "JSON")
			}

			if script.Len() > 0 {
				if _, err := script.WriteTo(io.Export); err != nil {
					return 0, 0, err
				}

				fmt.Fprintf(io.UserOutput, msg)

				restoredCount++
				continue
			}
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

		if len(fileEntry.AlternatePath) > 0 || len(opt.AlternateRestorePath) > 0 {
			fullAternatePath := clog.GetFullPath(path.BuildPath(root, fileEntry.AlternatePath))

			if err = localFile.Save(fullAternatePath, fileWithSecrets); err != nil {
				return 0, 0, err
			}
		}

		fmt.Fprint(io.UserOutput, "Retrieving [")
		color.New(color.FgBlue).Fprintf(io.UserOutput, path.BuildPath(root, fileEntry.Path))
		fmt.Fprint(io.UserOutput, "]")

		if len(opt.Version) > 0 {
			fmt.Fprintf(io.UserOutput, "(%s)", opt.Version)
		}

		fmt.Fprint(io.UserOutput, " <- [")
		color.New(color.Bold).Fprintf(io.UserOutput, remoteComp.store.Name())
		fmt.Fprintln(io.UserOutput, "]")

		restoredCount++

		//-------------------------------------------------
		//- Save the time the user last pulled file.
		//-------------------------------------------------
		if err := clog.RecordPull(fileEntry.Key(), time.Now()); err != nil {
			logger.L.Print(err)
			continue
		}
	}

	return restoredCount, fileCount, nil
}

func init() {
	RootCmd.AddCommand(pullCmd)

	pullCmd.Flags().BoolVarP(&uo.ExportEnv, "export", "e", false, "Append export command to environment variables and send to stdout.")
	pullCmd.Flags().StringVarP(&uo.Tags, "tags", "t", "", "Specify a list of tags used to filter files.")
	pullCmd.Flags().StringVarP(&uo.ExportFormat, "format", "g", "", "Format environment variables and send to stdout")
	pullCmd.Flags().StringVarP(&uo.Version, "ver", "v", "", "Set a version to identify a file specific state.")
	pullCmd.Flags().BoolVarP(&uo.InjectSecrets, "inject-secrets", "i", false, "Generate *.secrets file containing configuration including secrets.")
	pullCmd.Flags().StringVarP(&uo.AlternateRestorePath, "alt", "a", "", "Set an alternate path to clone the file to during a restore.")
	pullCmd.Flags().BoolVarP(&uo.NoOverwrite, "no-overwrite", "n", false, "Only pulls the environment variables that are not exported in the current environment.")
}
