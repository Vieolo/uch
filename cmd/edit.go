package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"filippo.io/age"
	"github.com/spf13/cobra"
	"github.com/vieolo/termange"
	"github.com/vieolo/uch/internal/config"
	"github.com/vieolo/uch/internal/crypto"
	"gopkg.in/yaml.v3"
)

var editAdmin bool

const sensitivePlaceholder = "<encrypted — use 'uch edit --admin' to reveal>"

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit the stored commands in your $EDITOR",
	Long: `Opens config.yaml in $EDITOR (default: vim).
Without --admin, sensitive commands are shown as a placeholder and you cannot
change their body. With --admin, you are prompted for the master password, the
sensitive commands are shown in plaintext, and any changes are re-encrypted on
save.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, paths, err := config.Load()
		must(err, "Could not load config")

		var id *age.X25519Identity
		if editAdmin {
			id = requireIdentity(paths)
		}

		view := buildEditView(cfg, id)

		data, err := yaml.Marshal(view)
		must(err, "Could not marshal view")

		tmpPath := writeTempForEdit(data)
		defer os.Remove(tmpPath)

		runEditor(tmpPath)

		edited, err := os.ReadFile(tmpPath)
		must(err, "Could not read edited file")

		var next config.Config
		if err := yaml.Unmarshal(edited, &next); err != nil {
			fail("Edited file is not valid YAML: %v", err)
		}
		if next.Commands == nil {
			next.Commands = map[string]config.Command{}
		}

		if err := reconcileEdit(&next, cfg, editAdmin, id); err != nil {
			fail("%v", err)
		}

		must(config.Save(next), "Could not save config")
		termange.PrintSuccessln("config.yaml saved.")
	},
}

// buildEditView returns a copy of cfg with sensitive bodies replaced for the
// editor view. In admin mode they are decrypted; otherwise they are masked.
func buildEditView(cfg config.Config, id *age.X25519Identity) config.Config {
	out := config.Config{Commands: make(map[string]config.Command, len(cfg.Commands))}
	for name, entry := range cfg.Commands {
		if entry.Sensitive {
			if id != nil {
				plain, err := crypto.DecryptString(entry.CmdEncrypted, id)
				must(err, fmt.Sprintf("Could not decrypt %q", name))
				entry.Cmd = plain
			} else {
				entry.Cmd = sensitivePlaceholder
			}
			entry.CmdEncrypted = ""
		}
		out.Commands[name] = entry
	}
	return out
}

// reconcileEdit walks the edited config and merges sensitive entries with the
// original. In admin mode it re-encrypts bodies; otherwise it restores the
// original ciphertext and rejects edits that would require admin.
func reconcileEdit(next *config.Config, orig config.Config, admin bool, id *age.X25519Identity) error {
	for name, entry := range next.Commands {
		prev, hadPrev := orig.Commands[name]

		if admin {
			if entry.Sensitive {
				if entry.Cmd == "" {
					return fmt.Errorf("sensitive command %q has an empty body", name)
				}
				cipher, err := crypto.EncryptString(entry.Cmd, id)
				if err != nil {
					return fmt.Errorf("encrypt %q: %w", name, err)
				}
				entry.Cmd = ""
				entry.CmdEncrypted = cipher
			} else if hadPrev && prev.Sensitive {
				// Demoted sensitive -> normal. The body is in Cmd as the user typed it.
				entry.CmdEncrypted = ""
			}
			next.Commands[name] = entry
			continue
		}

		// Non-admin: protect sensitive entries from body changes.
		if entry.Sensitive {
			if !hadPrev || !prev.Sensitive {
				return fmt.Errorf("cannot mark %q sensitive without --admin (use 'uch add --sensitive' or 'uch edit --admin')", name)
			}
			if entry.Cmd != "" && entry.Cmd != sensitivePlaceholder {
				return fmt.Errorf("cannot change the body of sensitive command %q without --admin", name)
			}
			entry.Cmd = ""
			entry.CmdEncrypted = prev.CmdEncrypted
			next.Commands[name] = entry
			continue
		}
		if hadPrev && prev.Sensitive {
			return fmt.Errorf("cannot unmark sensitive command %q without --admin", name)
		}
	}
	return nil
}

func writeTempForEdit(data []byte) string {
	tmp, err := os.CreateTemp("", "uch-edit-*.yaml")
	must(err, "Could not create temp file")
	tmpPath := tmp.Name()
	if err := os.Chmod(tmpPath, 0600); err != nil {
		os.Remove(tmpPath)
		fail("Could not chmod temp file: %v", err)
	}
	if _, err := tmp.Write(data); err != nil {
		os.Remove(tmpPath)
		fail("Could not write temp file: %v", err)
	}
	tmp.Close()
	return tmpPath
}

func runEditor(path string) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	termange.PrintInfof("Opening config.yaml in %s...\n", editor)
	c := exec.Command(editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		fail("Editor exited with error: %v", err)
	}
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.Flags().BoolVar(&editAdmin, "admin", false, "Decrypt sensitive commands for editing; requires the master password")
}
