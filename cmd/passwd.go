package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vieolo/filange"
	"github.com/vieolo/termange"
	"github.com/vieolo/uch/internal/config"
	"github.com/vieolo/uch/internal/crypto"
	"github.com/vieolo/uch/internal/prompt"
)

var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Change the master password",
	Long: `Re-encrypts ~/.uch/identity.age with a new password.
All existing sensitive commands remain valid; only the identity file changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, paths, err := config.Load()
		must(err, "Could not load config")

		if !filange.FileExists(paths.IdentityPath) {
			fail("Encryption is not initialized. Run 'uch init' first.")
		}

		oldPw, err := prompt.Password("Current password:")
		must(err, "Could not read password")

		// Verify old password before asking for a new one.
		if _, err := crypto.LoadIdentity(paths.IdentityPath, oldPw); err != nil {
			fail("%v", err)
		}

		newPw, err := prompt.PasswordWithConfirm("New password:")
		must(err, "Could not read new password")

		must(crypto.ChangePassword(paths.IdentityPath, oldPw, newPw), "Could not change password")
		termange.PrintSuccessln("Master password changed.")
	},
}

func init() {
	rootCmd.AddCommand(passwdCmd)
}
