package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	localFile "github.com/turnerlabs/cstore/components/file"
	"github.com/turnerlabs/cstore/components/logger"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/path"
	"github.com/turnerlabs/cstore/components/prompt"
	"github.com/turnerlabs/cstore/components/store"
	"github.com/turnerlabs/cstore/components/token"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Store file(s) remotely.",
	Long:  `Store file(s) remotely.`,
	Run: func(cmd *cobra.Command, userSpecifiedFilePaths []string) {
		setupUserOptions(userSpecifiedFilePaths)

		if err := Push(uo, ioStreams); err != nil {
			fmt.Fprintf(ioStreams.UserOutput, "%sERROR:%s ", uo.Format.Red, uo.Format.NoColor)
			logger.L.Fatalf("%s\n\n", err)
		}
	},
}

// Push ...
func Push(opt cfg.UserOptions, io models.IO) error {
	filesPushed := []string{}
	fileCount := 0
	errorOccurred := false

	//-------------------------------------------------
	//- Get or create the local catalog for push.
	//-------------------------------------------------
	clog, err := catalog.GetMake(opt.Catalog, io)
	if err != nil {
		return err
	}

	//-------------------------------------------------
	//- Process each file the user wants to push.
	//-------------------------------------------------
	fmt.Fprintln(io.UserOutput)
	for _, filePath := range getFilePathsToPush(clog, opt) {

		file, err := localFile.GetBy(clog.GetFullPath(filePath))
		if err != nil {
			fmt.Fprintf(io.UserOutput, "%sERROR:%s %s\n\n", opt.Format.Red, opt.Format.NoColor, err)
			errorOccurred = true
			continue
		}

		fileEntry, _ := clog.LookupEntry(filePath, file)

		//-------------------------------------------------
		//- Set file options based on command line flags
		//-------------------------------------------------
		fileEntry = updateUserOptions(fileEntry, opt)

		//-------------------------------------------------
		//- If file is a catalog, link it to this catalog.
		//-------------------------------------------------
		if fileEntry.IsRef {
			fmt.Fprintf(io.UserOutput, "Linking %s   %s \n", fileEntry.Path, checkMark)
			if err := clog.UpdateEntry(fileEntry); err != nil {
				fmt.Fprintf(io.UserOutput, "%sERROR:%s %s\n\n", opt.Format.Red, opt.Format.NoColor, err)
				errorOccurred = true
			}
			continue
		} else {
			fileCount++
		}

		//--------------------------------------------------
		//- Get the remote store and vault components ready.
		//--------------------------------------------------
		remoteComp, err := getRemoteComponents(&fileEntry, clog, opt.Prompt, io)
		if err != nil {
			logger.L.Print(err)
			errorOccurred = true
			continue
		}

		//--------------------------------------------------
		//- Begin push process.
		//--------------------------------------------------
		if len(opt.Version) > 0 {
			fmt.Fprintf(io.UserOutput, "Pushing [%s%s%s](%s) -> [%s%s%s]\n", opt.Format.Blue, fileEntry.Path, opt.Format.NoColor, opt.Version, opt.Format.Bold, remoteComp.store.Name(), opt.Format.UnBold)
		} else {
			fmt.Fprintf(io.UserOutput, "Pushing [%s%s%s] -> [%s%s%s]\n", opt.Format.Blue, fileEntry.Path, opt.Format.NoColor, opt.Format.Bold, remoteComp.store.Name(), opt.Format.UnBold)
		}

		//--------------------------------------------------------
		//- Ensure file has not been modified by another user.
		//--------------------------------------------------------
		if lastModified, err := remoteComp.store.Changed(&fileEntry, file, opt.Version); err != nil {
			fmt.Fprintf(io.UserOutput, "%sERROR:%s Failed to determine when '%s' version %s was last modified.\n\n", opt.Format.Red, opt.Format.NoColor, filePath, opt.Version)
			logger.L.Print(err)
			errorOccurred = true
			continue
		} else {
			if !fileEntry.IsCurrent(lastModified, clog.Context) {
				if !prompt.Confirm(fmt.Sprintf("Remote file '%s' was modified on %s. Overwrite?", filePath, lastModified.Format(time.RFC822)), false, io) {
					fmt.Fprintf(io.UserOutput, "Skipping %s\n", filePath)
					errorOccurred = true
					continue
				}
			}
		}

		//----------------------------------------------------
		//- If user specified, push secrets to secret store.
		//----------------------------------------------------
		if opt.ModifySecrets {
			if !fileEntry.SupportsSecrets() {
				fmt.Fprintf(io.UserOutput, "%sERROR:%s Secrets not supported for %s due to incompatible file type %s.\n\n", opt.Format.Red, opt.Format.NoColor, filePath, fileEntry.Type)
				errorOccurred = true
				continue
			}

			tokens, err := token.Find(file, fileEntry.Type, true)
			if err != nil {
				fmt.Fprintf(io.UserOutput, "%sERROR:%s Failed to find tokens in file %s.\n\n", opt.Format.Red, opt.Format.NoColor, filePath)
				logger.L.Print(err)
				errorOccurred = true
				continue
			}

			if len(tokens) == 0 {
				fmt.Fprintf(io.UserOutput, "%sERROR:%s To set secrets, tokens in %s must be in the format %s{{ENV/TOKEN::VALUE}}%s. Learn about additional limitations at https://github.com/turnerlabs/cstore/blob/master/docs/SECRETS.md.\n\n", opt.Format.Red, opt.Format.NoColor, filePath, opt.Format.Bold, opt.Format.UnBold)
			}

			for _, t := range tokens {
				if err := remoteComp.secrets.Set(clog.Context, t.Secret(), t.Prop, t.Value); err != nil {
					logger.L.Fatal(err)
					errorOccurred = true
				}
			}

			file = token.RemoveSecrets(file)

			if err = localFile.Save(clog.GetFullPath(fileEntry.Path), file); err != nil {
				logger.L.Print(err)
				errorOccurred = true
			}
		}

		//-------------------------------------------------
		//- Validate version and file version data.
		//-------------------------------------------------
		if len(opt.Version) > 0 {
			if !remoteComp.store.Supports(store.VersionFeature) {
				fmt.Fprintf(io.UserOutput, "%sERROR:%s %s store does not support %s feature.\n\n", opt.Format.Red, opt.Format.NoColor, remoteComp.store.Name(), store.VersionFeature)
				errorOccurred = true
				continue
			}

			if fileEntry.Missing(opt.Version) {
				fileEntry.Versions = append(fileEntry.Versions, opt.Version)
			}
		}

		//-------------------------------------------------
		//- Push file to file store.
		//-------------------------------------------------
		updated := time.Now()
		if err = remoteComp.store.Push(&fileEntry, file, opt.Version); err != nil {
			logger.L.Print(err)
			errorOccurred = true
			continue
		}

		//-------------------------------------------------
		//- Update the catalog with file entry changes.
		//-------------------------------------------------
		if err := clog.UpdateEntry(fileEntry); err != nil {
			logger.L.Print(err)
			errorOccurred = true
			continue
		}

		//-------------------------------------------------
		//- Save the time the user last pulled file.
		//-------------------------------------------------
		if err := clog.RecordPull(fileEntry.Key(), updated); err != nil {
			logger.L.Print(err)
			errorOccurred = true
			continue
		}

		//---------------------------------------------------------------------
		//- Create the ghost .cstore reference file when not in cStore.yml dir.
		//---------------------------------------------------------------------
		justThePath := path.RemoveFileName(filePath)

		if len(clog.GetFullPath(justThePath)) > 0 {
			if err := catalog.WriteGhost(clog.GetFullPath(justThePath), catalog.Ghost{
				Location: justThePath,
			}); err != nil {
				logger.L.Print(err)
				errorOccurred = true
			}
		}

		filesPushed = append(filesPushed, fileEntry.Path)
	}

	//-------------------------------------------------
	//- Save the catalog with updated files locally.
	//-------------------------------------------------
	if err := catalog.Write(clog.GetFullPath(opt.Catalog), clog); err != nil {
		return err
	}

	//-------------------------------------------------
	//- If user specific, delete local files.
	//-------------------------------------------------
	if opt.DeleteLocalFiles {
		for _, path := range filesPushed {
			os.Remove(clog.GetFullPath(path))
			os.Remove(fmt.Sprintf("%s.secrets", clog.GetFullPath(path)))
		}
	}

	fmt.Fprintf(io.UserOutput, "\n%s%d of %d file(s) pushed to remote store.%s\n\n", opt.Format.Bold, len(filesPushed), fileCount, opt.Format.UnBold)

	if errorOccurred {
		return errors.New("issues were encountered for some files")
	}

	return nil
}

func init() {
	RootCmd.AddCommand(pushCmd)

	pushCmd.Flags().StringVarP(&uo.Store, "store", "s", "", "Set the context store used to store files. The 'stores' command lists options.")
	pushCmd.Flags().BoolVarP(&uo.DeleteLocalFiles, "delete", "d", false, "Delete the local file after pushing.")
	pushCmd.Flags().StringVarP(&uo.Tags, "tags", "t", "", "Set a list of tags used to identify the file.")
	pushCmd.Flags().StringVarP(&uo.Version, "ver", "v", "", "Set a version to identify the file current state.")
	pushCmd.Flags().StringVarP(&uo.AlternateRestorePath, "alt", "a", "", "Set an alternate path to clone the file to during a restore.")
	pushCmd.Flags().BoolVarP(&uo.ModifySecrets, "modify-secrets", "m", false, "Store secrets for tokens in file.")
}
