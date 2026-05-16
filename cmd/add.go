package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/vieolo/filange"
	"github.com/vieolo/termange"
	"github.com/vieolo/termange/tui"
	"github.com/vieolo/uch/internal/config"
	"github.com/vieolo/uch/internal/crypto"
	"github.com/vieolo/uch/internal/prompt"
)

var (
	addSensitive   bool
	addCmd_string  string
	addDescription string
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new command interactively",
	Long: `Adds a command under the given nickname.
Pass --cmd to skip the prompt for the command body.
Pass --sensitive to encrypt the command (requires 'uch init' first).
Variables can be added afterwards by editing the file with 'uch edit'.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.TrimSpace(args[0])
		if name == "" {
			fail("Command name cannot be empty")
		}

		cfg, paths, err := config.Load()
		must(err, "Could not load config")

		if _, exists := cfg.Commands[name]; exists {
			fail("Command %q already exists. Edit it with 'uch edit' or remove it first.", name)
		}

		body := addCmd_string
		if body == "" {
			body = tui.TextInput(tui.TextInputOptions{Prompt: "Command:"})
		}
		if strings.TrimSpace(body) == "" {
			fail("Command body cannot be empty")
		}

		entry := config.Command{Description: addDescription}

		if addSensitive {
			if !filange.FileExists(paths.IdentityPath) {
				fail("Encryption is not initialized. Run 'uch init' first.")
			}
			pw, err := prompt.Password("Master password:")
			must(err, "Could not read password")
			id, err := crypto.LoadIdentity(paths.IdentityPath, pw)
			must(err, "Could not unlock identity")

			cipher, err := crypto.EncryptString(body, id)
			must(err, "Could not encrypt command")
			entry.Sensitive = true
			entry.CmdEncrypted = cipher
		} else {
			entry.Cmd = body
		}

		cfg.Commands[name] = entry
		must(config.Save(cfg), "Could not save config")

		termange.PrintSuccessf("Added %q\n", name)
		if !addSensitive {
			termange.PrintInfoln("To add variables (e.g. {{platform}}), edit the file with 'uch edit'.")
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolVarP(&addSensitive, "sensitive", "s", false, "Encrypt the command with the master password")
	addCmd.Flags().StringVarP(&addCmd_string, "cmd", "c", "", "The command body (skips the prompt)")
	addCmd.Flags().StringVarP(&addDescription, "description", "d", "", "Optional description")
}
