package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withTempHome points $HOME at a temp dir for the duration of the test so
// GetPaths / Load / Save operate on a sandbox.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestSaveStampsVersionAndCreatedBy(t *testing.T) {
	withTempHome(t)
	SetCreatedBy("uch test-1.2.3")
	t.Cleanup(func() { SetCreatedBy("") })

	if err := Save(Config{Commands: map[string]Command{"x": {Cmd: "echo x"}}}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	paths, err := GetPaths()
	if err != nil {
		t.Fatalf("GetPaths: %v", err)
	}
	raw, err := os.ReadFile(paths.ConfigPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	got := string(raw)
	if !strings.Contains(got, "version: 1") {
		t.Errorf("missing version stamp in:\n%s", got)
	}
	if !strings.Contains(got, "created_by: uch test-1.2.3") {
		t.Errorf("missing created_by breadcrumb in:\n%s", got)
	}
}

func TestLoadMigratesLegacyUnversionedConfig(t *testing.T) {
	withTempHome(t)
	paths, err := GetPaths()
	if err != nil {
		t.Fatalf("GetPaths: %v", err)
	}
	if err := os.MkdirAll(paths.Dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Legacy config has no version field.
	legacy := `encryption:
    enabled: false
commands:
    hello:
        cmd: echo hello
`
	if err := os.WriteFile(paths.ConfigPath, []byte(legacy), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, _, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Version != CurrentSchemaVersion {
		t.Errorf("Version = %d, want %d", cfg.Version, CurrentSchemaVersion)
	}
	if cfg.Commands["hello"].Cmd != "echo hello" {
		t.Errorf("did not preserve command body after migration: %+v", cfg.Commands)
	}
}

func TestLoadRejectsFutureSchema(t *testing.T) {
	withTempHome(t)
	paths, _ := GetPaths()
	_ = os.MkdirAll(paths.Dir, 0700)
	future := `version: 99
encryption:
    enabled: false
commands: {}
`
	if err := os.WriteFile(paths.ConfigPath, []byte(future), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, _, err := Load()
	if err == nil {
		t.Fatal("expected error loading config from a newer schema")
	}
	if !strings.Contains(err.Error(), "upgrade uch") {
		t.Errorf("error message should hint at upgrading uch, got: %v", err)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	withTempHome(t)
	paths, _ := GetPaths()
	_ = os.MkdirAll(paths.Dir, 0700)
	bad := `version: 1
encryption:
    enabled: false
commands:
    hello:
        cmd: echo hi
        cmnd: typo
`
	if err := os.WriteFile(paths.ConfigPath, []byte(bad), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, _, err := Load()
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestAtomicWritePreservesPermsAndCleansTemp(t *testing.T) {
	withTempHome(t)
	if err := Save(Config{Commands: map[string]Command{"x": {Cmd: "echo x"}}}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	paths, _ := GetPaths()
	info, err := os.Stat(paths.ConfigPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("config perm = %o, want 0600", info.Mode().Perm())
	}
	// No temp leftovers.
	entries, err := os.ReadDir(paths.Dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".uch-tmp-") {
			t.Errorf("temp file leaked: %s", filepath.Join(paths.Dir, e.Name()))
		}
	}
}
