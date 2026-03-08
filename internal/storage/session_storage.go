package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Session represents a minimal session structure for storage purposes
type Session struct {
	ID            string     `json:"id"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	Records       []Record   `json:"records"`
	CurrentRecordID string   `json:"current_record_id"`
	IsPaused      bool       `json:"is_paused"`
}

// Record represents a minimal record structure for storage purposes
type Record struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func (s *Storage) getSessionPath(sessionID string) string {
	// Parse session ID to extract date for monthly folder organization
	// Session ID format: session-YYYY-MM-DDTHH-MM-SS
	parts := strings.Split(sessionID, "-")
	if len(parts) >= 4 {
		// Extract year and month from session ID
		year := parts[1]
		month := parts[2]
		monthlyFolder := fmt.Sprintf("%s-%s", year, month)
		return filepath.Join(s.baseDir, sessionsDir, monthlyFolder, fmt.Sprintf("%s.json", sessionID))
	}
	// Fallback to old format for compatibility
	return filepath.Join(s.baseDir, sessionsDir, fmt.Sprintf("%s.json", sessionID))
}

func (s *Storage) getCurrentSessionPath() string {
	return filepath.Join(s.baseDir, sessionsDir, "current.json")
}

func (s *Storage) LoadCurrentSession() (*Session, error) {
	currentSessionPath := s.getCurrentSessionPath()
	
	// Check if current session exists with new format
	if _, err := os.Stat(currentSessionPath); err == nil {
		content, err := os.ReadFile(currentSessionPath)
		if err != nil {
			return nil, fmt.Errorf("could not read current session: %w", err)
		}
		
		var sess Session
		if err := json.Unmarshal(content, &sess); err != nil {
			return nil, fmt.Errorf("could not parse session: %w", err)
		}
		
		// Verify that the actual session file still exists
		// This respects manual deletions by users
		sessionFilePath := s.getSessionPath(sess.ID)
		if _, err := os.Stat(sessionFilePath); err != nil {
			// Session file was manually deleted, treat as no current session
			return nil, fmt.Errorf("no current session found")
		}
		
		return &sess, nil
	}
	
	// Check for old format current session files
	files, err := os.ReadDir(filepath.Join(s.baseDir, sessionsDir))
	if err != nil {
		return nil, fmt.Errorf("could not read sessions directory: %w", err)
	}
	
	// Look for any current session files (old format)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		filename := file.Name()
		if strings.HasPrefix(filename, "current-") && strings.HasSuffix(filename, ".json") {
			oldSessionPath := filepath.Join(s.baseDir, sessionsDir, filename)
			content, err := os.ReadFile(oldSessionPath)
			if err != nil {
				continue // Skip files we can't read
			}
			
			var sess Session
			if err := json.Unmarshal(content, &sess); err != nil {
				continue // Skip files we can't parse
			}
			
			// Check if this session is still active (no end time)
			if sess.EndTime == nil {
				// Verify that the actual session file still exists
				// This respects manual deletions by users
				sessionFilePath := s.getSessionPath(sess.ID)
				if _, err := os.Stat(sessionFilePath); err != nil {
					// Session file was manually deleted, skip this session
					continue
				}
				
				// Migrate to new format by saving with new filename
				if err := s.SaveSession(&sess); err != nil {
					return nil, fmt.Errorf("could not migrate session to new format: %w", err)
				}
				// Remove old file
				if err := os.Remove(oldSessionPath); err != nil {
					// Log warning but don't fail
					fmt.Printf("Warning: could not remove old session file %s: %v\n", oldSessionPath, err)
				}
				return &sess, nil
			}
		}
	}
	
	return nil, fmt.Errorf("no current session found")
}

func (s *Storage) SaveSession(sess *Session) error {
	// Save as current session
	currentSessionPath := s.getCurrentSessionPath()
	content, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal session: %w", err)
	}
	
	if err := os.WriteFile(currentSessionPath, content, 0644); err != nil {
		return fmt.Errorf("could not write current session: %w", err)
	}
	
	// Also save with session ID (UTC timestamp format)
	sessionPath := s.getSessionPath(sess.ID)
	
	// Create monthly directory if it doesn't exist
	sessionDir := filepath.Dir(sessionPath)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("could not create monthly session directory: %w", err)
	}
	
	if err := os.WriteFile(sessionPath, content, 0644); err != nil {
		return fmt.Errorf("could not write session file: %w", err)
	}
	
	return nil
}

// GetAllSessions returns all sessions from the sessions directory
func (s *Storage) GetAllSessions() ([]*Session, error) {
	sessionsDir := filepath.Join(s.baseDir, sessionsDir)
	
	var sessions []*Session
	
	// First, check for sessions in the root sessions directory (old format)
	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read sessions directory: %w", err)
	}
	
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		// Skip current.json as it's just a pointer to the current session
		if file.Name() == "current.json" {
			continue
		}
		
		filePath := filepath.Join(sessionsDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			// Skip files that can't be read
			continue
		}
		
		var session Session
		if err := json.Unmarshal(content, &session); err != nil {
			// Skip files that can't be unmarshaled
			continue
		}
		
		sessions = append(sessions, &session)
	}
	
	// Then, check for sessions in monthly subdirectories (new format)
	monthlyDirs, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read sessions directory for monthly folders: %w", err)
	}
	
	for _, monthlyDir := range monthlyDirs {
		if !monthlyDir.IsDir() {
			continue
		}
		
		// Check if this looks like a monthly folder (YYYY-MM format)
		if !isMonthlyFolder(monthlyDir.Name()) {
			continue
		}
		
		monthlyPath := filepath.Join(sessionsDir, monthlyDir.Name())
		monthlyFiles, err := os.ReadDir(monthlyPath)
		if err != nil {
			// Skip directories that can't be read
			continue
		}
		
		for _, file := range monthlyFiles {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
				continue
			}
			
			filePath := filepath.Join(monthlyPath, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				// Skip files that can't be read
				continue
			}
			
			var session Session
			if err := json.Unmarshal(content, &session); err != nil {
				// Skip files that can't be unmarshaled
				continue
			}
			
			sessions = append(sessions, &session)
		}
	}
	
	return sessions, nil
}

// isMonthlyFolder checks if a folder name matches the YYYY-MM pattern
func isMonthlyFolder(name string) bool {
	parts := strings.Split(name, "-")
	if len(parts) != 2 {
		return false
	}
	
	// Check if both parts are numeric
	if len(parts[0]) != 4 || len(parts[1]) != 2 {
		return false
	}
	
	for _, c := range parts[0] {
		if c < '0' || c > '9' {
			return false
		}
	}
	
	for _, c := range parts[1] {
		if c < '0' || c > '9' {
			return false
		}
	}
	
	return true
}

// GetSessionsByMonthYear returns all sessions for a specific month and year
func (s *Storage) GetSessionsByMonthYear(month, year int) ([]*Session, error) {
	allSessions, err := s.GetAllSessions()
	if err != nil {
		return nil, fmt.Errorf("could not get all sessions: %w", err)
	}
	
	var filteredSessions []*Session
	
	for _, session := range allSessions {
		if session.StartTime.Month() == time.Month(month) && session.StartTime.Year() == year {
			filteredSessions = append(filteredSessions, session)
		}
	}
	
	return filteredSessions, nil
}