package archive

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chiisen/speclet/internal/change"
	"github.com/chiisen/speclet/internal/project"
	"github.com/chiisen/speclet/internal/specdoc"
	"github.com/chiisen/speclet/internal/task"
)

type specUpdate struct {
	target   string
	temp     string
	contents []byte
	original []byte
	mode     fs.FileMode
	existed  bool
	applied  bool
}

var replaceStagedFile = os.Rename
var moveChangeDirectory = os.Rename

// Run validates, applies, and archives one completed change.
func Run(specletRoot, name string) (result error) {
	if err := requireDirectory(specletRoot); err != nil {
		return fmt.Errorf("speclet root: %w", err)
	}
	release, err := project.AcquireWriteLock(specletRoot)
	if err != nil {
		return err
	}
	defer func() {
		if err := release(); err != nil && result == nil {
			result = err
		}
	}()
	detail, err := change.Read(specletRoot, name)
	if err != nil {
		return err
	}
	tasks, err := os.ReadFile(detail.TasksPath)
	if err != nil {
		return fmt.Errorf("read tasks.md: %w", err)
	}
	done, total, err := task.Count(tasks)
	if err != nil {
		return fmt.Errorf("validate tasks.md: %w", err)
	}
	if done != total {
		return fmt.Errorf("change %q has unfinished tasks (%d/%d done)", name, done, total)
	}

	changesDir := filepath.Join(specletRoot, "changes")
	archiveDir := filepath.Join(changesDir, "archive")
	if err := requireDirectory(changesDir); err != nil {
		return fmt.Errorf("changes directory: %w", err)
	}
	if err := requireDirectory(archiveDir); err != nil {
		return fmt.Errorf("archive directory: %w", err)
	}
	changeDir := filepath.Join(changesDir, name)
	archivePath := filepath.Join(archiveDir, time.Now().Format("2006-01-02")+"-"+name)
	if _, err := os.Lstat(archivePath); err == nil {
		return fmt.Errorf("archive destination already exists: %s", archivePath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect archive destination: %w", err)
	}

	specsDir := filepath.Join(specletRoot, "specs")
	if err := requireDirectory(specsDir); err != nil {
		return fmt.Errorf("formal specs directory: %w", err)
	}
	updates := make([]specUpdate, 0, len(detail.DeltaPaths))
	seenTargets := make(map[string]bool)
	for _, deltaPath := range detail.DeltaPaths {
		rel, err := filepath.Rel(filepath.Join(changeDir, "specs"), deltaPath)
		if err != nil || filepath.IsAbs(rel) || filepath.Clean(rel) != rel || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("delta path escapes change directory: %s", deltaPath)
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) != 2 || parts[1] != "spec.md" || parts[0] == "" || parts[0] == "." || parts[0] == ".." {
			return fmt.Errorf("delta path must be specs/<capability>/spec.md: %s", rel)
		}
		capabilityDir := filepath.Join(specsDir, parts[0])
		if _, err := os.Lstat(capabilityDir); err == nil {
			if err := requireDirectory(capabilityDir); err != nil {
				return fmt.Errorf("formal spec capability %q: %w", parts[0], err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect formal spec capability %q: %w", parts[0], err)
		}
		target := filepath.Join(capabilityDir, "spec.md")
		if seenTargets[target] {
			return fmt.Errorf("multiple deltas target %s", target)
		}
		seenTargets[target] = true
		delta, err := os.ReadFile(deltaPath)
		if err != nil {
			return fmt.Errorf("read delta %s: %w", deltaPath, err)
		}
		operations, err := specdoc.ParseDelta(delta)
		if err != nil {
			return fmt.Errorf("validate delta %s: %w", deltaPath, err)
		}
		var original []byte
		mode := fs.FileMode(0o644)
		existed := false
		if info, err := os.Lstat(target); err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				return fmt.Errorf("formal spec target is not a regular file: %s", target)
			}
			original, err = os.ReadFile(target)
			if err != nil {
				return fmt.Errorf("read formal spec %s: %w", target, err)
			}
			mode, existed = info.Mode().Perm(), true
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect formal spec %s: %w", target, err)
		}
		updated, err := specdoc.ApplyDelta(original, delta)
		if err != nil {
			return fmt.Errorf("apply delta %s: %w", deltaPath, err)
		}
		if len(operations) != 0 {
			updates = append(updates, specUpdate{
				target:   target,
				contents: append([]byte(nil), updated...),
				original: append([]byte(nil), original...),
				mode:     mode,
				existed:  existed,
			})
		}
	}

	return commit(changeDir, archivePath, updates)
}

func commit(changeDir, archivePath string, updates []specUpdate) error {
	createdDirs := make([]string, 0)
	cleanupDirs := func() {
		for i := len(createdDirs) - 1; i >= 0; i-- {
			_ = os.Remove(createdDirs[i])
		}
	}
	for i := range updates {
		parent := filepath.Dir(updates[i].target)
		if _, err := os.Lstat(parent); os.IsNotExist(err) {
			if err := os.Mkdir(parent, 0o755); err != nil {
				cleanupDirs()
				return fmt.Errorf("create formal spec directory %s: %w", parent, err)
			}
			createdDirs = append(createdDirs, parent)
		} else if err != nil {
			cleanupDirs()
			return fmt.Errorf("inspect formal spec directory %s: %w", parent, err)
		}
	}

	for i := range updates {
		file, err := os.CreateTemp(filepath.Dir(updates[i].target), ".speclet-archive-*")
		if err != nil {
			cleanupUpdates(updates)
			cleanupDirs()
			return fmt.Errorf("stage formal spec %s: %w", updates[i].target, err)
		}
		updates[i].temp = file.Name()
		if _, err := file.Write(updates[i].contents); err != nil {
			_ = file.Close()
			cleanupUpdates(updates)
			cleanupDirs()
			return fmt.Errorf("write staged formal spec %s: %w", updates[i].target, err)
		}
		if err := file.Chmod(updates[i].mode); err != nil {
			_ = file.Close()
			cleanupUpdates(updates)
			cleanupDirs()
			return fmt.Errorf("set staged formal spec mode: %w", err)
		}
		if err := file.Close(); err != nil {
			cleanupUpdates(updates)
			cleanupDirs()
			return fmt.Errorf("close staged formal spec: %w", err)
		}
	}

	for i := range updates {
		if err := replaceStagedFile(updates[i].temp, updates[i].target); err != nil {
			rollbackErr := rollback(updates[:i])
			cleanupUpdates(updates)
			cleanupDirs()
			return combineError(fmt.Errorf("replace formal spec %s: %w", updates[i].target, err), rollbackErr)
		}
		updates[i].applied = true
		updates[i].temp = ""
	}
	if _, err := os.Lstat(archivePath); err == nil {
		rollbackErr := rollback(updates)
		cleanupDirs()
		return combineError(fmt.Errorf("archive destination already exists: %s", archivePath), rollbackErr)
	} else if !os.IsNotExist(err) {
		rollbackErr := rollback(updates)
		cleanupDirs()
		return combineError(fmt.Errorf("inspect archive destination: %w", err), rollbackErr)
	}
	if err := moveChangeDirectory(changeDir, archivePath); err != nil {
		rollbackErr := rollback(updates)
		cleanupDirs()
		return combineError(fmt.Errorf("archive change: %w", err), rollbackErr)
	}
	cleanupUpdates(updates)
	return nil
}

func rollback(updates []specUpdate) error {
	var failures []string
	for i := len(updates) - 1; i >= 0; i-- {
		if !updates[i].applied {
			continue
		}
		if updates[i].existed {
			if err := atomicWrite(updates[i].target, updates[i].original, updates[i].mode); err != nil {
				failures = append(failures, fmt.Sprintf("restore %s: %v", updates[i].target, err))
			}
		} else if err := os.Remove(updates[i].target); err != nil && !os.IsNotExist(err) {
			failures = append(failures, fmt.Sprintf("remove %s: %v", updates[i].target, err))
		}
	}
	if len(failures) == 0 {
		return nil
	}
	return fmt.Errorf("rollback failed: %s", strings.Join(failures, "; "))
}

func atomicWrite(path string, contents []byte, mode fs.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".speclet-rollback-*")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)
	if _, err := file.Write(contents); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

func cleanupUpdates(updates []specUpdate) {
	for _, update := range updates {
		if update.temp != "" {
			_ = os.Remove(update.temp)
		}
	}
}

func combineError(primary, rollbackErr error) error {
	if rollbackErr == nil {
		return primary
	}
	return fmt.Errorf("%v; %w", primary, rollbackErr)
}

func requireDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("not a regular directory")
	}
	return nil
}
