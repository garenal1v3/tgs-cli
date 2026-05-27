package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/searchtgcli/tgs/internal/config"
	"github.com/searchtgcli/tgs/internal/storage"
)

const configFileName = ".tgs.yaml"

type configFile struct {
	Profile string `yaml:"profile"`
}

// FindConfig walks up from startDir looking for a .tgs.yaml file.
// Returns the profile name from the file, or "" if not found.
func FindConfig(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, configFileName)
		data, err := os.ReadFile(path)
		if err == nil {
			var cfg configFile
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return "", fmt.Errorf("parse %s: %w", path, err)
			}
			return cfg.Profile, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

// Resolve returns the active profile name.
// Priority: flag > TGS_PROFILE env > .tgs.yaml file > "default".
func Resolve(flag string, cwd string) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("TGS_PROFILE"); env != "" {
		return env
	}
	if name, err := FindConfig(cwd); err == nil && name != "" {
		return name
	}
	return "default"
}

// Switch writes (or overwrites) .tgs.yaml in dir with the given profile name.
func Switch(name, dir string) error {
	cfg := configFile{Profile: name}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	path := filepath.Join(dir, configFileName)
	return os.WriteFile(path, data, 0o644)
}

// ProfileInfo holds metadata about a stored profile.
type ProfileInfo struct {
	Name     string
	Phone    string
	Username string
	UserID   string
}

// List returns all profiles found under the profiles directory.
func List() ([]ProfileInfo, error) {
	profilesDir := filepath.Join(config.DataDir(), "profiles")
	entries, err := os.ReadDir(profilesDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read profiles dir: %w", err)
	}

	var profiles []ProfileInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info := ProfileInfo{Name: e.Name()}
		dbPath := filepath.Join(profilesDir, e.Name(), "session.db")
		s, err := storage.NewSession(dbPath)
		if err != nil {
			profiles = append(profiles, info)
			continue
		}
		meta, _ := s.LoadAllMeta()
		_ = s.Close()
		info.Phone = meta["phone"]
		info.Username = meta["username"]
		info.UserID = meta["user_id"]
		profiles = append(profiles, info)
	}
	return profiles, nil
}

// Delete removes the profile directory for the given name.
// Returns an error if the profile does not exist.
func Delete(name string) error {
	dir := config.ProfileDir(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("profile %q does not exist", name)
	}
	return os.RemoveAll(dir)
}
