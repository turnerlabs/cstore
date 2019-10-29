package cmd

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/turnerlabs/cstore/components/convert"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/display"
	"github.com/turnerlabs/cstore/components/env"
	localFile "github.com/turnerlabs/cstore/components/file"
	"github.com/turnerlabs/cstore/components/logger"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/path"
	"github.com/turnerlabs/cstore/components/remote"
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

	exportBuffer := bytes.Buffer{}

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
	//- ability to get only versioned files when the catalog is
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
		return 0, 0, fmt.Errorf("requested files not cataloged")
	}

	for _, fileEntry := range files {

		//-----------------------------------------------------
		//- Override saved file settings with user preferences.
		//-----------------------------------------------------
		fileEntry = remote.OverrideFileSettings(fileEntry, opt)

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
		remoteComp, err := remote.InitComponents(&fileEntryTemp, clog, opt, io)
		if err != nil {
			display.Error(fmt.Errorf("PullFailedException3: %s (%s)", getPath(root, fileEntry.Path, opt.Version), err), io.UserOutput)
			continue
		}

		//----------------------------------------------------
		//- Pull remote file from store.
		//----------------------------------------------------
		file, _, err := remoteComp.Store.Pull(&fileEntry, opt.Version)
		if err != nil {
			display.Error(fmt.Errorf("PullFailedException4: %s (%s)", getPath(root, fileEntry.Path, opt.Version), err), io.UserOutput)
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

		if opt.InjectSecrets || opt.ModifySecrets {
			if !fileEntry.SupportsSecrets() {
				display.Error(fmt.Errorf("IncompatibleFileError: %s secrets not supported", fileEntry.Path), io.UserOutput)
				continue
			}

			tokens, err := token.Find(fileWithSecrets, fileEntry.Type, false)
			if err != nil {
				display.Error(fmt.Errorf("MissingTokensError: failed to find tokens in file %s (%s)", fileEntry.Path, err), io.UserOutput)
			}

			for k, t := range tokens {

				value, err := remoteComp.Secrets.Get(clog.Context, t.Secret(), t.Prop)
				if err != nil {
					display.Error(fmt.Errorf("GetSecretValueError: failed to get value for %s/%s for %s (%s)", t.Secret(), t.Prop, path.BuildPath(root, fileEntry.Path), err), io.UserOutput)
					continue
				}

				t.Value = value
				tokens[k] = t
			}

			if opt.ModifySecrets {
				file, err = token.Replace(file, fileEntry.Type, tokens, true)

				if err != nil {
					display.Error(fmt.Errorf("TokenReplacementError: failed to replace tokens in file %s (%s)", fileEntry.Path, err), io.UserOutput)
				}
			}

			if opt.InjectSecrets {
				fileWithSecrets, err = token.Replace(fileWithSecrets, fileEntry.Type, tokens, false)
				if err != nil {
					display.Error(fmt.Errorf("TokenReplacementError: failed to replace tokens in file %s (%s)", fileEntry.Path, err), io.UserOutput)
				}
			}
		}

		//----------------------------------------------------
		//- If user specifies, send export commands to stdout.
		//----------------------------------------------------
		if opt.ExportEnv || len(opt.ExportFormat) > 0 {

			if !compatibleFormat(opt.ExportFormat, fileEntry.Type) {
				display.Error(fmt.Errorf("IncompatibleExportFormat: file %s is incompatible with export format %s", fileEntry.Path, opt.ExportFormat), io.UserOutput)
				continue
			}

			if _, err := exportBuffer.Write(fileWithSecrets); err != nil {
				logger.L.Print(err)
			}

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
		color.New(color.Bold).Fprintf(io.UserOutput, remoteComp.Store.Name())
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

	if opt.ExportEnv || len(opt.ExportFormat) > 0 {

		msg := "\n%s sent to stdout\n"

		switch opt.ExportFormat {
		case "task-def-secrets":
			msg = fmt.Sprintf(msg, "AWS task definition secrets")
			exportBuffer, err = toTaskDefSecretFormat(exportBuffer.Bytes())
			if err != nil {
				logger.L.Print(err)
			}
		case "task-def-env":
			msg = fmt.Sprintf(msg, "AWS task definition environment")
			exportBuffer, err = toTaskDefEnvFormat(exportBuffer.Bytes())
			if err != nil {
				logger.L.Print(err)
			}
		case "json-object":
			msg = fmt.Sprintf(msg, "JSON object")
			exportBuffer, err = convert.ToJSONObjectFormat(exportBuffer.Bytes())
			if err != nil {
				logger.L.Print(err)
			}
		case "terminal-export":
			msg = fmt.Sprintf(msg, "terminal export commands")
			exportBuffer, err = bufferExportScript(exportBuffer.Bytes())
			if err != nil {
				logger.L.Print(err)
			}
		default:
			msg = fmt.Sprintf(msg, "data")
		}

		if exportBuffer.Len() > 0 {
			if _, err := exportBuffer.WriteTo(io.Export); err != nil {
				return 0, 0, err
			}

			fmt.Fprintf(io.UserOutput, msg)
		}
	}

	return restoredCount, fileCount, nil
}

func getPath(root, filepath, version string) string {

	if len(version) > 0 {
		return fmt.Sprintf("%s (%s)", path.BuildPath(root, filepath), version)
	}

	return path.BuildPath(root, filepath)
}

func compatibleFormat(format, fileType string) bool {

	switch format {
	case "task-def-secrets", "task-def-env", "terminal-export", "json-object":
		return fileType == "env"
	case "":
		return true
	default:
		return false
	}
}

const (
	exportToken      = "export"
	formatToken      = "format"
	injectToken      = "inject-secrets"
	modifyToken      = "modify-secrets"
	noOverwriteToken = "no-overwrite"
)

func init() {
	RootCmd.AddCommand(pullCmd)

	pullCmd.Flags().BoolVarP(&uo.ExportEnv, exportToken, "e", false, "Append export command to environment variables and send to stdout.")
	pullCmd.Flags().StringVarP(&uo.Tags, tagsToken, "t", "", "Specify a list of tags used to filter files.")
	pullCmd.Flags().StringVarP(&uo.ExportFormat, formatToken, "g", "", "Format environment variables and send to stdout")
	pullCmd.Flags().StringVarP(&uo.Version, "ver", "v", "", "Set a version to identify a file specific state.")
	pullCmd.Flags().BoolVarP(&uo.InjectSecrets, injectToken, "i", false, "Generate *.secrets file containing configuration including secrets.")
	pullCmd.Flags().BoolVarP(&uo.ModifySecrets, modifyToken, "m", false, "Pulls configuration with secret tokens and secrets.")
	pullCmd.Flags().StringVarP(&uo.AlternateRestorePath, altToken, "a", "", "Set an alternate path to clone the file to during a restore.")
	pullCmd.Flags().BoolVarP(&uo.NoOverwrite, noOverwriteToken, "n", false, "Only pulls the environment variables that are not exported in the current environment.")

	viper.BindPFlag(exportToken, RootCmd.PersistentFlags().Lookup(exportToken))
	viper.BindPFlag(tagsToken, RootCmd.PersistentFlags().Lookup(tagsToken))
	viper.BindPFlag(formatToken, RootCmd.PersistentFlags().Lookup(formatToken))
	viper.BindPFlag(injectToken, RootCmd.PersistentFlags().Lookup(injectToken))
	viper.BindPFlag(modifyToken, RootCmd.PersistentFlags().Lookup(modifyToken))
	viper.BindPFlag(altToken, RootCmd.PersistentFlags().Lookup(altToken))
	viper.BindPFlag(noOverwriteToken, RootCmd.PersistentFlags().Lookup(noOverwriteToken))
}
