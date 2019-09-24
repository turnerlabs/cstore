package cmd

import (
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
			display.Error(err, ioStreams.UserOutput)
			os.Exit(1)
		}
	},
}

// Push ...
func Push(opt cfg.UserOptions, io models.IO) error {
	filesPushed := []string{}
	fileCount := 0

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
			display.Error(err, io.UserOutput)
			continue
		}

		fileEntry, update := clog.LookupEntry(filePath, file)

		//-------------------------------------------------
		//- Set file options based on command line flags
		//-------------------------------------------------
		fileEntry = updateUserOptions(fileEntry, update, opt)

		//-------------------------------------------------
		//- If file is a catalog, link it to this catalog.
		//-------------------------------------------------
		if fileEntry.IsRef {
			fmt.Fprintf(io.UserOutput, "Linking %s   %s \n", fileEntry.Path, checkMark)
			if err := clog.UpdateEntry(fileEntry); err != nil {
				display.Error(err, io.UserOutput)
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
			display.Error(err, io.UserOutput)
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
			display.Error(fmt.Errorf("Failed to determine when '%s' version %s was last modified. (%s)", filePath, opt.Version, err), io.UserOutput)
			continue
		} else {
			if !fileEntry.IsCurrent(lastModified, clog.Context) {
				if !prompt.Confirm(fmt.Sprintf("Remote file '%s' was modified on %s. Overwrite?", filePath, lastModified.Format(time.RFC822)), prompt.Warn, io) {
					fmt.Fprintf(io.UserOutput, "Skipping %s\n", filePath)
					continue
				}
			}
		}

		//----------------------------------------------------
		//- If user specified, push secrets to secret store.
		//----------------------------------------------------
		if opt.ModifySecrets {
			if !fileEntry.SupportsSecrets() {
				display.Error(fmt.Errorf("Secrets not supported for %s due to incompatible file type %s.", filePath, fileEntry.Type), io.UserOutput)
				continue
			}

			tokens, err := token.Find(file, fileEntry.Type, true)
			if err != nil {
				display.Error(fmt.Errorf("Failed to find tokens in file %s. (%s)", filePath, err), io.UserOutput)
				continue
			}

			if len(tokens) == 0 {
				display.Error(fmt.Errorf("To set secrets, tokens in %s must be in the format {{ENV/TOKEN::VALUE}}. Learn about additional limitations at https://github.com/turnerlabs/cstore/blob/master/docs/SECRETS.md.", filePath), io.UserOutput)
			}

			for _, t := range tokens {
				if err := remoteComp.secrets.Set(clog.Context, t.Secret(), t.Prop, t.Value); err != nil {
					logger.L.Fatal(err)
				}
			}

			file = token.RemoveSecrets(file)

			if err = localFile.Save(clog.GetFullPath(fileEntry.Path), file); err != nil {
				logger.L.Print(err)
			}
		}

		//-------------------------------------------------
		//- Validate version and file version data.
		//-------------------------------------------------
		if len(opt.Version) > 0 {
			if !remoteComp.store.SupportsFeature(store.VersionFeature) {
				display.Error(fmt.Errorf("%s store does not support %s feature.", remoteComp.store.Name(), store.VersionFeature), io.UserOutput)
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
			display.Error(err, io.UserOutput)
			continue
		}

		//-------------------------------------------------
		//- Update the catalog with file entry changes.
		//-------------------------------------------------
		if err := clog.UpdateEntry(fileEntry); err != nil {
			display.Error(err, io.UserOutput)
			continue
		}

		//-------------------------------------------------
		//- Save the time the user last pulled file.
		//-------------------------------------------------
		if err := clog.RecordPull(fileEntry.Key(), time.Now().Add(time.Second*1)); err != nil {
			logger.L.Print(err)
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
			}
		}

		filesPushed = append(filesPushed, fileEntry.Path)

		//-------------------------------------------------
		//- If user specified, delete local files.
		//-------------------------------------------------
		if fileEntry.DeleteAfterPush {
			os.Remove(clog.GetFullPath(fileEntry.Path))
			os.Remove(fmt.Sprintf("%s.secrets", clog.GetFullPath(fileEntry.Path)))
		}

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

	color.New(color.Bold).Fprintf(io.UserOutput, "\n%d of %d file(s) pushed to remote store.\n\n", len(filesPushed), fileCount)

	return nil
}

func init() {
	RootCmd.AddCommand(pushCmd)

	pushCmd.Flags().StringVarP(&uo.Store, "store", "s", "", "Set the context store used to store files. The 'stores' command lists options.")
	pushCmd.Flags().StringVarP(&uo.DeleteLocalFiles, "delete", "d", "", "Delete the local file after any successful pushes.")
	pushCmd.Flags().StringVarP(&uo.Tags, "tags", "t", "", "Set a list of tags used to identify the file.")
	pushCmd.Flags().StringVarP(&uo.Version, "ver", "v", "", "Set a version to identify the file current state.")
	pushCmd.Flags().StringVarP(&uo.AlternateRestorePath, "alt", "a", "", "Set an alternate path to clone the file to during a restore.")
	pushCmd.Flags().BoolVarP(&uo.ModifySecrets, "modify-secrets", "m", false, "Store secrets for tokens in file.")
}
