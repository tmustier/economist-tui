package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Cookies []Cookie `json:"cookies"`
}

type Cookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

const (
	configDirName       = "economist-tui"
	legacyConfigDirName = "economist-cli"
)

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", configDirName)
}

func LegacyConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", legacyConfigDirName)
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func legacyConfigPath() string {
	return filepath.Join(LegacyConfigDir(), "config.json")
}

func BrowserDataDir() string {
	return filepath.Join(ConfigDir(), "browser-data")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			legacyData, legacyErr := os.ReadFile(legacyConfigPath())
			if legacyErr != nil {
				if os.IsNotExist(legacyErr) {
					return &Config{}, nil
				}
				return nil, legacyErr
			}
			data = legacyData
		} else {
			return nil, err
		}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0600)
}

func IsLoggedIn() bool {
	cfg, err := Load()
	if err != nil {
		return false
	}
	return len(cfg.Cookies) > 0
}
