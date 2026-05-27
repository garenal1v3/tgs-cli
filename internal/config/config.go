package config

import (
	"os"
	"path/filepath"
)

const appName = "tgs"

func ConfigDir() string {
	if dir := os.Getenv("TGS_CONFIG_DIR"); dir != "" {
		return dir
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", appName)
}

func DataDir() string {
	if dir := os.Getenv("TGS_DATA_DIR"); dir != "" {
		return dir
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", appName)
}

func ProfileDir(name string) string {
	return filepath.Join(DataDir(), "profiles", name)
}

// ActiveProfile resolves the active profile name.
// Priority: flag > env > "default".
func ActiveProfile(flag string) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("TGS_PROFILE"); env != "" {
		return env
	}
	return "default"
}
