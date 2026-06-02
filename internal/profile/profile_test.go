package profile

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/garenal1v3/tgs-cli/internal/config"
	"github.com/garenal1v3/tgs-cli/internal/storage"
)

func TestFindConfig_InCurrentDir(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, ".tgs.yaml")
	if err := os.WriteFile(cfgFile, []byte("profile: work"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	got, err := FindConfig(dir)
	if err != nil {
		t.Fatalf("FindConfig() error: %v", err)
	}
	if got != "work" {
		t.Errorf("FindConfig() = %q, want \"work\"", got)
	}
}

func TestFindConfig_WalksUp(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".tgs.yaml"), []byte("profile: parent"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	got, err := FindConfig(child)
	if err != nil {
		t.Fatalf("FindConfig() error: %v", err)
	}
	if got != "parent" {
		t.Errorf("FindConfig() = %q, want \"parent\"", got)
	}
}

func TestFindConfig_NotFound(t *testing.T) {
	dir := t.TempDir()
	got, err := FindConfig(dir)
	if err != nil {
		t.Fatalf("FindConfig() error: %v", err)
	}
	if got != "" {
		t.Errorf("FindConfig() = %q, want empty string", got)
	}
}

func TestResolve_FlagWins(t *testing.T) {
	t.Setenv("TGS_PROFILE", "env-profile")
	got := Resolve("flag-profile", t.TempDir())
	if got != "flag-profile" {
		t.Errorf("Resolve() = %q, want \"flag-profile\"", got)
	}
}

func TestResolve_EnvWins(t *testing.T) {
	t.Setenv("TGS_PROFILE", "env-profile")
	got := Resolve("", t.TempDir())
	if got != "env-profile" {
		t.Errorf("Resolve() = %q, want \"env-profile\"", got)
	}
}

func TestResolve_FileWins(t *testing.T) {
	t.Setenv("TGS_PROFILE", "")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".tgs.yaml"), []byte("profile: file-profile"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	got := Resolve("", dir)
	if got != "file-profile" {
		t.Errorf("Resolve() = %q, want \"file-profile\"", got)
	}
}

func TestResolve_Default(t *testing.T) {
	t.Setenv("TGS_PROFILE", "")
	got := Resolve("", t.TempDir())
	if got != "default" {
		t.Errorf("Resolve() = %q, want \"default\"", got)
	}
}

func TestSwitch_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := Switch("work", dir); err != nil {
		t.Fatalf("Switch() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".tgs.yaml"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	var cfg configFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if cfg.Profile != "work" {
		t.Errorf("config profile = %q, want \"work\"", cfg.Profile)
	}
}

func TestSwitch_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".tgs.yaml"), []byte("profile: old"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	if err := Switch("new", dir); err != nil {
		t.Fatalf("Switch() error: %v", err)
	}

	got, _ := FindConfig(dir)
	if got != "new" {
		t.Errorf("after Switch, FindConfig() = %q, want \"new\"", got)
	}
}

func TestList_Empty(t *testing.T) {
	t.Setenv("TGS_DATA_DIR", t.TempDir())
	profiles, err := List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("List() returned %d profiles, want 0", len(profiles))
	}
}

func TestList_WithProfiles(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("TGS_DATA_DIR", dataDir)

	for _, name := range []string{"default", "work"} {
		dbPath := filepath.Join(config.ProfileDir(name), "session.db")
		s, err := storage.NewSession(dbPath)
		if err != nil {
			t.Fatalf("NewSession(%s) error: %v", name, err)
		}
		if err := s.StoreMeta("phone", []byte("+100000000"+name)); err != nil {
			t.Fatalf("StoreMeta(%s) error: %v", name, err)
		}
		if err := s.Close(); err != nil {
			t.Fatalf("Close(%s) error: %v", name, err)
		}
	}

	profiles, err := List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("List() returned %d profiles, want 2", len(profiles))
	}
}

func TestDelete_RemovesProfile(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("TGS_DATA_DIR", dataDir)

	dbPath := filepath.Join(config.ProfileDir("deleteme"), "session.db")
	s, err := storage.NewSession(dbPath)
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	if err := Delete("deleteme"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	if _, err := os.Stat(config.ProfileDir("deleteme")); !os.IsNotExist(err) {
		t.Error("expected profile directory to be removed")
	}
}

func TestDelete_NonexistentProfile(t *testing.T) {
	t.Setenv("TGS_DATA_DIR", t.TempDir())
	err := Delete("ghost")
	if err == nil {
		t.Error("Delete() should return error for nonexistent profile")
	}
}
