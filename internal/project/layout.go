package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const configContents = "version: 1\n"

// Init creates a Speclet data directory inside target, or validates an existing one.
func Init(target string) error {
	if target == "" {
		return fmt.Errorf("project path is empty")
	}
	projectPath, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}
	root := filepath.Join(projectPath, "speclet")
	if _, err := os.Lstat(root); err == nil {
		if err := validateLayout(root); err != nil {
			return fmt.Errorf("invalid existing speclet layout: %w", err)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect speclet directory: %w", err)
	}

	tmp, err := os.MkdirTemp(projectPath, ".speclet-init-")
	if err != nil {
		return fmt.Errorf("prepare speclet layout: %w", err)
	}
	defer os.RemoveAll(tmp)

	for _, dir := range []string{
		"specs",
		filepath.Join("changes", "archive"),
	} {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(tmp, "config.yaml"), []byte(configContents), 0o644); err != nil {
		return fmt.Errorf("write config.yaml: %w", err)
	}
	if err := os.Rename(tmp, root); err != nil {
		return fmt.Errorf("install speclet layout: %w", err)
	}
	return nil
}

// FindRoot searches start and its parents for the nearest valid speclet directory.
func FindRoot(start string) (string, error) {
	if start == "" {
		return "", fmt.Errorf("working directory is empty")
	}
	current, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	for {
		candidate := filepath.Join(current, "speclet")
		if _, err := os.Lstat(candidate); err == nil {
			if err := validateLayout(candidate); err != nil {
				return "", fmt.Errorf("invalid speclet layout at %s: %w", candidate, err)
			}
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("inspect %s: %w", candidate, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf("speclet project not found from %s", start)
}

func validateLayout(root string) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("speclet is not a directory")
	}

	configPath := filepath.Join(root, "config.yaml")
	configInfo, err := os.Lstat(configPath)
	if err != nil {
		return fmt.Errorf("config.yaml: %w", err)
	}
	if configInfo.Mode()&os.ModeSymlink != 0 || !configInfo.Mode().IsRegular() {
		return fmt.Errorf("config.yaml is not a regular file")
	}
	contents, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config.yaml: %w", err)
	}
	if !utf8.Valid(contents) {
		return fmt.Errorf("config.yaml is not valid UTF-8")
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
				return fmt.Errorf("config.yaml must declare version: 1")
			}
		}
	}
	if versionCount != 1 {
		return fmt.Errorf("config.yaml must declare version: 1")
	}
	for _, dir := range []string{"specs", "changes", filepath.Join("changes", "archive")} {
		path := filepath.Join(root, dir)
		info, err := os.Lstat(path)
		if err != nil {
			return fmt.Errorf("%s: %w", dir, err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("%s is not a directory", dir)
		}
	}
	return nil
}
