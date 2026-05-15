package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vieolo/filange"
	"gopkg.in/yaml.v3"
)

type VarType string

const (
	VarString  VarType = "string"
	VarSelect  VarType = "select"
	VarConfirm VarType = "confirm"
)

// Variable describes a placeholder in a command's `cmd` string.
// The user is prompted for a value at run time according to Type.
type Variable struct {
	Type    VarType  `yaml:"type"`
	Prompt  string   `yaml:"prompt,omitempty"`
	Options []string `yaml:"options,omitempty"`
	Default string   `yaml:"default,omitempty"`
}

// Command is one stored command. Either Cmd or CmdEncrypted is set; never both.
// CmdEncrypted is an age-armored ASCII envelope when Sensitive is true.
type Command struct {
	Description  string              `yaml:"description,omitempty"`
	Cmd          string              `yaml:"cmd,omitempty"`
	CmdEncrypted string              `yaml:"cmd_encrypted,omitempty"`
	Sensitive    bool                `yaml:"sensitive,omitempty"`
	Variables    map[string]Variable `yaml:"variables,omitempty"`
}

type Encryption struct {
	Enabled bool `yaml:"enabled"`
}

type Config struct {
	Encryption Encryption         `yaml:"encryption"`
	Commands   map[string]Command `yaml:"commands"`
}

type Paths struct {
	Dir          string
	ConfigPath   string
	IdentityPath string
}

func GetPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("could not resolve home directory: %w", err)
	}
	dir := filepath.Join(home, ".uch")
	return Paths{
		Dir:          dir,
		ConfigPath:   filepath.Join(dir, "config.yaml"),
		IdentityPath: filepath.Join(dir, "identity.age"),
	}, nil
}

// Load reads ~/.uch/config.yaml, creating an empty one if it does not exist.
func Load() (Config, Paths, error) {
	paths, err := GetPaths()
	if err != nil {
		return Config{}, paths, err
	}

	if err := filange.CreateDirIfNotExists(paths.Dir, os.FileMode(0700)); err != nil {
		return Config{}, paths, fmt.Errorf("could not create %s: %w", paths.Dir, err)
	}

	if !filange.FileExists(paths.ConfigPath) {
		empty := Config{Commands: map[string]Command{}}
		if err := Save(empty); err != nil {
			return Config{}, paths, err
		}
		return empty, paths, nil
	}

	data, err := os.ReadFile(paths.ConfigPath)
	if err != nil {
		return Config{}, paths, fmt.Errorf("could not read %s: %w", paths.ConfigPath, err)
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Config{}, paths, fmt.Errorf("malformed config.yaml: %w", err)
	}
	if c.Commands == nil {
		c.Commands = map[string]Command{}
	}
	return c, paths, nil
}

// Save writes the config back to ~/.uch/config.yaml with 0600 permissions.
func Save(c Config) error {
	paths, err := GetPaths()
	if err != nil {
		return err
	}
	if err := filange.CreateDirIfNotExists(paths.Dir, os.FileMode(0700)); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}
	if err := os.WriteFile(paths.ConfigPath, data, os.FileMode(0600)); err != nil {
		return fmt.Errorf("could not write %s: %w", paths.ConfigPath, err)
	}
	return nil
}

// HasSensitiveCommand reports whether any command has Sensitive=true.
func (c Config) HasSensitiveCommand() bool {
	for _, cmd := range c.Commands {
		if cmd.Sensitive {
			return true
		}
	}
	return false
}
