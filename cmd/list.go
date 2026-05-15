package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/vieolo/termange"
	"github.com/vieolo/uch/internal/config"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all stored commands",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _, err := config.Load()
		must(err, "Could not load config")

		if len(cfg.Commands) == 0 {
			termange.PrintInfoln("No commands stored yet. Add one with 'uch add <name>'.")
			return
		}

		names := make([]string, 0, len(cfg.Commands))
		for name := range cfg.Commands {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			entry := cfg.Commands[name]
			marker := ""
			if entry.Sensitive {
				marker = " [sensitive]"
			}
			body := entry.Cmd
			if entry.Sensitive {
				body = "<encrypted>"
			}
			fmt.Printf("  %s%s\n", name, marker)
			if entry.Description != "" {
				fmt.Printf("      %s\n", entry.Description)
			}
			fmt.Printf("      %s\n", body)
			if len(entry.Variables) > 0 {
				fmt.Printf("      vars: ")
				for i, v := range entry.Variables {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Printf("%s(%s)", v.Name, v.Type)
				}
				fmt.Println()
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
