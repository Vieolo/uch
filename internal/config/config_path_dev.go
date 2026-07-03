//go:build dev

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve home directory: %w", err)
	}
	return filepath.Join(home, ".uch-sandbox"), nil
}
