package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// AcquireWriteLock serializes mutating workflows for one local Speclet project.
func AcquireWriteLock(root string) (func() error, error) {
	if err := validateLayout(root); err != nil {
		return nil, fmt.Errorf("invalid speclet layout: %w", err)
	}
	lockPath := filepath.Join(root, ".operation.lock")
	if err := os.Mkdir(lockPath, 0o700); err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("another Speclet write is in progress, or a stale lock exists at %s", lockPath)
		}
		return nil, fmt.Errorf("acquire Speclet write lock: %w", err)
	}
	released := false
	return func() error {
		if released {
			return nil
		}
		released = true
		if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("release Speclet write lock: %w", err)
		}
		return nil
	}, nil
}
