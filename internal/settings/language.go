package settings

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chiisen/speclet/internal/i18n"
)

type userSettings struct {
	Language string `json:"language"`
}

// Language returns the saved locale, or English when the user has not set one.
func Language() (locale string, configured bool, err error) {
	path, err := path()
	if err != nil {
		return "", false, err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return i18n.English, false, nil
	}
	if err != nil {
		return "", false, i18n.Errorf("inspect user settings %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", false, i18n.Errorf("user settings path is not a regular file: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return "", false, i18n.Errorf("read user settings %s: %w", path, err)
	}
	defer file.Close()

	var value userSettings
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return "", false, i18n.Errorf("parse user settings: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = i18n.Errorf("multiple JSON values")
		}
		return "", false, i18n.Errorf("parse user settings: %w", err)
	}
	if !i18n.IsSupported(value.Language) {
		return "", false, i18n.Errorf("unsupported language %q; supported languages: %s", value.Language, supportedCodes())
	}
	return value.Language, true, nil
}

// SetLanguage writes a supported locale to the user's OS-specific config directory.
func SetLanguage(locale string) error {
	if !i18n.IsSupported(locale) {
		return i18n.Errorf("unsupported language %q; supported languages: %s", locale, supportedCodes())
	}
	path, err := path()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return i18n.Errorf("create user config directory: %w", err)
	}
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return i18n.Errorf("user settings path is not a regular file: %s", path)
		}
	} else if !os.IsNotExist(err) {
		return i18n.Errorf("inspect user settings %s: %w", path, err)
	}

	file, err := os.CreateTemp(dir, ".speclet-settings-*")
	if err != nil {
		return i18n.Errorf("create temporary settings file: %w", err)
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return i18n.Errorf("write settings file: %w", err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(userSettings{Language: locale}); err != nil {
		_ = file.Close()
		return i18n.Errorf("write settings file: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return i18n.Errorf("sync settings file: %w", err)
	}
	if err := file.Close(); err != nil {
		return i18n.Errorf("close settings file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return i18n.Errorf("install settings file: %w", err)
	}
	return nil
}

func path() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", i18n.Errorf("user config directory is unavailable: %w", err)
	}
	return filepath.Join(configDir, "speclet", "config.json"), nil
}

func supportedCodes() string {
	languages := i18n.SupportedLanguages()
	codes := make([]string, len(languages))
	for i, language := range languages {
		codes[i] = language.Code
	}
	return strings.Join(codes, ", ")
}
