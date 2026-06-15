package cmd

import (
	"fmt"
	"os"
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
	Short: "Add a new command to the config",
	Long: `Adds a command under the given nickname.

Pass the nickname of the command as the argument.

Pass --cmd to skip the prompt for the command body.
Pass --sensitive to encrypt the command (requires 'uch init' first).
Variables can be added afterwards by editing the file with 'uch edit'.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// Checking if name is not empty
		name := strings.ToLower(strings.TrimSpace(args[0]))
		if name == "" {
			fail("Please provide the nickname of the command you wish to add")
		}

		// Load the existing config
		cfg, paths, err := config.Load()
		must(err, "Could not load config")

		// Checking if the command already exists or not
		if _, exists := cfg.Commands[name]; exists {
			termange.PrintErrorf("Command %q already exists. You have two options:\n", name)
			fmt.Println(" - Edit it with 'uch edit'")
			fmt.Println(" - Remove it first with 'uch remove'")
			os.Exit(1)
		}

		// Prompting the
		body := addCmd_string
		if body == "" {
			body = tui.TextInput(tui.TextInputOptions{Prompt: "The command to be added:"})
		}
		if strings.TrimSpace(body) == "" {
			fail("Command body cannot be empty")
		}

		// Setting the optional description
		entry := config.Command{Description: addDescription}

		if addSensitive {
			// Encrypting the command if the sensitive flag is provided

			// Checking if the encryption is initialized or not
			if !filange.FileExists(paths.IdentityPath) {
				fail("You have passed the --sensitive flag which will encrypt the command. However, encryption is not initialized. Run 'uch init' first")
			}

			// Getting the password from the user
			pw, err := prompt.Password("uch password:")
			must(err, "Could not read password")
			id, err := crypto.LoadIdentity(paths.IdentityPath, pw)
			must(err, "Could not unlock identity!")

			cipher, err := crypto.EncryptString(body, id)
			must(err, "Could not encrypt command")
			entry.Sensitive = true
			entry.CmdEncrypted = cipher
		} else {
			// The command is not encrypted and it is saved as plain text
			entry.Cmd = body
		}

		// Saving the config file
		cfg.Commands[name] = entry
		must(config.Save(cfg), "Could not save config!")

		termange.PrintSuccessf("Added %q\n", name)
		if !addSensitive {
			termange.PrintInfoln("To add variables (e.g. {{platform}}), edit the file with 'uch edit'.")
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolVarP(&addSensitive, "sensitive", "s", false, "Encrypt the command with the uch password")
	addCmd.Flags().StringVarP(&addCmd_string, "cmd", "c", "", "The command body (skips the prompt)")
	addCmd.Flags().StringVarP(&addDescription, "description", "d", "", "Optional description")
}
