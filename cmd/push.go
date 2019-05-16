package cmd

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/display"
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
			display.Error(fmt.Sprintf("%s\n", err), ioStreams.UserOutput)
			os.Exit(1)
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
			display.Error(err.Error(), io.UserOutput)
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
				display.Error(err.Error(), io.UserOutput)
				errorOccurred = true
			}
			continue
		} else {
			fileCount++
		}

		//--------------------------------------------------
		//- Get the remote store and vault components ready.
		//--------------------------------------------------
		remoteComp, err := getRemoteComponents(&fileEntry, clog, opt, io)
		if err != nil {
			logger.L.Print(err)
			errorOccurred = true
			continue
		}

		//--------------------------------------------------
		//- Begin push process.
		//--------------------------------------------------
		fmt.Fprint(io.UserOutput, "Pushing [")
		color.New(color.FgBlue).Fprintf(io.UserOutput, fileEntry.Path)
		fmt.Fprint(io.UserOutput, "]")

		if len(opt.Version) > 0 {
			fmt.Fprintf(io.UserOutput, "(%s)", opt.Version)
		}

		fmt.Fprint(io.UserOutput, " -> [")
		color.New(color.Bold).Fprintf(io.UserOutput, remoteComp.store.Name())
		fmt.Fprintln(io.UserOutput, "]")

		//--------------------------------------------------------
		//- Ensure file has not been modified by another user.
		//--------------------------------------------------------
		if lastModified, err := remoteComp.store.Changed(&fileEntry, file, opt.Version); err != nil {
			display.Error(fmt.Sprintf("Failed to determine when '%s' version %s was last modified.\n", filePath, opt.Version), io.UserOutput)
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
				display.Error(fmt.Sprintf("Secrets not supported for %s due to incompatible file type %s.\n", filePath, fileEntry.Type), io.UserOutput)
				errorOccurred = true
				continue
			}

			tokens, err := token.Find(file, fileEntry.Type, true)
			if err != nil {
				display.Error(fmt.Sprintf("Failed to find tokens in file %s.\n", filePath), io.UserOutput)
				logger.L.Print(err)
				errorOccurred = true
				continue
			}

			if len(tokens) == 0 {
				display.Error(fmt.Sprintf("To set secrets, tokens in %s must be in the format {{ENV/TOKEN::VALUE}}. Learn about additional limitations at https://github.com/turnerlabs/cstore/blob/master/docs/SECRETS.md.\n", filePath), io.UserOutput)
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
				display.Error(fmt.Sprintf("%s store does not support %s feature.\n", remoteComp.store.Name(), store.VersionFeature), io.UserOutput)
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
		if err = remoteComp.store.Push(&fileEntry, file, opt.Version); err != nil {
			display.Error(err.Error(), io.UserOutput)
			errorOccurred = true
			continue
		}

		//-------------------------------------------------
		//- Update the catalog with file entry changes.
		//-------------------------------------------------
		if err := clog.UpdateEntry(fileEntry); err != nil {
			display.Error(err.Error(), io.UserOutput)
			errorOccurred = true
			continue
		}

		//-------------------------------------------------
		//- Save the time the user last pulled file.
		//-------------------------------------------------
		if err := clog.RecordPull(fileEntry.Key(), time.Now().Add(time.Second*1)); err != nil {
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
	//- Locally save the catalog with updated files.
	//-------------------------------------------------
	original, _ := catalog.Get(opt.Catalog)

	if !reflect.DeepEqual(original, clog) {
		if err := catalog.Write(clog.GetFullPath(opt.Catalog), clog); err != nil {
			return err
		}
	}

	//-------------------------------------------------
	//- If user specified, delete local files.
	//-------------------------------------------------
	if opt.DeleteLocalFiles {
		for _, path := range filesPushed {
			os.Remove(clog.GetFullPath(path))
			os.Remove(fmt.Sprintf("%s.secrets", clog.GetFullPath(path)))
		}
	}

	color.New(color.Bold).Fprintf(io.UserOutput, "\n%d of %d file(s) pushed to remote store.\n\n", len(filesPushed), fileCount)

	if errorOccurred {
		return errors.New("push failed")
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
