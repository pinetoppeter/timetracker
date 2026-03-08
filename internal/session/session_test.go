package session

import (
	"testing"
	"time"

	"github.com/pinetoppeter/timetracker/internal/record"
	"github.com/stretchr/testify/assert"
)

// TestSessionErrors tests the error constants
func TestSessionErrors(t *testing.T) {
	// Test that the error constants are properly defined
	assert.Error(t, ErrNoActiveSession)
	assert.Equal(t, "no active session", ErrNoActiveSession.Error())

	assert.Error(t, ErrSessionPaused)
	assert.Equal(t, "session is paused", ErrSessionPaused.Error())
}

// TestSessionStruct tests the basic session structure
func TestSessionStruct(t *testing.T) {
	// Test creating a session with basic fields
	now := time.Now()
	session := &Session{
		ID:              "test-session-123",
		StartTime:       now,
		EndTime:         nil, // Session is active
		Records:         make([]*record.Record, 0),
		CurrentRecordID: "record-1",
		IsPaused:        false,
	}

	assert.Equal(t, "test-session-123", session.ID)
	assert.Equal(t, now, session.StartTime)
	assert.Nil(t, session.EndTime)
	assert.Equal(t, 0, len(session.Records))
	assert.Equal(t, "record-1", session.CurrentRecordID)
	assert.False(t, session.IsPaused)
}

// TestSessionStartResultStruct tests the SessionStartResult structure
func TestSessionStartResultStruct(t *testing.T) {
	now := time.Now()
	session := &Session{
		ID:        "test-session",
		StartTime: now,
	}

	result := &SessionStartResult{
		Session:      session,
		RecordName:   "test-record",
		IsNewSession: true,
	}

	assert.Equal(t, session, result.Session)
	assert.Equal(t, "test-record", result.RecordName)
	assert.True(t, result.IsNewSession)
}

// TestSessionEndResultStruct tests the SessionEndResult structure
func TestSessionEndResultStruct(t *testing.T) {
	now := time.Now()
	session := &Session{
		ID:        "test-session",
		StartTime: now,
	}

	result := &SessionEndResult{
		Session:      session,
		RecordName:   "test-record",
		RecordsNamed: 3,
	}

	assert.Equal(t, session, result.Session)
	assert.Equal(t, "test-record", result.RecordName)
	assert.Equal(t, 3, result.RecordsNamed)
}

// TestRecordHelperFunctions tests the record helper functions
func TestRecordHelperFunctions(t *testing.T) {
	// Test creating a record
	now := time.Now()
	rec := record.NewRecord("test-record", now)

	assert.NotEmpty(t, rec.ID)
	assert.Equal(t, "test-record", rec.Name)
	assert.Equal(t, now, rec.StartTime)
	assert.Nil(t, rec.EndTime)
	assert.NotNil(t, rec.Metadata)

	// Test record duration calculation
	rec.StartTime = now.Add(-time.Hour)
	rec.EndTime = nil // Still running
	duration := rec.Duration()
	assert.True(t, duration > time.Hour-5*time.Second)
	assert.True(t, duration < time.Hour+5*time.Second)

	// Test with ended record
	endTime := now.Add(-30 * time.Minute)
	rec.EndTime = &endTime
	duration = rec.Duration()
	assert.Equal(t, 30*time.Minute, duration)

	// Test IsRunning
	assert.False(t, rec.IsRunning())
	rec.EndTime = nil
	assert.True(t, rec.IsRunning())
}

// TestSessionWithRecords tests session with multiple records
func TestSessionWithRecords(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-2 * time.Hour)

	// Create records
	rec1 := &record.Record{
		ID:        "record-1",
		Name:      "first-record",
		StartTime: startTime,
		EndTime:   nil,
	}

	rec2 := &record.Record{
		ID:        "record-2",
		Name:      "second-record",
		StartTime: startTime.Add(30 * time.Minute),
		EndTime:   nil,
	}

	// Create session with records
	session := &Session{
		ID:              "test-session",
		StartTime:       startTime,
		Records:         []*record.Record{rec1, rec2},
		CurrentRecordID: "record-2",
		IsPaused:        false,
	}

	assert.Equal(t, 2, len(session.Records))
	assert.Equal(t, "record-2", session.CurrentRecordID)
	assert.Equal(t, "second-record", session.Records[1].Name)
}

// TestSessionStateTransitions tests session state changes
func TestSessionStateTransitions(t *testing.T) {
	now := time.Now()

	// Test active session
	session := &Session{
		ID:        "test-session",
		StartTime: now,
		EndTime:   nil,
		IsPaused:  false,
	}

	assert.Nil(t, session.EndTime)
	assert.False(t, session.IsPaused)

	// Test paused session
	session.IsPaused = true
	assert.True(t, session.IsPaused)

	// Test ended session
	endTime := now.Add(1 * time.Hour)
	session.EndTime = &endTime
	session.IsPaused = false
	assert.NotNil(t, session.EndTime)
	assert.Equal(t, endTime, *session.EndTime)
}
