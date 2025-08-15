package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type SpotConfig struct {
	Org    string `yaml:"org"`
	Token  string `yaml:"token"`
	Region string `yaml:"region"`
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
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
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

func GetOrg(cmd *cobra.Command) (string, error) {

	org, _ := cmd.Flags().GetString("org")
	if org == "" {
		cfg, err := LoadConfig()
		if err == nil && cfg.Org != "" {
			org = cfg.Org
		}
	}
	if org == "" {
		return "", fmt.Errorf("organization not specified (use --org or run 'spotcli configure')")
	}
	return org, nil
}

func GetRegion(cmd *cobra.Command) (string, error) {

	region, _ := cmd.Flags().GetString("region")
	if region == "" {
		cfg, err := LoadConfig()
		if err == nil && cfg.Region != "" {
			region = cfg.Region
		}
	}
	if region == "" {
		return "", fmt.Errorf("region not specified (use --region or run 'spotcli configure')")
	}
	return region, nil
}
