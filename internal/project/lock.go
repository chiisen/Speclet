package project

import (
	"os"
	"path/filepath"

	"github.com/chiisen/speclet/internal/i18n"
)

// AcquireWriteLock serializes mutating workflows for one local Speclet project.
func AcquireWriteLock(root string) (func() error, error) {
	if err := validateLayout(root); err != nil {
		return nil, i18n.Errorf("invalid speclet layout: %w", err)
	}
	lockPath := filepath.Join(root, ".operation.lock")
	if err := os.Mkdir(lockPath, 0o700); err != nil {
		if os.IsExist(err) {
			return nil, i18n.Errorf("another Speclet write is in progress, or a stale lock exists at %s", lockPath)
		}
		return nil, i18n.Errorf("acquire Speclet write lock: %w", err)
	}
	released := false
	return func() error {
		if released {
			return nil
		}
		released = true
		if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
			return i18n.Errorf("release Speclet write lock: %w", err)
		}
		return nil
	}, nil
}
