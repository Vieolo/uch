package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vieolo/filange"
	"github.com/vieolo/uch/internal/fsutil"
	"gopkg.in/yaml.v3"
)

// CurrentSchemaVersion is the schema version this binary writes. Bump only on
// breaking format changes (renamed fields, restructured types, anything that
// requires migration logic). New commands, flags, and behaviors do NOT bump it.
const CurrentSchemaVersion = 1

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
	Version    int                `yaml:"version"`
	CreatedBy  string             `yaml:"created_by,omitempty"`
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

// createdByBreadcrumb is set by the cmd package at startup and embedded into
// every Save. Purely diagnostic — never read by Load logic.
var createdByBreadcrumb string

// SetCreatedBy records a "this file was written by ..." breadcrumb that will
// be stamped into every subsequent Save. The cmd package calls this at init
// with the CLI version.
func SetCreatedBy(s string) {
	createdByBreadcrumb = s
}

// Load reads ~/.uch/config.yaml, creating an empty one if it does not exist.
// Unknown fields are rejected (catches typos and "config from a newer uch").
// Configs from older schemas are migrated to CurrentSchemaVersion before return.
func Load() (Config, Paths, error) {
	paths, err := GetPaths()
	if err != nil {
		return Config{}, paths, err
	}

	if err := filange.CreateDirIfNotExists(paths.Dir, os.FileMode(0700)); err != nil {
		return Config{}, paths, fmt.Errorf("could not create %s: %w", paths.Dir, err)
	}

	if !filange.FileExists(paths.ConfigPath) {
		empty := Config{Version: CurrentSchemaVersion, Commands: map[string]Command{}}
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
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		// yaml.v3 wraps unknown-field errors with a helpful line number, but
		// the phrasing is technical. Surface a friendlier hint when we can.
		if strings.Contains(err.Error(), "field") && strings.Contains(err.Error(), "not found") {
			return Config{}, paths, fmt.Errorf("config.yaml has a field this version of uch does not recognize — either a typo or a newer schema: %w", err)
		}
		return Config{}, paths, fmt.Errorf("malformed config.yaml: %w", err)
	}

	if c.Version > CurrentSchemaVersion {
		return Config{}, paths, fmt.Errorf("config.yaml has schema version %d but this uch only supports up to %d — upgrade uch", c.Version, CurrentSchemaVersion)
	}
	if c.Version < CurrentSchemaVersion {
		c, err = migrate(c)
		if err != nil {
			return Config{}, paths, fmt.Errorf("migrate config from schema %d: %w", c.Version, err)
		}
	}

	if c.Commands == nil {
		c.Commands = map[string]Command{}
	}
	return c, paths, nil
}

// migrate applies forward migrations until c.Version == CurrentSchemaVersion.
// Each step handles exactly one N -> N+1 transition. Today there are no past
// schemas, so we just accept legacy (version: 0, i.e. missing) configs as v1.
func migrate(c Config) (Config, error) {
	for c.Version < CurrentSchemaVersion {
		switch c.Version {
		case 0:
			// Legacy unversioned config: structurally identical to v1, just
			// missing the version field. Stamp it and move on.
			c.Version = 1
		default:
			return c, fmt.Errorf("no migration path from schema %d", c.Version)
		}
	}
	return c, nil
}

// Save writes the config back to ~/.uch/config.yaml atomically with 0600
// permissions. It always stamps the current schema version and the diagnostic
// created_by breadcrumb so callers don't have to remember to.
func Save(c Config) error {
	paths, err := GetPaths()
	if err != nil {
		return err
	}
	if err := filange.CreateDirIfNotExists(paths.Dir, os.FileMode(0700)); err != nil {
		return err
	}

	c.Version = CurrentSchemaVersion
	if createdByBreadcrumb != "" {
		c.CreatedBy = createdByBreadcrumb
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}
	if err := fsutil.AtomicWriteFile(paths.ConfigPath, data, os.FileMode(0600)); err != nil {
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
