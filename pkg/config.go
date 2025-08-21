package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type SpotConfig struct {
	Org          string `yaml:"org"`
	RefreshToken string `yaml:"refresh_token"`
	AccessToken  string `yaml:"access_token"`
	Region       string `yaml:"region"`
}

// GetConfigPath returns the ~/.spot_config path
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".spot_config"), nil
}

func LoadConfig() (*SpotConfig, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("spot config not found, run 'spotcli configure' to configure your default orgID, token, and region")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return nil, fmt.Errorf("spot config not found, run 'spotcli configure' to configure your default orgID, token, and region")
		}
		return nil, err
	}

	var cfg SpotConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *SpotConfig) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600) // 600 = rw-------
}

func GetCLIEssentials(cmd *cobra.Command) (*SpotConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
