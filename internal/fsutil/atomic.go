// Package fsutil holds small filesystem helpers shared across uch.
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWriteFile writes data to path with perm using the standard
// write-temp-then-rename dance. On POSIX this guarantees the destination
// either points at the old file or the new one — never a partial write —
// even if the process or machine crashes mid-write.
//
// The temp file is created in the same directory so the rename is on the
// same filesystem (rename across mounts is not atomic).
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".uch-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	// Best-effort cleanup of the temp if anything below fails.
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("fsync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp into place: %w", err)
	}
	return nil
}
