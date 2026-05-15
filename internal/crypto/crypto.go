// Package crypto wraps filippo.io/age for uch's encryption needs.
//
// Model:
//   - One long-lived X25519 identity stores in ~/.uch/identity.age, itself
//     encrypted with the user's master password via age's scrypt recipient.
//   - Sensitive commands are encrypted to the X25519 recipient and stored
//     ASCII-armored inside config.yaml.
//   - The password never touches disk; it only unlocks the identity in memory.
package crypto

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// ErrWrongPassword is returned when the master password fails to unlock the identity.
var ErrWrongPassword = errors.New("wrong password")

// GenerateIdentity creates a new X25519 identity, encrypts it with the given
// password using age's scrypt recipient, and writes it ASCII-armored to path.
func GenerateIdentity(path, password string) error {
	id, err := age.GenerateX25519Identity()
	if err != nil {
		return fmt.Errorf("generate identity: %w", err)
	}
	return writeEncryptedIdentity(path, id, password)
}

// LoadIdentity reads an age-armored, scrypt-encrypted identity file and
// decrypts it with the given password.
func LoadIdentity(path, password string) (*age.X25519Identity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read identity: %w", err)
	}

	scryptID, err := age.NewScryptIdentity(password)
	if err != nil {
		return nil, fmt.Errorf("scrypt identity: %w", err)
	}

	armored := armor.NewReader(bytes.NewReader(data))
	r, err := age.Decrypt(armored, scryptID)
	if err != nil {
		// age returns a wrapped error containing "incorrect passphrase" when
		// the scrypt identity fails to unwrap the file key.
		if strings.Contains(err.Error(), "incorrect passphrase") {
			return nil, ErrWrongPassword
		}
		return nil, fmt.Errorf("decrypt identity: %w", err)
	}

	keyText, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read decrypted identity: %w", err)
	}

	id, err := age.ParseX25519Identity(strings.TrimSpace(string(keyText)))
	if err != nil {
		return nil, fmt.Errorf("parse identity: %w", err)
	}
	return id, nil
}

// ChangePassword decrypts the identity at path with oldPassword and re-encrypts
// it with newPassword. The same identity (and therefore all existing ciphertexts)
// remains valid.
func ChangePassword(path, oldPassword, newPassword string) error {
	id, err := LoadIdentity(path, oldPassword)
	if err != nil {
		return err
	}
	return writeEncryptedIdentity(path, id, newPassword)
}

func writeEncryptedIdentity(path string, id *age.X25519Identity, password string) error {
	scryptRecipient, err := age.NewScryptRecipient(password)
	if err != nil {
		return fmt.Errorf("scrypt recipient: %w", err)
	}

	var buf bytes.Buffer
	armored := armor.NewWriter(&buf)
	w, err := age.Encrypt(armored, scryptRecipient)
	if err != nil {
		return fmt.Errorf("encrypt identity: %w", err)
	}
	if _, err := io.WriteString(w, id.String()); err != nil {
		return fmt.Errorf("write identity: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close encrypt writer: %w", err)
	}
	if err := armored.Close(); err != nil {
		return fmt.Errorf("close armor writer: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), os.FileMode(0600)); err != nil {
		return fmt.Errorf("write identity file: %w", err)
	}
	return nil
}

// EncryptString encrypts plaintext to the public side of id, returning an
// ASCII-armored age envelope safe to embed in YAML.
func EncryptString(plaintext string, id *age.X25519Identity) (string, error) {
	var buf bytes.Buffer
	armored := armor.NewWriter(&buf)
	w, err := age.Encrypt(armored, id.Recipient())
	if err != nil {
		return "", fmt.Errorf("encrypt: %w", err)
	}
	if _, err := io.WriteString(w, plaintext); err != nil {
		return "", fmt.Errorf("write plaintext: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("close encrypt writer: %w", err)
	}
	if err := armored.Close(); err != nil {
		return "", fmt.Errorf("close armor writer: %w", err)
	}
	return buf.String(), nil
}

// DecryptString decrypts an ASCII-armored age envelope using id.
func DecryptString(armored string, id *age.X25519Identity) (string, error) {
	r := armor.NewReader(strings.NewReader(armored))
	dec, err := age.Decrypt(r, id)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	plaintext, err := io.ReadAll(dec)
	if err != nil {
		return "", fmt.Errorf("read plaintext: %w", err)
	}
	return string(plaintext), nil
}
