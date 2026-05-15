package cmd

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/vieolo/termange"
	"github.com/vieolo/uch/internal/config"
	"github.com/vieolo/uch/internal/crypto"
	"github.com/vieolo/uch/internal/prompt"
)

var runCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Run a stored command by its nickname",
	Long: `Runs the command stored under <name>.
If the command is sensitive, you will be prompted for the master password.
If the command defines variables, you will be prompted for each one before execution.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, paths, err := config.Load()
		must(err, "Could not load config")

		entry, ok := cfg.Commands[name]
		if !ok {
			fail("Command %q does not exist", name)
		}

		// Resolve the command body (decrypt if needed).
		body := entry.Cmd
		if entry.Sensitive {
			id := requireIdentity(paths)
			body, err = crypto.DecryptString(entry.CmdEncrypted, id)
			must(err, "Could not decrypt command")
		}
		if body == "" {
			fail("Command %q has an empty body", name)
		}

		// Resolve variables.
		values, err := prompt.ResolveVariables(entry.Variables)
		must(err, "Could not resolve variables")

		final, err := prompt.Substitute(body, values)
		must(err, "Could not substitute variables")

		// Execute via /bin/sh -c so the user can use shell features (pipes, redirects).
		// stdin/stdout/stderr are wired to the terminal so interactive commands work.
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		termange.PrintInfof("$ %s\n", final)
		c := exec.Command(shell, "-c", final)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			// Surface the underlying exit code if possible.
			if exitErr, ok := err.(*exec.ExitError); ok {
				if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					os.Exit(ws.ExitStatus())
				}
				os.Exit(1)
			}
			fail("Command failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
