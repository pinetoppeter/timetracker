package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigPath(t *testing.T) {
	// Test that GetConfigPath returns the correct path in .timetracker directory
	configPath, err := GetConfigPath()
	require.NoError(t, err)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	expectedPath := filepath.Join(homeDir, ".timetracker", "config.json")
	assert.Equal(t, expectedPath, configPath)
}

func TestCreateDefaultConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "timetracker-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// Test creating a default config
	config, err := CreateDefaultConfig(configPath)
	require.NoError(t, err)

	// Verify the config has timezone set
	assert.NotEmpty(t, config.Timezone)
	// Timezone can be "Local" or a specific timezone like "UTC" or "America/New_York"
	assert.True(t, len(config.Timezone) > 0, "Timezone should not be empty")

	// Verify the config file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Verify the config file content
	fileContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	
	assert.Contains(t, string(fileContent), "timezone")
	assert.NotContains(t, string(fileContent), "rounding")
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "timetracker-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create .timetracker directory structure
	timetrackerDir := filepath.Join(tempDir, ".timetracker")
	err = os.MkdirAll(timetrackerDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(timetrackerDir, "config.json")

	// Create a test config file with a known timezone
	testConfig := `{
		"timezone": "UTC"
	}`
	
	err = os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Mock home directory to return our temp directory
	originalUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() { userHomeDir = originalUserHomeDir }()

	// Test loading the config
	config, err := LoadConfig()
	require.NoError(t, err)

	// Verify the config was loaded correctly
	assert.Equal(t, "UTC", config.Timezone)
}

func TestConfigGetLocation(t *testing.T) {
	// Test GetLocation method with different timezone values
	testCases := []struct {
		timezone string
		error    bool
	}{
		{"UTC", false},
		{"America/New_York", false},
		{"Europe/London", false},
		{"", false}, // Empty should default to UTC
	}

	for _, tc := range testCases {
		config := &Config{Timezone: tc.timezone}
		loc, err := config.GetLocation()
		
		if tc.error {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, loc)
		}
	}
}

// Helper function to mock GetConfigPath for testing
var getConfigPath = GetConfigPath

