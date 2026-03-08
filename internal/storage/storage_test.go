package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pinetoppeter/timetracker/internal/record"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlugify(t *testing.T) {
	// Test the Slugify function with various inputs
	testCases := []struct {
		input    string
		expected string
	}{
		{"Test Record", "test-record"},
		{"Test_Record", "test-record"},
		{"TestRecord", "testrecord"},
		{"Test Record With Spaces", "test-record-with-spaces"},
		{"Test-Record-With-Dashes", "test-record-with-dashes"},
		{"Test_Record_With_Underscores", "test-record-with-underscores"},
		{"German Umlauts äöüß", "german-umlauts-aeoeuess"},
		{"Mixed CASE Record", "mixed-case-record"},
		{"Record123", "record123"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := Slugify(tc.input)
		assert.Equal(t, tc.expected, result, "Slugify(%q) should return %q", tc.input, tc.expected)
	}
}

func TestSaveRecordMetadata(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "timetracker-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a storage instance that uses our temp directory
	store := &Storage{baseDir: tempDir}

	// Create records directory
	recordsDir := filepath.Join(tempDir, "records")
	err = os.MkdirAll(recordsDir, 0755)
	require.NoError(t, err)

	// Test saving record metadata
	testRecord := &record.Record{
		ID:        "test-record-1",
		Name:      "test-record", // Name should already be slugified
		StartTime: time.Now(),
		Metadata: map[string]interface{}{
			"project": "Test Project",
			"client":  "Test Client",
		},
	}

	err = store.SaveRecordMetadata(testRecord)
	require.NoError(t, err)

	// Verify the record metadata file was created
	recordPath := filepath.Join(recordsDir, "test-record.json")
	_, err = os.Stat(recordPath)
	require.NoError(t, err)

	// Verify the content
	fileContent, err := os.ReadFile(recordPath)
	require.NoError(t, err)

	var metadata RecordMetadata
	err = json.Unmarshal(fileContent, &metadata)
	require.NoError(t, err)

	assert.Equal(t, "test-record", metadata.Name)
	assert.Equal(t, "Test Project", metadata.Metadata["project"])
	assert.Equal(t, "Test Client", metadata.Metadata["client"])
}

func TestLoadRecordMetadata(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "timetracker-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a storage instance that uses our temp directory
	store := &Storage{baseDir: tempDir}

	// Create records directory
	recordsDir := filepath.Join(tempDir, "records")
	err = os.MkdirAll(recordsDir, 0755)
	require.NoError(t, err)

	// Create a test record metadata file
	testMetadata := RecordMetadata{
		Name: "test-record",
		Metadata: map[string]interface{}{
			"project":  "Test Project",
			"priority": "high",
		},
	}

	metadataContent, err := json.MarshalIndent(testMetadata, "", "  ")
	require.NoError(t, err)

	recordPath := filepath.Join(recordsDir, "test-record.json")
	err = os.WriteFile(recordPath, metadataContent, 0644)
	require.NoError(t, err)

	// Test loading the record metadata
	loadedMetadata, err := store.LoadRecordMetadata("test-record")
	require.NoError(t, err)

	assert.Equal(t, "test-record", loadedMetadata.Name)
	assert.Equal(t, "Test Project", loadedMetadata.Metadata["project"])
	assert.Equal(t, "high", loadedMetadata.Metadata["priority"])
}

func TestLoadRecordMetadataNonExistent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "timetracker-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a storage instance that uses our temp directory
	store := &Storage{baseDir: tempDir}

	// Test loading non-existent record metadata
	loadedMetadata, err := store.LoadRecordMetadata("non-existent-record")
	require.NoError(t, err)

	assert.Equal(t, "non-existent-record", loadedMetadata.Name)
	assert.NotNil(t, loadedMetadata.Metadata)
	assert.Equal(t, 0, len(loadedMetadata.Metadata))
}
