package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vieolo/termange"
	"github.com/vieolo/termange/tui"
	"github.com/vieolo/uch/internal/config"
)

var removeForce bool

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm"},
	Short:   "Remove a stored command",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, _, err := config.Load()
		must(err, "Could not load config")

		entry, ok := cfg.Commands[name]
		if !ok {
			fail("Command %q does not exist", name)
		}

		if !removeForce {
			prompt := "Remove " + name + "?"
			if entry.Sensitive {
				prompt = "Remove sensitive command " + name + "? (the encrypted body will be lost)"
			}
			if !tui.Confirm(tui.ConfirmOptions{Prompt: prompt}) {
				termange.PrintInfoln("Cancelled.")
				return
			}
		}

		delete(cfg.Commands, name)
		must(config.Save(cfg), "Could not save config")
		termange.PrintSuccessf("Removed %q\n", name)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Skip the confirmation prompt")
}
