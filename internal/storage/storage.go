package storage

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pinetoppeter/timetracker/internal/config"
	"github.com/pinetoppeter/timetracker/internal/record"
)

type Storage struct {
	baseDir string
}

const (
	defaultBaseDir = ".timetracker"
	sessionsDir    = "sessions"
	recordsDir     = "records"
)

func NewStorage() (*Storage, error) {
	// Try to load config to get data folder
	cfg, err := config.LoadConfig()
	if err != nil {
		// If config doesn't exist, the setup process should have been triggered
		// This should not happen in normal operation, but we'll use a sensible default
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not get home directory: %w", err)
		}
		// Don't create directories here - they'll be created when first used
		return &Storage{baseDir: filepath.Join(homeDir, defaultBaseDir)}, nil
	}

	// Use data folder from config
	dataFolder := cfg.DataFolder
	if dataFolder == "" {
		// Fallback to default location if data folder not configured
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not get home directory: %w", err)
		}
		dataFolder = filepath.Join(homeDir, defaultBaseDir)
	}

	// Create necessary directories in the data folder
	if err := os.MkdirAll(filepath.Join(dataFolder, sessionsDir), 0755); err != nil {
		return nil, fmt.Errorf("could not create sessions directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(dataFolder, recordsDir), 0755); err != nil {
		return nil, fmt.Errorf("could not create records directory: %w", err)
	}

	return &Storage{baseDir: dataFolder}, nil
}

func (s *Storage) GetBaseDir() string {
	return s.baseDir
}

func (s *Storage) getRecordPath(recordName string) string {
	slug := Slugify(recordName)
	return filepath.Join(s.baseDir, recordsDir, fmt.Sprintf("%s.json", slug))
}

func (s *Storage) getSessionCSVPath(sessionID string) string {
	// Parse session ID to extract date for monthly folder organization
	// Session ID format: session-YYYY-MM-DDTHH-MM-SS
	parts := strings.Split(sessionID, "-")
	if len(parts) >= 4 {
		// Extract year and month from session ID
		year := parts[1]
		month := parts[2]
		monthlyFolder := fmt.Sprintf("%s-%s", year, month)
		return filepath.Join(s.baseDir, sessionsDir, monthlyFolder, fmt.Sprintf("%s.csv", sessionID))
	}
	// Fallback to old format for compatibility
	return filepath.Join(s.baseDir, sessionsDir, fmt.Sprintf("%s.csv", sessionID))
}

func (s *Storage) ExportSessionToCSV(sess *Session) error {
	csvPath := s.getSessionCSVPath(sess.ID)

	// Create monthly directory if it doesn't exist
	sessionDir := filepath.Dir(csvPath)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("could not create monthly session directory: %w", err)
	}

	file, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("could not create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Record ID", "Name", "Start Time", "End Time", "Duration (minutes)"}); err != nil {
		return fmt.Errorf("could not write CSV header: %w", err)
	}

	// Write records
	for _, rec := range sess.Records {
		startTime := rec.StartTime.Format("2006-01-02 15:04:05")
		endTime := ""
		duration := 0

		if rec.EndTime != nil {
			endTime = rec.EndTime.Format("2006-01-02 15:04:05")
			duration = int(rec.EndTime.Sub(rec.StartTime).Minutes())
		} else {
			// If record is still running, calculate duration from start to now
			duration = int(time.Since(rec.StartTime).Minutes())
		}

		if err := writer.Write([]string{rec.ID, rec.Name, startTime, endTime, fmt.Sprintf("%d", duration)}); err != nil {
			return fmt.Errorf("could not write CSV record: %w", err)
		}
	}

	return nil
}

// RecordMetadata represents the metadata-only structure for record JSON files
// This excludes timestamps since they are stored in session files
type RecordMetadata struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (s *Storage) SaveRecordMetadata(rec *record.Record) error {
	if rec.Name == "" {
		return nil // Don't save metadata for unnamed records
	}

	// Create metadata-only structure (excluding timestamps)
	// Only include metadata if it has been explicitly set by the user
	metadata := RecordMetadata{
		Name: rec.Name,
	}

	// Only include metadata in the file if it has been explicitly set
	// This allows users to manually edit the JSON files to add metadata
	if len(rec.Metadata) > 0 {
		metadata.Metadata = rec.Metadata
	}

	recordPath := s.getRecordPath(rec.Name)
	content, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal record metadata: %w", err)
	}

	if err := os.WriteFile(recordPath, content, 0644); err != nil {
		return fmt.Errorf("could not write record metadata file: %w", err)
	}

	return nil
}

func (s *Storage) GetRecordNames() ([]string, error) {
	recordsDir := filepath.Join(s.baseDir, recordsDir)
	files, err := os.ReadDir(recordsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read records directory: %w", err)
	}

	var names []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if strings.HasSuffix(filename, ".json") {
			name := strings.TrimSuffix(filename, ".json")
			names = append(names, name)
		}
	}

	return names, nil
}

// LoadRecordMetadata loads the metadata for a specific record
func (s *Storage) LoadRecordMetadata(recordName string) (*RecordMetadata, error) {
	recordPath := s.getRecordPath(recordName)

	content, err := os.ReadFile(recordPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty metadata if file doesn't exist
			return &RecordMetadata{
				Name:     recordName,
				Metadata: make(map[string]interface{}),
			}, nil
		}
		return nil, fmt.Errorf("could not read record metadata file: %w", err)
	}

	var metadata RecordMetadata
	if err := json.Unmarshal(content, &metadata); err != nil {
		return nil, fmt.Errorf("could not unmarshal record metadata: %w", err)
	}

	return &metadata, nil
}

func Slugify(name string) string {
	// Enhanced slugify implementation with support for German characters
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Replace German umlauts and special characters with their ASCII equivalents
	name = strings.ReplaceAll(name, "ä", "ae")
	name = strings.ReplaceAll(name, "ö", "oe")
	name = strings.ReplaceAll(name, "ü", "ue")
	name = strings.ReplaceAll(name, "ß", "ss")

	// Remove any remaining non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result.WriteRune(c)
		}
	}

	return result.String()
}
