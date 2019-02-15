package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turnerlabs/cstore/components/catalog"
	"github.com/turnerlabs/cstore/components/cfg"
	"github.com/turnerlabs/cstore/components/models"
)

const (
	secretsToken = "secrets"
	accessToken  = "access"
	catalogToken = "catalog"
	promptToken  = "prompt"
	loggingToken = "logging"
	commandToken = "store-command"
)

var (
	cfgFile   string
	uo        cfg.UserOptions
	ioStreams = models.IO{
		UserOutput: color.Output,
		UserInput:  os.Stdin,
		Export:     os.Stdout,
	}
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringP(accessToken, "c", "", "Set the vault used to get store credentials and encryption keys. The 'vaults' command lists options.")
	RootCmd.PersistentFlags().StringP(secretsToken, "x", "", "Set the vault used to get and store secrets. The 'vaults' command lists options.")
	RootCmd.PersistentFlags().StringP(catalogToken, "f", catalog.DefaultFileName, "Catalog file to use for current command.")
	RootCmd.PersistentFlags().StringP(commandToken, "", "", "Command to send to the store.")
	RootCmd.PersistentFlags().BoolP(promptToken, "p", false, "Prompt user for configuration.")
	RootCmd.PersistentFlags().BoolP(loggingToken, "l", false, "Set the format of the output to be log friendly instead of terminal friendly.")

	viper.BindPFlag(catalogToken, RootCmd.PersistentFlags().Lookup(catalogToken))
	viper.BindPFlag(secretsToken, RootCmd.PersistentFlags().Lookup(secretsToken))
	viper.BindPFlag(accessToken, RootCmd.PersistentFlags().Lookup(accessToken))
	viper.BindPFlag(promptToken, RootCmd.PersistentFlags().Lookup(promptToken))
	viper.BindPFlag(loggingToken, RootCmd.PersistentFlags().Lookup(loggingToken))
	viper.BindPFlag(commandToken, RootCmd.PersistentFlags().Lookup(commandToken))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("user")          // name of config file (without extension)
	viper.AddConfigPath("$HOME/.cstore") // adding home directory as first search path
	viper.AutomaticEnv()                 // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func setupUserOptions(userSpecifiedFilePaths []string) {
	uo.Catalog = viper.GetString(catalogToken)
	uo.SecretsVault = viper.GetString(secretsToken)
	uo.AccessVault = viper.GetString(accessToken)
	uo.Prompt = viper.GetBool(promptToken)
	uo.StoreCommand = viper.GetString(commandToken)

	uo.AddPaths(userSpecifiedFilePaths)
	uo.ParseTags()

	if viper.GetBool(loggingToken) {
		color.NoColor = true
	}
}
