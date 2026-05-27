package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir_Default(t *testing.T) {
	t.Setenv("TGS_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "tgs")
	got := ConfigDir()
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestConfigDir_EnvOverride(t *testing.T) {
	t.Setenv("TGS_CONFIG_DIR", "/tmp/tgs-config")
	got := ConfigDir()
	if got != "/tmp/tgs-config" {
		t.Errorf("ConfigDir() = %q, want /tmp/tgs-config", got)
	}
}

func TestDataDir_Default(t *testing.T) {
	t.Setenv("TGS_DATA_DIR", "")
	t.Setenv("XDG_DATA_HOME", "")
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".local", "share", "tgs")
	got := DataDir()
	if got != want {
		t.Errorf("DataDir() = %q, want %q", got, want)
	}
}

func TestDataDir_EnvOverride(t *testing.T) {
	t.Setenv("TGS_DATA_DIR", "/tmp/tgs-data")
	got := DataDir()
	if got != "/tmp/tgs-data" {
		t.Errorf("DataDir() = %q, want /tmp/tgs-data", got)
	}
}

func TestProfileDir(t *testing.T) {
	t.Setenv("TGS_DATA_DIR", "/tmp/tgs-data")
	got := ProfileDir("work")
	want := "/tmp/tgs-data/profiles/work"
	if got != want {
		t.Errorf("ProfileDir(\"work\") = %q, want %q", got, want)
	}
}

func TestActiveProfile_Default(t *testing.T) {
	t.Setenv("TGS_PROFILE", "")
	got := ActiveProfile("")
	if got != "default" {
		t.Errorf("ActiveProfile(\"\") = %q, want \"default\"", got)
	}
}

func TestActiveProfile_EnvOverride(t *testing.T) {
	t.Setenv("TGS_PROFILE", "work")
	got := ActiveProfile("")
	if got != "work" {
		t.Errorf("ActiveProfile(\"\") = %q, want \"work\"", got)
	}
}

func TestActiveProfile_FlagOverride(t *testing.T) {
	t.Setenv("TGS_PROFILE", "work")
	got := ActiveProfile("personal")
	if got != "personal" {
		t.Errorf("ActiveProfile(\"personal\") = %q, want \"personal\"", got)
	}
}
