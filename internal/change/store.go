package change

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/chiisen/speclet/internal/i18n"
	"github.com/chiisen/speclet/internal/task"
)

type Summary struct {
	Name  string `json:"name"`
	Done  int    `json:"done"`
	Total int    `json:"total"`
}

type Detail struct {
	Name         string
	ProposalPath string
	TasksPath    string
	DeltaPaths   []string
}

// Create creates a change directory without replacing an existing change.
func Create(root, name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	if err := requireDirectory(root); err != nil {
		return i18n.Errorf("speclet root: %w", err)
	}
	changesDir := filepath.Join(root, "changes")
	if err := requireDirectory(changesDir); err != nil {
		return i18n.Errorf("changes directory: %w", err)
	}
	archiveDir := filepath.Join(changesDir, "archive")
	if err := requireDirectory(archiveDir); err != nil {
		return i18n.Errorf("archive directory: %w", err)
	}
	changeDir := filepath.Join(changesDir, name)
	if _, err := os.Lstat(changeDir); err == nil {
		return i18n.Errorf("change %q already exists", name)
	} else if !os.IsNotExist(err) {
		return i18n.Errorf("inspect change %q: %w", name, err)
	}
	tempDir, err := os.MkdirTemp(archiveDir, ".speclet-create-*")
	if err != nil {
		return i18n.Errorf("prepare change %q: %w", name, err)
	}
	defer os.RemoveAll(tempDir)
	if err := os.Chmod(tempDir, 0o755); err != nil {
		return i18n.Errorf("set change directory mode: %w", err)
	}
	if err := os.Mkdir(filepath.Join(tempDir, "specs"), 0o755); err != nil {
		return i18n.Errorf("create change specs: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "proposal.md"), []byte("# Change: "+name+"\n\n## Why\n\n## What Changes\n"), 0o644); err != nil {
		return i18n.Errorf("create proposal.md: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "tasks.md"), nil, 0o644); err != nil {
		return i18n.Errorf("create tasks.md: %w", err)
	}
	if _, err := os.Lstat(changeDir); err == nil {
		return i18n.Errorf("change %q already exists", name)
	} else if !os.IsNotExist(err) {
		return i18n.Errorf("inspect change %q: %w", name, err)
	}
	if err := os.Rename(tempDir, changeDir); err != nil {
		return i18n.Errorf("publish change %q: %w", name, err)
	}
	return nil
}

// List returns active changes sorted by name and their task counts.
func List(root string) ([]Summary, error) {
	changesDir := filepath.Join(root, "changes")
	if err := requireDirectory(changesDir); err != nil {
		return nil, i18n.Errorf("changes directory: %w", err)
	}
	entries, err := os.ReadDir(changesDir)
	if err != nil {
		return nil, i18n.Errorf("read changes directory: %w", err)
	}
	summaries := make([]Summary, 0)
	for _, entry := range entries {
		if entry.Name() == "archive" || !entry.IsDir() {
			continue
		}
		changeDir := filepath.Join(changesDir, entry.Name())
		if err := requireDirectory(changeDir); err != nil {
			return nil, i18n.Errorf("change %q: %w", entry.Name(), err)
		}
		tasksPath := filepath.Join(changeDir, "tasks.md")
		if err := requireRegularFile(tasksPath); err != nil {
			return nil, i18n.Errorf("change %q tasks.md: %w", entry.Name(), err)
		}
		src, err := os.ReadFile(tasksPath)
		if err != nil {
			return nil, i18n.Errorf("read %s: %w", tasksPath, err)
		}
		done, total, err := task.Count(src)
		if err != nil {
			return nil, i18n.Errorf("change %q tasks.md: %w", entry.Name(), err)
		}
		summaries = append(summaries, Summary{Name: entry.Name(), Done: done, Total: total})
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].Name < summaries[j].Name })
	return summaries, nil
}

// Read returns the principal files for one active change.
func Read(root, name string) (Detail, error) {
	if err := validateName(name); err != nil {
		return Detail{}, err
	}
	if err := requireDirectory(root); err != nil {
		return Detail{}, i18n.Errorf("speclet root: %w", err)
	}
	if err := requireDirectory(filepath.Join(root, "changes")); err != nil {
		return Detail{}, i18n.Errorf("changes directory: %w", err)
	}
	changeDir := filepath.Join(root, "changes", name)
	if err := requireDirectory(changeDir); err != nil {
		return Detail{}, i18n.Errorf("change %q: %w", name, err)
	}
	detail := Detail{
		Name:         name,
		ProposalPath: filepath.Join(changeDir, "proposal.md"),
		TasksPath:    filepath.Join(changeDir, "tasks.md"),
	}
	for _, path := range []string{detail.ProposalPath, detail.TasksPath} {
		if err := requireRegularFile(path); err != nil {
			return Detail{}, i18n.Errorf("change %q: %w", name, err)
		}
	}
	specsDir := filepath.Join(changeDir, "specs")
	if err := requireDirectory(specsDir); err != nil {
		return Detail{}, i18n.Errorf("change %q specs: %w", name, err)
	}
	err := filepath.WalkDir(specsDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return i18n.Errorf("symbolic links are not supported: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() == "spec.md" {
			if err := requireRegularFile(path); err != nil {
				return err
			}
			detail.DeltaPaths = append(detail.DeltaPaths, path)
		}
		return nil
	})
	if err != nil {
		return Detail{}, i18n.Errorf("read change specs: %w", err)
	}
	sort.Strings(detail.DeltaPaths)
	return detail, nil
}

func validateName(name string) error {
	if strings.TrimSpace(name) == "" || name == "." || name == ".." || strings.Contains(name, "..") || filepath.IsAbs(name) || strings.ContainsAny(name, `/\\`) {
		return i18n.Errorf("invalid change name %q", name)
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return i18n.Errorf("invalid change name %q: control characters are not allowed", name)
		}
	}
	return nil
}

func requireDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return i18n.Errorf("not a regular directory")
	}
	return nil
}

func requireRegularFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return i18n.Errorf("not a regular file")
	}
	return nil
}
