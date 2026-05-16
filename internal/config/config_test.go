package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vieolo/uch/internal/config"
	"github.com/vieolo/uch/internal/crypto"
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
	config.SetCreatedBy("uch test-1.2.3")
	t.Cleanup(func() { config.SetCreatedBy("") })

	cfg := config.Config{Commands: map[string]config.Command{"x": {Cmd: "echo x"}}}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	paths, err := config.GetPaths()
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
	paths, err := config.GetPaths()
	if err != nil {
		t.Fatalf("GetPaths: %v", err)
	}
	if err := os.MkdirAll(paths.Dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Legacy config: no version field, structurally v1.
	legacy := `commands:
    hello:
        cmd: echo hello
`
	if err := os.WriteFile(paths.ConfigPath, []byte(legacy), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, _, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Version != config.CurrentSchemaVersion {
		t.Errorf("Version = %d, want %d", cfg.Version, config.CurrentSchemaVersion)
	}
	if cfg.Commands["hello"].Cmd != "echo hello" {
		t.Errorf("did not preserve command body after migration: %+v", cfg.Commands)
	}
}

func TestLoadRejectsFutureSchema(t *testing.T) {
	withTempHome(t)
	paths, _ := config.GetPaths()
	_ = os.MkdirAll(paths.Dir, 0700)
	future := `version: 99
commands: {}
`
	if err := os.WriteFile(paths.ConfigPath, []byte(future), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, _, err := config.Load()
	if err == nil {
		t.Fatal("expected error loading config from a newer schema")
	}
	if !strings.Contains(err.Error(), "upgrade uch") {
		t.Errorf("error message should hint at upgrading uch, got: %v", err)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	withTempHome(t)
	paths, _ := config.GetPaths()
	_ = os.MkdirAll(paths.Dir, 0700)
	bad := `version: 1
commands:
    hello:
        cmd: echo hi
        cmnd: typo
`
	if err := os.WriteFile(paths.ConfigPath, []byte(bad), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, _, err := config.Load()
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestAtomicWritePreservesPermsAndCleansTemp(t *testing.T) {
	withTempHome(t)
	cfg := config.Config{Commands: map[string]config.Command{"x": {Cmd: "echo x"}}}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	paths, _ := config.GetPaths()
	info, err := os.Stat(paths.ConfigPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("config perm = %o, want 0600", info.Mode().Perm())
	}
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

func TestValidateCatchesInconsistencies(t *testing.T) {
	cases := []struct {
		name    string
		cmd     config.Command
		wantErr string
	}{
		{"sensitive_without_ciphertext", config.Command{Sensitive: true}, "no cmd_encrypted body"},
		{"ciphertext_without_sensitive", config.Command{CmdEncrypted: "x"}, "without sensitive: true"},
		{"both_cmd_and_ciphertext", config.Command{Cmd: "echo", CmdEncrypted: "x", Sensitive: true}, "both cmd and cmd_encrypted"},
		{"empty_body", config.Command{}, "no body"},
		{"select_without_options", config.Command{
			Cmd:       "echo {{p}}",
			Variables: []config.Variable{{Name: "p", Type: config.VarSelect}},
		}, "no options"},
		{"variable_without_name", config.Command{
			Cmd:       "echo {{p}}",
			Variables: []config.Variable{{Type: config.VarString}},
		}, "no name"},
		{"duplicate_variable", config.Command{
			Cmd: "echo {{p}}-{{p}}",
			Variables: []config.Variable{
				{Name: "p", Type: config.VarString},
				{Name: "p", Type: config.VarString},
			},
		}, "duplicate variable"},
		{"unknown_variable_type", config.Command{
			Cmd:       "echo {{p}}",
			Variables: []config.Variable{{Name: "p", Type: "weird"}},
		}, "unknown type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := config.Config{Commands: map[string]config.Command{"x": tc.cmd}}
			err := c.Validate()
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q should contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestValidateAcceptsValidConfig(t *testing.T) {
	c := config.Config{
		Commands: map[string]config.Command{
			"plain": {Cmd: "echo hi"},
			"with_vars": {
				Cmd: "deploy --env {{env}} --confirm={{ok}}",
				Variables: []config.Variable{
					{Name: "env", Type: config.VarSelect, Options: []string{"staging", "prod"}},
					{Name: "ok", Type: config.VarConfirm},
				},
			},
			"with_cwd_env": {
				Cmd: "ls",
				Cwd: "~/projects",
				Env: map[string]string{"FOO": "bar"},
			},
		},
	}
	if err := c.Validate(); err != nil {
		t.Errorf("Validate on valid config: %v", err)
	}
}

// TestSensitiveCiphertextSurvivesYAMLRoundtrip is the load-bearing test that
// proves an age-armored ciphertext can be saved into config.yaml and read
// back without corruption. Any change to YAML serialization or armor format
// has to keep this passing.
func TestSensitiveCiphertextSurvivesYAMLRoundtrip(t *testing.T) {
	withTempHome(t)
	identityPath := filepath.Join(t.TempDir(), "id.age")
	const pw = "test-pw"
	if err := crypto.GenerateIdentity(identityPath, pw); err != nil {
		t.Fatalf("GenerateIdentity: %v", err)
	}
	id, err := crypto.LoadIdentity(identityPath, pw)
	if err != nil {
		t.Fatalf("LoadIdentity: %v", err)
	}

	// Use a 4KB body with embedded newlines, leading whitespace per line, and
	// trailing whitespace — the kinds of content YAML block scalars sometimes
	// re-flow.
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("    line with leading spaces and trailing tab \t \n")
	}
	plaintext := sb.String()

	cipher, err := crypto.EncryptString(plaintext, id)
	if err != nil {
		t.Fatalf("EncryptString: %v", err)
	}

	cfg := config.Config{
		Commands: map[string]config.Command{
			"secret": {Sensitive: true, CmdEncrypted: cipher},
		},
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, _, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	gotCipher := loaded.Commands["secret"].CmdEncrypted
	if gotCipher == "" {
		t.Fatal("ciphertext was empty after roundtrip")
	}

	recovered, err := crypto.DecryptString(gotCipher, id)
	if err != nil {
		t.Fatalf("DecryptString after roundtrip: %v", err)
	}
	if recovered != plaintext {
		t.Errorf("plaintext changed across YAML roundtrip\nwant len %d, got len %d", len(plaintext), len(recovered))
	}
}
