package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vieolo/uch/internal/config"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "uch",
	Short: "Centralize storage and execution of commands",
	Long:  `uch helps you to store the commands that you commonly re-use and execute them using their nicknames`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
//
// Version-aware setup runs here (not in init) because main.go injects
// ThisGyByte after package init has finished.
func Execute() {
	v := currentVersion()
	rootCmd.Version = v
	config.SetCreatedBy("uch " + v)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}



