package session

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pinetoppeter/timetracker/internal/config"
	"github.com/pinetoppeter/timetracker/internal/record"
	storagelib "github.com/pinetoppeter/timetracker/internal/storage"
)

// SessionStartResult contains information about a started session
type SessionStartResult struct {
	Session      *Session
	RecordName   string
	IsNewSession bool
}

// SessionEndResult contains information about an ended session
type SessionEndResult struct {
	Session      *Session
	RecordName   string
	RecordsNamed int
}

type Session struct {
	ID              string
	StartTime       time.Time
	EndTime         *time.Time
	Records         []*record.Record
	CurrentRecordID string
	IsPaused        bool
}

type SessionManager struct {
	config  *config.Config
	storage *storagelib.Storage
}

var (
	ErrNoActiveSession = errors.New("no active session")
	ErrSessionPaused   = errors.New("session is paused")
)

func NewSessionManager(cfg *config.Config, store *storagelib.Storage) *SessionManager {
	return &SessionManager{
		config:  cfg,
		storage: store,
	}
}

func (sm *SessionManager) loadCurrentSession() (*Session, error) {
	storageSession, err := sm.storage.LoadCurrentSession()
	if err != nil {
		return nil, err
	}

	// Convert storage session to our session type
	session := &Session{
		ID:              storageSession.ID,
		StartTime:       storageSession.StartTime,
		EndTime:         storageSession.EndTime,
		CurrentRecordID: storageSession.CurrentRecordID,
		IsPaused:        storageSession.IsPaused,
	}

	// Convert records
	for _, storageRecord := range storageSession.Records {
		rec := &record.Record{
			ID:        storageRecord.ID,
			Name:      storageRecord.Name,
			StartTime: storageRecord.StartTime,
			EndTime:   storageRecord.EndTime,
			Metadata:  storageRecord.Metadata,
		}
		session.Records = append(session.Records, rec)
	}

	return session, nil
}

func (sm *SessionManager) saveSession(sess *Session) error {
	// Convert our session to storage session type
	storageSession := &storagelib.Session{
		ID:              sess.ID,
		StartTime:       sess.StartTime,
		EndTime:         sess.EndTime,
		CurrentRecordID: sess.CurrentRecordID,
		IsPaused:        sess.IsPaused,
	}

	// Convert records - exclude metadata from session files
	// Metadata should only be stored in individual record JSON files
	for _, rec := range sess.Records {
		storageRecord := storagelib.Record{
			ID:        rec.ID,
			Name:      rec.Name,
			StartTime: rec.StartTime,
			EndTime:   rec.EndTime,
			Metadata:  nil, // Explicitly set to nil to exclude from session files
		}
		storageSession.Records = append(storageSession.Records, storageRecord)
	}

	return sm.storage.SaveSession(storageSession)
}

func (sm *SessionManager) StartSession() (*SessionStartResult, error) {
	return sm.StartSessionWithName("")
}

func (sm *SessionManager) StartSessionWithName(name string) (*SessionStartResult, error) {
	// Check if there's already an active session for today
	session, err := sm.loadCurrentSession()
	if err == nil {
		// Check if the existing session has been ended (stopped)
		// If it has an EndTime, it was stopped and we should create a new session
		if session.EndTime != nil {
			// Existing session was stopped, so we'll create a new one
			// (fall through to the new session creation code below)
		} else {
			// Existing session is still active, resume it
			// Get current record name for existing session
			currentRecord := sm.getCurrentRecordFromSession(session)
			recordName := ""
			if currentRecord != nil && currentRecord.Name != "" {
				recordName = currentRecord.Name
			}
			return &SessionStartResult{
				Session:      session,
				RecordName:   recordName,
				IsNewSession: false,
			}, nil
		}
	}

	// Create new session
	loc, err := sm.config.GetLocation()
	if err != nil {
		return nil, fmt.Errorf("could not get timezone: %w", err)
	}

	now := time.Now().In(loc)
	// Use UTC datetime string with seconds and nanoseconds for session ID to ensure uniqueness
	utcNow := time.Now().UTC()
	session = &Session{
		ID:        fmt.Sprintf("session-%s-%d", utcNow.Format("2006-01-02T15-04-05"), utcNow.UnixNano()),
		StartTime: now,
		Records:   make([]*record.Record, 0),
		IsPaused:  false,
	}

	// Create first record with the provided name
	firstRecord := record.NewRecord(name, now)
	session.Records = append(session.Records, firstRecord)
	session.CurrentRecordID = firstRecord.ID

	// Save session
	if err := sm.saveSession(session); err != nil {
		return nil, fmt.Errorf("could not save session: %w", err)
	}

	// Use the provided name (no prompting - users can provide name via command argument)
	if name != "" {
		firstRecord.Name = storagelib.Slugify(name)
	}

	// Preserve existing metadata when creating a record with an existing name
	if firstRecord.Name != "" {
		// Load existing metadata if it exists
		existingMetadata, err := sm.storage.LoadRecordMetadata(firstRecord.Name)
		if err == nil && existingMetadata != nil && len(existingMetadata.Metadata) > 0 {
			// Preserve existing metadata
			firstRecord.Metadata = existingMetadata.Metadata
		}
		
		if err := sm.storage.SaveRecordMetadata(firstRecord); err != nil {
			return nil, fmt.Errorf("could not save record metadata: %w", err)
		}
	}

	// Save updated session
	if err := sm.saveSession(session); err != nil {
		return nil, fmt.Errorf("could not save session after naming record: %w", err)
	}

	return &SessionStartResult{
		Session:      session,
		RecordName:   firstRecord.Name,
		IsNewSession: true,
	}, nil
}

func (sm *SessionManager) PauseSession() error {
	// Load current session
	session, err := sm.loadCurrentSession()
	if err != nil {
		return ErrNoActiveSession
	}

	// End the current record if it's running
	currentRecord := sm.getCurrentRecordFromSession(session)
	if currentRecord != nil && currentRecord.EndTime == nil {
		loc, err := sm.config.GetLocation()
		if err != nil {
			return fmt.Errorf("could not get timezone: %w", err)
		}
		now := time.Now().In(loc)
		currentRecord.EndTime = &now

		// Save record metadata
		if err := sm.storage.SaveRecordMetadata(currentRecord); err != nil {
			return fmt.Errorf("could not save record metadata: %w", err)
		}
	}

	// Save session state
	if err := sm.saveSession(session); err != nil {
		return fmt.Errorf("could not save paused session: %w", err)
	}

	return nil
}

// AddMetadataToCurrentRecord adds metadata to the current running record
func (sm *SessionManager) AddMetadataToCurrentRecord(key, value string) error {
	// Load current session
	session, err := sm.loadCurrentSession()
	if err != nil {
		return ErrNoActiveSession
	}

	currentRecord := sm.getCurrentRecordFromSession(session)
	if currentRecord == nil {
		return errors.New("no current record")
	}

	// Load existing metadata from record JSON file if it exists
	if currentRecord.Name != "" {
		recordMetadata, err := sm.storage.LoadRecordMetadata(currentRecord.Name)
		if err == nil && recordMetadata != nil && len(recordMetadata.Metadata) > 0 {
			// Use existing metadata from record JSON file
			currentRecord.Metadata = recordMetadata.Metadata
		}
	}

	// Initialize metadata map if still nil
	if currentRecord.Metadata == nil {
		currentRecord.Metadata = make(map[string]interface{})
	}

	// Parse the value based on its format
	var parsedValue interface{}

	// Try to parse as boolean
	if value == "true" || value == "True" || value == "TRUE" {
		parsedValue = true
	} else if value == "false" || value == "False" || value == "FALSE" {
		parsedValue = false
		// Try to parse as integer
	} else if intVal, err := strconv.Atoi(value); err == nil {
		parsedValue = intVal
		// Try to parse as float
	} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		parsedValue = floatVal
	} else {
		// Treat as string
		parsedValue = value
	}

	currentRecord.Metadata[key] = parsedValue

	// Save the updated session (metadata will be nil in session file)
	if err := sm.saveSession(session); err != nil {
		return fmt.Errorf("could not save session after adding metadata: %w", err)
	}

	// Also save record metadata if the record has a name
	if currentRecord.Name != "" {
		if err := sm.storage.SaveRecordMetadata(currentRecord); err != nil {
			return fmt.Errorf("could not save record metadata: %w", err)
		}
	}

	return nil
}

// AddMetadataToRecord adds metadata to a specific record by name
func (sm *SessionManager) AddMetadataToRecord(recordName, key, value string) error {
	// Load the record metadata
	recordMetadata, err := sm.storage.LoadRecordMetadata(recordName)
	if err != nil {
		return fmt.Errorf("could not load record metadata: %w", err)
	}

	// Initialize metadata map if nil
	if recordMetadata.Metadata == nil {
		recordMetadata.Metadata = make(map[string]interface{})
	}

	// Parse the value based on its format
	var parsedValue interface{}

	// Try to parse as boolean
	if value == "true" || value == "True" || value == "TRUE" {
		parsedValue = true
	} else if value == "false" || value == "False" || value == "FALSE" {
		parsedValue = false
		// Try to parse as integer
	} else if intVal, err := strconv.Atoi(value); err == nil {
		parsedValue = intVal
		// Try to parse as float
	} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		parsedValue = floatVal
	} else {
		// Treat as string
		parsedValue = value
	}

	recordMetadata.Metadata[key] = parsedValue

	// Save the updated record metadata
	if err := sm.storage.SaveRecordMetadata(&record.Record{
		Name:     recordName,
		Metadata: recordMetadata.Metadata,
	}); err != nil {
		return fmt.Errorf("could not save record metadata: %w", err)
	}

	// Add the metadata field to export schema if it doesn't exist
	// Determine the field type based on the parsed value
	fieldType := "string" // default
	switch parsedValue.(type) {
	case bool:
		fieldType = "boolean"
	case int, int32, int64:
		fieldType = "number"
	case float32, float64:
		fieldType = "number"
	}

	if err := sm.storage.AddMetadataFieldToSchema(key, fieldType); err != nil {
		// Log error but don't fail the metadata addition
		fmt.Printf("Warning: Could not update export schema for field '%s': %v\n", key, err)
	}

	return nil
}

// GetCurrentRecordMetadata gets the metadata for the current record
// Checks both the session metadata and the record JSON file
func (sm *SessionManager) GetCurrentRecordMetadata() (map[string]interface{}, error) {
	// Load current session
	session, err := sm.loadCurrentSession()
	if err != nil {
		return nil, ErrNoActiveSession
	}

	currentRecord := sm.getCurrentRecordFromSession(session)
	if currentRecord == nil {
		return nil, errors.New("no current record")
	}

	// If record has no name, return empty metadata
	if currentRecord.Name == "" {
		return make(map[string]interface{}), nil
	}

	// First, try to load metadata from the record JSON file
	// This ensures manual edits to the JSON file are respected
	recordMetadata, err := sm.storage.LoadRecordMetadata(currentRecord.Name)
	if err == nil && recordMetadata != nil && len(recordMetadata.Metadata) > 0 {
		// Return metadata from the JSON file (includes manual edits)
		return recordMetadata.Metadata, nil
	}

	// Fall back to session metadata if no JSON metadata exists
	if currentRecord.Metadata == nil {
		return make(map[string]interface{}), nil
	}

	return currentRecord.Metadata, nil
}

// GetRecordMetadata gets the metadata for a specific record by name
func (sm *SessionManager) GetRecordMetadata(recordName string) (map[string]interface{}, error) {
	// Load metadata from the record JSON file
	recordMetadata, err := sm.storage.LoadRecordMetadata(recordName)
	if err != nil {
		return nil, fmt.Errorf("could not load record metadata: %w", err)
	}

	if recordMetadata.Metadata == nil {
		return make(map[string]interface{}), nil
	}

	return recordMetadata.Metadata, nil
}

func (sm *SessionManager) ResumeSession() error {
	// Load current session
	session, err := sm.loadCurrentSession()
	if err != nil {
		return ErrNoActiveSession
	}

	if !session.IsPaused {
		return nil // Already running
	}

	session.IsPaused = false

	// Save session state
	if err := sm.saveSession(session); err != nil {
		return fmt.Errorf("could not save resumed session: %w", err)
	}

	return nil
}

func (sm *SessionManager) ResumeLastRecord() error {
	// Load current session
	session, err := sm.loadCurrentSession()
	if err != nil {
		return ErrNoActiveSession
	}

	// Find the last named record
	var lastNamedRecord *record.Record
	for i := len(session.Records) - 1; i >= 0; i-- {
		rec := session.Records[i]
		if rec.Name != "" {
			lastNamedRecord = rec
			break
		}
	}

	if lastNamedRecord == nil {
		return errors.New("no named records found in current session")
	}

	// End the current record if it's running
	currentRecord := sm.getCurrentRecordFromSession(session)
	if currentRecord != nil && currentRecord.EndTime == nil {
		loc, err := sm.config.GetLocation()
		if err != nil {
			return fmt.Errorf("could not get timezone: %w", err)
		}
		now := time.Now().In(loc)
		currentRecord.EndTime = &now

		// Save record metadata
		if err := sm.storage.SaveRecordMetadata(currentRecord); err != nil {
			return fmt.Errorf("could not save record metadata: %w", err)
		}
	}

	// Create new record with the same name as the last named record
	loc, err := sm.config.GetLocation()
	if err != nil {
		return fmt.Errorf("could not get timezone: %w", err)
	}

	now := time.Now().In(loc)
	newRecord := record.NewRecord(lastNamedRecord.Name, now)
	session.Records = append(session.Records, newRecord)
	session.CurrentRecordID = newRecord.ID

	// Save record metadata
	if err := sm.storage.SaveRecordMetadata(newRecord); err != nil {
		return fmt.Errorf("could not save record metadata: %w", err)
	}

	// Save session
	if err := sm.saveSession(session); err != nil {
		return fmt.Errorf("could not save session after resuming record: %w", err)
	}

	return nil
}

func (sm *SessionManager) SwitchRecord() error {
	return sm.SwitchAndNameRecord("")
}

func (sm *SessionManager) SwitchAndNameRecord(name string) error {
	// Load current session
	session, err := sm.loadCurrentSession()
	if err != nil {
		return ErrNoActiveSession
	}

	if session.IsPaused {
		// Automatically resume if paused
		session.IsPaused = false
	}

	loc, err := sm.config.GetLocation()
	if err != nil {
		return fmt.Errorf("could not get timezone: %w", err)
	}

	now := time.Now().In(loc)

	// End current record
	currentRecord := sm.getCurrentRecordFromSession(session)
	if currentRecord != nil {
		currentRecord.EndTime = &now
	}

	// Create new record with the provided name
	newRecord := record.NewRecord(name, now)
	session.Records = append(session.Records, newRecord)
	session.CurrentRecordID = newRecord.ID

	// Save session
	if err := sm.saveSession(session); err != nil {
		return fmt.Errorf("could not save session after switching record: %w", err)
	}

	// Save record metadata if the new record has a name
	if name != "" {
		if err := sm.storage.SaveRecordMetadata(newRecord); err != nil {
			return fmt.Errorf("could not save record metadata: %w", err)
		}
	}

	return nil
}

func (sm *SessionManager) NameCurrentRecord(name string) error {
	// Load current session
	session, err := sm.loadCurrentSession()
	if err != nil {
		return ErrNoActiveSession
	}

	currentRecord := sm.getCurrentRecordFromSession(session)
	if currentRecord == nil {
		return errors.New("no current record")
	}

	currentRecord.Name = storagelib.Slugify(name)

	// Save record metadata
	if err := sm.storage.SaveRecordMetadata(currentRecord); err != nil {
		return fmt.Errorf("could not save record metadata: %w", err)
	}

	// Save session
	if err := sm.saveSession(session); err != nil {
		return fmt.Errorf("could not save session after naming record: %w", err)
	}

	return nil
}

func (sm *SessionManager) EndSession() (*SessionEndResult, error) {
	// Load current session
	session, err := sm.loadCurrentSession()
	if err != nil {
		return nil, ErrNoActiveSession
	}

	loc, err := sm.config.GetLocation()
	if err != nil {
		return nil, fmt.Errorf("could not get timezone: %w", err)
	}

	now := time.Now().In(loc)

	// End current record if it's running
	currentRecord := sm.getCurrentRecordFromSession(session)
	currentRecordName := ""
	if currentRecord != nil {
		if currentRecord.EndTime == nil {
			currentRecord.EndTime = &now
		}
		currentRecordName = currentRecord.Name
	}

	// Prompt for unnamed records before ending session (non-blocking)
	recordsNamed, err := sm.promptForUnnamedRecords()
	if err != nil {
		// Log warning but don't fail - allow session to end even if records remain unnamed
		fmt.Printf("Warning: could not complete record naming (running in non-interactive mode): %v\n", err)
	} else if recordsNamed > 0 {
		// Reload session to get the updated record names from promptForUnnamedRecords
		session, err = sm.loadCurrentSession()
		if err != nil {
			return nil, fmt.Errorf("could not reload session after naming records: %w", err)
		}
		// Update current record and ensure it has end time after reload
		currentRecord = sm.getCurrentRecordFromSession(session)
		if currentRecord != nil {
			currentRecordName = currentRecord.Name
			// Ensure the current record has an end time (might have been lost during reload)
			if currentRecord.EndTime == nil {
				currentRecord.EndTime = &now
			}
		}
	}

	// End session
	session.EndTime = &now

	// Save final session state
	if err := sm.saveSession(session); err != nil {
		return nil, fmt.Errorf("could not save ended session: %w", err)
	}

	// Export session to CSV
	storageSession := &storagelib.Session{
		ID:              session.ID,
		StartTime:       session.StartTime,
		EndTime:         session.EndTime,
		CurrentRecordID: session.CurrentRecordID,
		IsPaused:        session.IsPaused,
	}

	// Convert records - exclude metadata from CSV export session data
	// Metadata should only be stored in individual record JSON files
	for _, rec := range session.Records {
		storageRecord := storagelib.Record{
			ID:        rec.ID,
			Name:      rec.Name,
			StartTime: rec.StartTime,
			EndTime:   rec.EndTime,
			Metadata:  nil, // Explicitly set to nil to exclude from CSV export
		}
		storageSession.Records = append(storageSession.Records, storageRecord)
	}

	if err := sm.storage.ExportSessionToCSV(storageSession); err != nil {
		return nil, fmt.Errorf("could not export session to CSV: %w", err)
	}

	return &SessionEndResult{
		Session:      session,
		RecordName:   currentRecordName,
		RecordsNamed: recordsNamed,
	}, nil
}

func (sm *SessionManager) EndSessionWithRecordNames(names []string) (*SessionEndResult, error) {
	// Load current session
	session, err := sm.loadCurrentSession()
	if err != nil {
		return nil, ErrNoActiveSession
	}

	loc, err := sm.config.GetLocation()
	if err != nil {
		return nil, fmt.Errorf("could not get timezone: %w", err)
	}

	now := time.Now().In(loc)

	// End current record if it's running
	currentRecord := sm.getCurrentRecordFromSession(session)
	currentRecordName := ""
	if currentRecord != nil {
		if currentRecord.EndTime == nil {
			currentRecord.EndTime = &now
		}
		currentRecordName = currentRecord.Name
	}

	// Name unnamed records using provided names
	recordsNamed := 0
	unnamedRecords := sm.getUnnamedRecordsFromSession(session)

	for i, rec := range unnamedRecords {
		if i < len(names) && names[i] != "" {
			rec.Name = storagelib.Slugify(names[i])

			// Save record metadata
			if err := sm.storage.SaveRecordMetadata(rec); err != nil {
				return nil, fmt.Errorf("could not save record metadata: %w", err)
			}
			recordsNamed++
		}
	}

	// End session
	session.EndTime = &now

	// Save final session state
	if err := sm.saveSession(session); err != nil {
		return nil, fmt.Errorf("could not save ended session: %w", err)
	}

	// Export session to CSV
	storageSession := &storagelib.Session{
		ID:              session.ID,
		StartTime:       session.StartTime,
		EndTime:         session.EndTime,
		CurrentRecordID: session.CurrentRecordID,
		IsPaused:        session.IsPaused,
	}

	// Convert records - exclude metadata from CSV export session data
	// Metadata should only be stored in individual record JSON files
	for _, rec := range session.Records {
		storageRecord := storagelib.Record{
			ID:        rec.ID,
			Name:      rec.Name,
			StartTime: rec.StartTime,
			EndTime:   rec.EndTime,
			// Metadata is not included in CSV export
		}
		storageSession.Records = append(storageSession.Records, storageRecord)
	}

	if err := sm.storage.ExportSessionToCSV(storageSession); err != nil {
		return nil, fmt.Errorf("could not export session to CSV: %w", err)
	}

	return &SessionEndResult{
		Session:      session,
		RecordName:   currentRecordName,
		RecordsNamed: recordsNamed,
	}, nil
}

func (sm *SessionManager) getUnnamedRecordsFromSession(session *Session) []*record.Record {
	var unnamedRecords []*record.Record
	for _, rec := range session.Records {
		if rec.Name == "" {
			unnamedRecords = append(unnamedRecords, rec)
		}
	}
	return unnamedRecords
}

func (sm *SessionManager) promptForRecordName(rec *record.Record) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Get autocomplete suggestions
	suggestions, err := sm.storage.GetRecordNames()
	if err != nil {
		fmt.Printf("Warning: could not load suggestions: %v\n", err)
	}

	// Show suggestions upfront if available (similar to switch command)
	if len(suggestions) > 0 {
		fmt.Println("\nAvailable record names (type number or name):")
		for i, suggestion := range suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
		fmt.Println("Or type a new name:")
	}

	for {
		fmt.Printf("Enter name for record (started at %s): ", rec.StartTime.Format("15:04:05"))
		name, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("could not read input: %w", err)
		}

		name = strings.TrimSpace(name)

		// Check if user entered a number to select from suggestions
		if len(suggestions) > 0 {
			// Try to parse as number for suggestion selection
			if selectedNum, err := strconv.Atoi(name); err == nil && selectedNum > 0 && selectedNum <= len(suggestions) {
				return suggestions[selectedNum-1], nil
			}
		}

		if name != "" {
			return name, nil
		}

		fmt.Println("Record name cannot be empty. Please try again.")

		// Show suggestions again if available
		if len(suggestions) > 0 {
			fmt.Println("\nAvailable suggestions (type number or name):")
			for i, suggestion := range suggestions {
				fmt.Printf("  %d. %s\n", i+1, suggestion)
			}
			fmt.Println("Or type a new name:")
		}
	}
}

func (sm *SessionManager) promptForUnnamedRecords() (int, error) {
	// Load current session
	session, err := sm.loadCurrentSession()
	if err != nil {
		return 0, ErrNoActiveSession
	}

	// Find all unnamed records in chronological order
	var unnamedRecords []*record.Record
	for _, rec := range session.Records {
		if rec.Name == "" {
			unnamedRecords = append(unnamedRecords, rec)
		}
	}

	if len(unnamedRecords) == 0 {
		return 0, nil // No unnamed records
	}

	fmt.Printf("Found %d unnamed record(s). Please provide names:\n", len(unnamedRecords))

	// Prompt for each unnamed record
	for i, rec := range unnamedRecords {
		fmt.Printf("\nRecord %d/%d:\n", i+1, len(unnamedRecords))
		name, err := sm.promptForRecordName(rec)
		if err != nil {
			return 0, fmt.Errorf("could not get name for record: %w", err)
		}

		rec.Name = storagelib.Slugify(name)

		// Save record metadata
		if err := sm.storage.SaveRecordMetadata(rec); err != nil {
			return 0, fmt.Errorf("could not save record metadata: %w", err)
		}
	}

	// Save updated session
	if err := sm.saveSession(session); err != nil {
		return 0, fmt.Errorf("could not save session after naming records: %w", err)
	}

	return len(unnamedRecords), nil
}

func (sm *SessionManager) getCurrentRecordFromSession(session *Session) *record.Record {
	if session == nil {
		return nil
	}

	for _, r := range session.Records {
		if r.ID == session.CurrentRecordID {
			return r
		}
	}

	return nil
}

func (sm *SessionManager) getCurrentRecord() *record.Record {
	// Load current session to get current record
	session, err := sm.loadCurrentSession()
	if err != nil {
		return nil
	}

	return sm.getCurrentRecordFromSession(session)
}
