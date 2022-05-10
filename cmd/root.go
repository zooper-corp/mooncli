package cmd

import (
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "mooncli",
	Short:            "A set of command line utilities for the Moonbeam project",
	PersistentPreRun: initLog,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initLog(cmd *cobra.Command, args []string) {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if !verbose {
		log.SetOutput(ioutil.Discard)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show logs")
	rootCmd.PersistentFlags().String(
		"chain",
		"moonbeam",
		"Chain endpoint url or network name [moonbeam,moonriver,moonbase]",
	)
}
