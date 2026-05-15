// Package prompt provides the user-facing prompts uch needs at run time:
// resolving placeholder variables, and reading a master password without echo.
package prompt

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/vieolo/termange/tui"
	"github.com/vieolo/uch/internal/config"
	"golang.org/x/term"
)

// ResolveVariables prompts the user for each variable defined on a command and
// returns name -> value. The values are ready to be substituted into the cmd
// string. Confirm variables resolve to "yes" or "no".
func ResolveVariables(vars map[string]config.Variable) (map[string]string, error) {
	out := make(map[string]string, len(vars))

	names := make([]string, 0, len(vars))
	for name := range vars {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		v := vars[name]
		label := v.Prompt
		if label == "" {
			label = name
		}

		switch v.Type {
		case config.VarString:
			if v.Default != "" {
				label = fmt.Sprintf("%s [%s]:", label, v.Default)
			} else {
				label = label + ":"
			}
			val := tui.TextInput(tui.TextInputOptions{Prompt: label})
			if strings.TrimSpace(val) == "" {
				val = v.Default
			}
			out[name] = val
		case config.VarSelect:
			if len(v.Options) == 0 {
				return nil, fmt.Errorf("variable %q is type select but has no options", name)
			}
			items := make([]tui.SelectItem, 0, len(v.Options))
			for _, opt := range v.Options {
				items = append(items, tui.SelectItem(opt))
			}
			choice, err := tui.Select(tui.SelectOptions{
				Title: label,
				Items: items,
			})
			if err != nil {
				return nil, fmt.Errorf("select for %q failed: %w", name, err)
			}
			if choice == "" {
				return nil, fmt.Errorf("no option selected for %q", name)
			}
			out[name] = choice
		case config.VarConfirm:
			ans := tui.Confirm(tui.ConfirmOptions{Prompt: label})
			if ans {
				out[name] = "yes"
			} else {
				out[name] = "no"
			}
		default:
			return nil, fmt.Errorf("variable %q has unknown type %q", name, v.Type)
		}
	}
	return out, nil
}

// Substitute replaces every {{name}} placeholder in cmd with its value, and
// returns an error if any placeholder remains unresolved.
func Substitute(cmd string, values map[string]string) (string, error) {
	for name, val := range values {
		cmd = strings.ReplaceAll(cmd, "{{"+name+"}}", val)
	}
	if i := strings.Index(cmd, "{{"); i >= 0 {
		end := strings.Index(cmd[i:], "}}")
		if end > 0 {
			return "", fmt.Errorf("unresolved placeholder %s", cmd[i:i+end+2])
		}
	}
	return cmd, nil
}

// Password reads a single password from the terminal without echoing.
func Password(prompt string) (string, error) {
	fmt.Print(prompt)
	if !strings.HasSuffix(prompt, " ") {
		fmt.Print(" ")
	}
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("password prompt requires an interactive terminal")
	}
	pw, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(pw), nil
}

// PasswordWithConfirm prompts twice and ensures the entries match. Used at init
// and when changing the master password.
func PasswordWithConfirm(prompt string) (string, error) {
	pw1, err := Password(prompt)
	if err != nil {
		return "", err
	}
	if pw1 == "" {
		return "", errors.New("password cannot be empty")
	}
	pw2, err := Password("Confirm password:")
	if err != nil {
		return "", err
	}
	if pw1 != pw2 {
		return "", errors.New("passwords do not match")
	}
	return pw1, nil
}
