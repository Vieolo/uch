package crypto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIdentityRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "identity.age")
	const pw = "correct horse battery staple"

	if err := GenerateIdentity(path, pw); err != nil {
		t.Fatalf("GenerateIdentity: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("identity file perms = %o, want 0600", info.Mode().Perm())
	}

	id, err := LoadIdentity(path, pw)
	if err != nil {
		t.Fatalf("LoadIdentity correct pw: %v", err)
	}

	cipher, err := EncryptString("ssh root@prod", id)
	if err != nil {
		t.Fatalf("EncryptString: %v", err)
	}
	plain, err := DecryptString(cipher, id)
	if err != nil {
		t.Fatalf("DecryptString: %v", err)
	}
	if plain != "ssh root@prod" {
		t.Errorf("roundtrip = %q, want %q", plain, "ssh root@prod")
	}
}

func TestWrongPassword(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "identity.age")
	if err := GenerateIdentity(path, "right"); err != nil {
		t.Fatalf("GenerateIdentity: %v", err)
	}
	if _, err := LoadIdentity(path, "wrong"); err != ErrWrongPassword {
		t.Errorf("got err = %v, want ErrWrongPassword", err)
	}
}

func TestChangePassword(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "identity.age")
	if err := GenerateIdentity(path, "old"); err != nil {
		t.Fatalf("GenerateIdentity: %v", err)
	}
	idOld, err := LoadIdentity(path, "old")
	if err != nil {
		t.Fatalf("LoadIdentity old: %v", err)
	}
	cipher, err := EncryptString("secret", idOld)
	if err != nil {
		t.Fatalf("EncryptString: %v", err)
	}

	if err := ChangePassword(path, "old", "new"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if _, err := LoadIdentity(path, "old"); err != ErrWrongPassword {
		t.Errorf("old pw still works after change: %v", err)
	}
	idNew, err := LoadIdentity(path, "new")
	if err != nil {
		t.Fatalf("LoadIdentity new: %v", err)
	}
	// Existing ciphertext should still decrypt with the same underlying identity.
	plain, err := DecryptString(cipher, idNew)
	if err != nil {
		t.Fatalf("DecryptString after passwd: %v", err)
	}
	if plain != "secret" {
		t.Errorf("post-passwd roundtrip = %q, want %q", plain, "secret")
	}
}
