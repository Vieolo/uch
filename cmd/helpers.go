package cmd

import (
	"errors"
	"os"

	"filippo.io/age"
	"github.com/vieolo/filange"
	"github.com/vieolo/termange"
	"github.com/vieolo/uch/internal/config"
	"github.com/vieolo/uch/internal/crypto"
	"github.com/vieolo/uch/internal/prompt"
)

// unlockIdentity prompts for the uch password and returns the decrypted
// identity. Returns ErrNoIdentity if the identity file does not exist.
var errNoIdentity = errors.New("identity not initialized")

func unlockIdentity(paths config.Paths) (*age.X25519Identity, error) {
	if !filange.FileExists(paths.IdentityPath) {
		return nil, errNoIdentity
	}
	pw, err := prompt.Password("uch password:")
	if err != nil {
		return nil, err
	}
	id, err := crypto.LoadIdentity(paths.IdentityPath, pw)
	if err != nil {
		return nil, err
	}
	return id, nil
}

// requireIdentity is the same as unlockIdentity but exits with a friendly error
// if the identity has not been initialized.
func requireIdentity(paths config.Paths) *age.X25519Identity {
	id, err := unlockIdentity(paths)
	if err != nil {
		if errors.Is(err, errNoIdentity) {
			termange.PrintErrorln("Encryption is not initialized. Run 'uch init' first.")
		} else {
			termange.PrintErrorf("%v\n", err)
		}
		os.Exit(1)
	}
	return id
}

// fail prints an error and exits 1.
func fail(format string, args ...any) {
	termange.PrintErrorf(format+"\n", args...)
	os.Exit(1)
}

// must converts a non-nil error into a fatal exit.
func must(err error, msg string) {
	if err != nil {
		fail("%s: %v", msg, err)
	}
}
