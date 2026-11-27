package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRootRequiresAValidSpecletDirectory(t *testing.T) {
	t.Run("missing project", func(t *testing.T) {
		start := t.TempDir()
		_, err := FindRoot(start)
		if err == nil || !strings.Contains(err.Error(), "speclet") {
			t.Fatalf("FindRoot() error = %v, want an error mentioning speclet", err)
		}
	})

	t.Run("nearest invalid project is not skipped", func(t *testing.T) {
		outer := t.TempDir()
		if err := Init(outer); err != nil {
			t.Fatal(err)
		}
		inner := filepath.Join(outer, "nested")
		if err := os.MkdirAll(filepath.Join(inner, "speclet"), 0o755); err != nil {
			t.Fatal(err)
		}

		_, err := FindRoot(inner)
		if err == nil || !strings.Contains(err.Error(), "invalid") {
			t.Fatalf("FindRoot() error = %v, want invalid nearest speclet error", err)
		}
	})
}

func TestInitRejectsInvalidExistingLayoutWithoutChangingIt(t *testing.T) {
	target := t.TempDir()
	root := filepath.Join(target, "speclet")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(root, "config.yaml")
	const original = "user data\n"
	if err := os.WriteFile(config, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Init(target); err == nil {
		t.Fatal("Init() error = nil, want invalid existing layout error")
	}
	got, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Fatalf("config.yaml = %q, want original %q", got, original)
	}
}

func TestFindRootUsesNearestValidProject(t *testing.T) {
	outer := t.TempDir()
	if err := Init(outer); err != nil {
		t.Fatal(err)
	}
	inner := filepath.Join(outer, "child with spaces", "巢狀")
	if err := Init(inner); err != nil {
		t.Fatal(err)
	}
	start := filepath.Join(inner, "work")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := FindRoot(start)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(inner, "speclet")
	if got != want {
		t.Fatalf("FindRoot() = %q, want %q", got, want)
	}
}

func TestInitCreatesLayoutAndPreservesExistingFiles(t *testing.T) {
	target := filepath.Join(t.TempDir(), "project with spaces", "規格")
	if err := Init(target); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(target, "speclet")
	for _, path := range []string{
		"config.yaml",
		"specs",
		"changes",
		filepath.Join("changes", "archive"),
	} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Errorf("expected %s to exist: %v", path, err)
		}
	}
	config := filepath.Join(root, "config.yaml")
	data, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "version: 1") {
		t.Fatalf("config.yaml = %q, want version: 1", data)
	}
	const custom = "version: 1\ncustom: keep\n"
	if err := os.WriteFile(config, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Init(target); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != custom {
		t.Fatalf("second Init() changed config.yaml: got %q, want %q", data, custom)
	}
}

func TestFindRootAndInitRejectManagedSymlinks(t *testing.T) {
	actual := t.TempDir()
	if err := Init(actual); err != nil {
		t.Fatal(err)
	}
	linkedProject := t.TempDir()
	if err := os.Symlink(filepath.Join(actual, "speclet"), filepath.Join(linkedProject, "speclet")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := FindRoot(linkedProject); err == nil {
		t.Fatal("FindRoot() accepted a symlinked speclet root")
	}
	if err := Init(linkedProject); err == nil {
		t.Fatal("Init() accepted a symlinked speclet root")
	}

	root := filepath.Join(actual, "speclet")
	outside := t.TempDir()
	if err := os.Remove(filepath.Join(root, "specs")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "specs")); err != nil {
		t.Fatal(err)
	}
	if _, err := FindRoot(actual); err == nil {
		t.Fatal("FindRoot() accepted a symlinked specs directory")
	}
}

func TestFindRootRejectsUnsupportedOrAmbiguousConfigVersion(t *testing.T) {
	for _, config := range []string{
		"version: 2\nmetadata:\n  version: 1\n",
		"version: 1\nversion: 1\n",
	} {
		t.Run(strings.ReplaceAll(strings.TrimSpace(config), "\n", "/"), func(t *testing.T) {
			projectDir := t.TempDir()
			root := filepath.Join(projectDir, "speclet")
			for _, dir := range []string{"specs", filepath.Join("changes", "archive")} {
				if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte(config), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := FindRoot(projectDir); err == nil {
				t.Fatalf("FindRoot() accepted config %q", config)
			}
		})
	}
}

func TestInitAcceptsValidVersionWithYAMLComment(t *testing.T) {
	projectDir := t.TempDir()
	if err := Init(projectDir); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(projectDir, "speclet", "config.yaml")
	const expected = "version: 1 # current format\ncustom: preserved\n"
	if err := os.WriteFile(config, []byte(expected), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Init(projectDir); err != nil {
		t.Fatalf("Init() rejected valid YAML comment: %v", err)
	}
	got, err := os.ReadFile(config)
	if err != nil || string(got) != expected {
		t.Fatalf("Init() changed config: %q, err=%v", got, err)
	}
}

func TestAcquireWriteLockSerializesWriters(t *testing.T) {
	projectDir := t.TempDir()
	if err := Init(projectDir); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(projectDir, "speclet")
	release, err := AcquireWriteLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcquireWriteLock(root); err == nil {
		t.Fatal("AcquireWriteLock() succeeded while another lock is held")
	}
	if err := release(); err != nil {
		t.Fatal(err)
	}
	release, err = AcquireWriteLock(root)
	if err != nil {
		t.Fatalf("AcquireWriteLock() after release: %v", err)
	}
	if err := release(); err != nil {
		t.Fatal(err)
	}
}
