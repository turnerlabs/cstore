package cmd

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/display"
	localFile "github.com/turnerlabs/cstore/components/file"
	"github.com/turnerlabs/cstore/components/logger"
	"github.com/turnerlabs/cstore/components/models"
	"github.com/turnerlabs/cstore/components/path"
	"github.com/turnerlabs/cstore/components/prompt"
	"github.com/turnerlabs/cstore/components/remote"
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
			fmt.Fprintf(io.UserOutput, "Linking %s   %s \n", filePath, checkMark)
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
		remoteComp, err := remote.InitComponents(&fileEntry, clog, opt, io)
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
		color.New(color.Bold).Fprintf(io.UserOutput, remoteComp.Store.Name())
		fmt.Fprintln(io.UserOutput, "]")

		//--------------------------------------------------------
		//- Ensure file has not been modified by another user.
		//--------------------------------------------------------
		if lastModified, err := remoteComp.Store.Changed(&fileEntry, file, opt.Version); err == nil {
			if !fileEntry.IsCurrent(lastModified, clog.Context) {
				if !prompt.Confirm(fmt.Sprintf("Remotely stored data '%s' was modified on %s. Overwrite?", filePath, lastModified.Format("01/02/06")), prompt.Warn, io) {
					fmt.Fprintf(io.UserOutput, "Skipping %s\n", filePath)
					continue
				}
			}
		}

		//----------------------------------------------------
		//- If user specified, push secrets to secret store.
		//----------------------------------------------------
		if fileEntry.SupportsSecrets() {

			tokens, err := token.Find(file, fileEntry.Type, true)
			if err != nil {
				display.Error(fmt.Errorf("Failed to parse tokens in file %s. (%s)", filePath, err), io.UserOutput)
				continue
			}

			if len(tokens) > 0 {

				for _, t := range tokens {
					if err := remoteComp.Secrets.Set(clog.Context, t.Secret(), t.Prop, t.Value); err != nil {
						logger.L.Fatal(err)
					}
				}

				file = token.RemoveSecrets(file)

				if err = localFile.Save(clog.GetFullPath(fileEntry.Path), file); err != nil {
					logger.L.Print(err)
				}
			}
		}

		//-------------------------------------------------
		//- Validate version and file version data.
		//-------------------------------------------------
		if len(opt.Version) > 0 {
			if !remoteComp.Store.SupportsFeature(store.VersionFeature) {
				display.Error(fmt.Errorf("%s store does not support %s", remoteComp.Store.Name(), store.VersionFeature), io.UserOutput)
				continue
			}

			if fileEntry.Missing(opt.Version) {
				fileEntry.Versions = append(fileEntry.Versions, opt.Version)
			}
		}

		//-------------------------------------------------
		//- Push file to file store.
		//-------------------------------------------------
		if err = remoteComp.Store.Push(&fileEntry, file, opt.Version); err != nil {
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

const (
	storeToken  = "store"
	deleteToken = "delete"
	tagsToken   = "tags"
	altToken    = "alt"
)

func init() {
	RootCmd.AddCommand(pushCmd)

	pushCmd.Flags().StringVarP(&uo.Store, storeToken, "s", "", "Set the context store used to store files. The 'stores' command lists options.")
	pushCmd.Flags().StringVarP(&uo.DeleteLocalFiles, deleteToken, "d", "", "Delete the local file after any successful pushes.")
	pushCmd.Flags().StringVarP(&uo.Tags, tagsToken, "t", "", "Set a list of tags used to identify the file.")
	pushCmd.Flags().StringVarP(&uo.Version, "ver", "v", "", "Set a version to identify the file current state.")
	pushCmd.Flags().StringVarP(&uo.AlternateRestorePath, altToken, "a", "", "Set an alternate path to clone the file to during a restore.")

	viper.BindPFlag(storeToken, RootCmd.PersistentFlags().Lookup(storeToken))
	viper.BindPFlag(deleteToken, RootCmd.PersistentFlags().Lookup(deleteToken))
	viper.BindPFlag(tagsToken, RootCmd.PersistentFlags().Lookup(tagsToken))
	viper.BindPFlag(altToken, RootCmd.PersistentFlags().Lookup(altToken))
}
