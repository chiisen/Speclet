package project

import (
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/chiisen/speclet/internal/i18n"
)

const configContents = "version: 1\n"

// Init creates a Speclet data directory inside target, or validates an existing one.
func Init(target string) error {
	if target == "" {
		return i18n.Errorf("project path is empty")
	}
	projectPath, err := filepath.Abs(target)
	if err != nil {
		return i18n.Errorf("resolve project path: %w", err)
	}
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		return i18n.Errorf("create project directory: %w", err)
	}
	root := filepath.Join(projectPath, "speclet")
	if _, err := os.Lstat(root); err == nil {
		if err := validateLayout(root); err != nil {
			return i18n.Errorf("invalid existing speclet layout: %w", err)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return i18n.Errorf("inspect speclet directory: %w", err)
	}

	tmp, err := os.MkdirTemp(projectPath, ".speclet-init-")
	if err != nil {
		return i18n.Errorf("prepare speclet layout: %w", err)
	}
	defer os.RemoveAll(tmp)

	for _, dir := range []string{
		"specs",
		filepath.Join("changes", "archive"),
	} {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0o755); err != nil {
			return i18n.Errorf("create %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(tmp, "config.yaml"), []byte(configContents), 0o644); err != nil {
		return i18n.Errorf("write config.yaml: %w", err)
	}
	if err := os.Rename(tmp, root); err != nil {
		return i18n.Errorf("install speclet layout: %w", err)
	}
	return nil
}

// FindRoot searches start and its parents for the nearest valid speclet directory.
func FindRoot(start string) (string, error) {
	if start == "" {
		return "", i18n.Errorf("working directory is empty")
	}
	current, err := filepath.Abs(start)
	if err != nil {
		return "", i18n.Errorf("resolve working directory: %w", err)
	}
	for {
		candidate := filepath.Join(current, "speclet")
		if _, err := os.Lstat(candidate); err == nil {
			if err := validateLayout(candidate); err != nil {
				return "", i18n.Errorf("invalid speclet layout at %s: %w", candidate, err)
			}
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", i18n.Errorf("inspect %s: %w", candidate, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", i18n.Errorf("speclet project not found from %s", start)
}

func validateLayout(root string) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return i18n.Errorf("speclet is not a directory")
	}

	configPath := filepath.Join(root, "config.yaml")
	configInfo, err := os.Lstat(configPath)
	if err != nil {
		return i18n.Errorf("config.yaml: %w", err)
	}
	if configInfo.Mode()&os.ModeSymlink != 0 || !configInfo.Mode().IsRegular() {
		return i18n.Errorf("config.yaml is not a regular file")
	}
	contents, err := os.ReadFile(configPath)
	if err != nil {
		return i18n.Errorf("read config.yaml: %w", err)
	}
	if !utf8.Valid(contents) {
		return i18n.Errorf("config.yaml is not valid UTF-8")
	}
	versionCount := 0
	for index, line := range strings.Split(string(contents), "\n") {
		if index == 0 {
			line = strings.TrimPrefix(line, "\uFEFF")
		}
		if strings.HasPrefix(line, "version:") {
			versionCount++
			versionLine := strings.TrimSpace(strings.TrimSuffix(line, "\r"))
			if comment := strings.IndexByte(versionLine, '#'); comment >= 0 {
				versionLine = strings.TrimSpace(versionLine[:comment])
			}
			if versionLine != "version: 1" {
				return i18n.Errorf("config.yaml must declare version: 1")
			}
		}
	}
	if versionCount != 1 {
		return i18n.Errorf("config.yaml must declare version: 1")
	}
	for _, dir := range []string{"specs", "changes", filepath.Join("changes", "archive")} {
		path := filepath.Join(root, dir)
		info, err := os.Lstat(path)
		if err != nil {
			return i18n.Errorf("%s: %w", dir, err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return i18n.Errorf("%s is not a directory", dir)
		}
	}
	return nil
}
