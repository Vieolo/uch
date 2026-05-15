package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

		// Resolve variables and substitute placeholders.
		values, err := prompt.ResolveVariables(entry.Variables)
		must(err, "Could not resolve variables")

		final, err := prompt.Substitute(body, values)
		must(err, "Could not substitute variables")

		// Execute via $SHELL -c so users get pipes, redirects, globbing, etc.
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

		if entry.Cwd != "" {
			cwd, err := expandHome(entry.Cwd)
			must(err, "Could not resolve cwd")
			c.Dir = cwd
		}
		if len(entry.Env) > 0 {
			c.Env = mergeEnv(os.Environ(), entry.Env)
		}

		if err := c.Run(); err != nil {
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

// expandHome resolves a leading "~/" to the user's home directory. Other
// patterns (bare "~", "~user") are intentionally not supported — let the
// shell handle anything more elaborate.
func expandHome(p string) (string, error) {
	if !strings.HasPrefix(p, "~/") {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expand ~/: %w", err)
	}
	return filepath.Join(home, p[2:]), nil
}

// mergeEnv overlays the per-command env onto the inherited environment.
// Per-command entries win on collision.
func mergeEnv(parent []string, overlay map[string]string) []string {
	merged := make([]string, 0, len(parent)+len(overlay))
	for _, kv := range parent {
		eq := strings.IndexByte(kv, '=')
		if eq < 0 {
			merged = append(merged, kv)
			continue
		}
		if _, ok := overlay[kv[:eq]]; ok {
			continue
		}
		merged = append(merged, kv)
	}
	for k, v := range overlay {
		merged = append(merged, k+"="+v)
	}
	return merged
}

func init() {
	rootCmd.AddCommand(runCmd)
}
