package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vieolo/filange"
	"github.com/vieolo/termange"
	"github.com/vieolo/uch/internal/config"
	"github.com/vieolo/uch/internal/crypto"
	"github.com/vieolo/uch/internal/prompt"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the encryption identity for sensitive commands",
	Long: `Sets a master password and creates the age identity at ~/.uch/identity.age.
The identity is required to store or run any command marked sensitive: true.
Run this once. Use 'uch passwd' to change the password later.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, paths, err := config.Load()
		if err != nil {
			termange.PrintErrorf("Could not load config: %v\n", err)
			os.Exit(1)
		}

		if filange.FileExists(paths.IdentityPath) {
			termange.PrintWarningf("Identity already exists at %s\n", paths.IdentityPath)
			termange.PrintInfoln("Use 'uch passwd' to change the password.")
			os.Exit(1)
		}

		termange.PrintInfoln("Pick a master password. You will be asked for it every time you run a sensitive command.")
		pw, err := prompt.PasswordWithConfirm("New password:")
		if err != nil {
			termange.PrintErrorf("%v\n", err)
			os.Exit(1)
		}

		if err := crypto.GenerateIdentity(paths.IdentityPath, pw); err != nil {
			termange.PrintErrorf("Could not create identity: %v\n", err)
			os.Exit(1)
		}

		cfg, _, err := config.Load()
		if err != nil {
			termange.PrintErrorf("Could not reload config: %v\n", err)
			os.Exit(1)
		}
		cfg.Encryption.Enabled = true
		if err := config.Save(cfg); err != nil {
			termange.PrintErrorf("Could not update config: %v\n", err)
			os.Exit(1)
		}

		termange.PrintSuccessf("Identity created at %s\n", paths.IdentityPath)
		termange.PrintWarningln("Back up this file. If you lose it, sensitive commands cannot be recovered.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
