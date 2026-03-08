package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Timezone   string `json:"timezone"`
	DataFolder string `json:"dataFolder"`
}

const (
	defaultConfigDir  = ".config/timetracker"
	defaultConfigFile = "config.json"
	altConfigDir      = ".timetracker"
)

// Make userHomeDir variable for testing
var userHomeDir = os.UserHomeDir

func GetConfigPath() (string, error) {
	homeDir, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}
	
	// Use only .timetracker directory for config (new unified location)
	configPath := filepath.Join(homeDir, altConfigDir, defaultConfigFile)
	
	// Create .timetracker directory if it doesn't exist
	configDirPath := filepath.Join(homeDir, altConfigDir)
	if err := os.MkdirAll(configDirPath, 0755); err != nil {
		return "", fmt.Errorf("could not create config directory: %w", err)
	}
	
	return configPath, nil
}

func GetConfigDir() (string, error) {
	homeDir, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}
	
	// Use only .timetracker directory for config (new unified location)
	configDirPath := filepath.Join(homeDir, altConfigDir)
	
	// Create .timetracker directory if it doesn't exist
	if err := os.MkdirAll(configDirPath, 0755); err != nil {
		return "", fmt.Errorf("could not create config directory: %w", err)
	}
	
	return configDirPath, nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	
	// Load existing config
	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file does not exist: %w", err)
		}
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	
	var config Config
	if err := json.Unmarshal(fileContent, &config); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}
	
	// No default values needed - timezone is the only field
	
	return &config, nil
}

// CreateDefaultConfig creates a new config with default values
func CreateDefaultConfig(path string) (*Config, error) {
	// Get system timezone
	loc := time.Now().Location().String()
	if loc == "" {
		loc = "UTC"
	}
	
	config := &Config{
		Timezone: loc,
	}
	
	// Save default config
	if err := SaveConfig(config, path); err != nil {
		return nil, err
	}
	
	return config, nil
}

func SaveConfig(config *Config, path string) error {
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}
	
	if err := os.WriteFile(path, configJSON, 0644); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}
	
	return nil
}

func (c *Config) GetLocation() (*time.Location, error) {
	if c.Timezone == "" {
		return time.UTC, nil
	}
	return time.LoadLocation(c.Timezone)
}