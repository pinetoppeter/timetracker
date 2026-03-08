package session

import (
	"fmt"
	"sync"

	"github.com/pinetoppeter/timetracker/internal/config"
	"github.com/pinetoppeter/timetracker/internal/storage"
)

var (
	sessionManager      *SessionManager
	sessionManagerMutex sync.Mutex
)

func getSessionManager() (*SessionManager, error) {
	sessionManagerMutex.Lock()
	defer sessionManagerMutex.Unlock()

	if sessionManager != nil {
		return sessionManager, nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("could not load config: %w", err)
	}

	store, err := storage.NewStorage()
	if err != nil {
		return nil, fmt.Errorf("could not initialize storage: %w", err)
	}

	sessionManager = NewSessionManager(cfg, store)
	return sessionManager, nil
}

func StartSession() (*SessionStartResult, error) {
	mgr, err := getSessionManager()
	if err != nil {
		return nil, err
	}
	result, err := mgr.StartSession()
	if err != nil {
		return nil, err
	}

	// Ensure we have the latest session data
	if result == nil {
		// This shouldn't happen, but just in case
		sess, err := mgr.loadCurrentSession()
		if err != nil {
			return nil, err
		}
		currentRecord := mgr.getCurrentRecord()
		recordName := ""
		if currentRecord != nil && currentRecord.Name != "" {
			recordName = currentRecord.Name
		}
		return &SessionStartResult{
			Session:      sess,
			RecordName:   recordName,
			IsNewSession: false,
		}, nil
	}

	return result, nil
}

func StartSessionWithName(name string) (*SessionStartResult, error) {
	mgr, err := getSessionManager()
	if err != nil {
		return nil, err
	}
	result, err := mgr.StartSessionWithName(name)
	if err != nil {
		return nil, err
	}

	// Ensure we have the latest session data
	if result == nil {
		// This shouldn't happen, but just in case
		sess, err := mgr.loadCurrentSession()
		if err != nil {
			return nil, err
		}
		currentRecord := mgr.getCurrentRecord()
		recordName := ""
		if currentRecord != nil && currentRecord.Name != "" {
			recordName = currentRecord.Name
		}
		return &SessionStartResult{
			Session:      sess,
			RecordName:   recordName,
			IsNewSession: false,
		}, nil
	}

	return result, nil
}

func PauseSession() error {
	// Load current session first to ensure we have the latest state
	mgr, err := getSessionManager()
	if err != nil {
		return err
	}
	_, err = mgr.loadCurrentSession()
	if err != nil {
		return err
	}
	return mgr.PauseSession()
}

func SwitchRecord() error {
	// Load current session first to ensure we have the latest state
	mgr, err := getSessionManager()
	if err != nil {
		return err
	}
	_, err = mgr.loadCurrentSession()
	if err != nil {
		return err
	}
	return mgr.SwitchRecord()
}

func SwitchAndNameRecord(name string) error {
	// Load current session first to ensure we have the latest state
	mgr, err := getSessionManager()
	if err != nil {
		return err
	}
	_, err = mgr.loadCurrentSession()
	if err != nil {
		return err
	}
	return mgr.SwitchAndNameRecord(name)
}

func ResumeLastRecord() error {
	// Load current session first to ensure we have the latest state
	mgr, err := getSessionManager()
	if err != nil {
		return err
	}
	_, err = mgr.loadCurrentSession()
	if err != nil {
		return err
	}
	return mgr.ResumeLastRecord()
}

func NameCurrentRecord(name string) error {
	// Load current session first to ensure we have the latest state
	mgr, err := getSessionManager()
	if err != nil {
		return err
	}
	_, err = mgr.loadCurrentSession()
	if err != nil {
		return err
	}
	return mgr.NameCurrentRecord(name)
}

func GetCurrentSession() (*Session, error) {
	// Load current session first to ensure we have the latest state
	mgr, err := getSessionManager()
	if err != nil {
		return nil, err
	}
	sess, err := mgr.loadCurrentSession()
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func EndSession() (*SessionEndResult, error) {
	// Load current session first to ensure we have the latest state
	mgr, err := getSessionManager()
	if err != nil {
		return nil, err
	}
	_, err = mgr.loadCurrentSession()
	if err != nil {
		return nil, err
	}
	return mgr.EndSession()
}

func EndSessionWithRecordNames(names []string) (*SessionEndResult, error) {
	// Load current session first to ensure we have the latest state
	mgr, err := getSessionManager()
	if err != nil {
		return nil, err
	}
	_, err = mgr.loadCurrentSession()
	if err != nil {
		return nil, err
	}
	return mgr.EndSessionWithRecordNames(names)
}

func AddMetadataToCurrentRecord(key, value string) error {
	// Load current session first to ensure we have the latest state
	mgr, err := getSessionManager()
	if err != nil {
		return err
	}
	_, err = mgr.loadCurrentSession()
	if err != nil {
		return err
	}
	return mgr.AddMetadataToCurrentRecord(key, value)
}

func AddMetadataToRecord(recordName, key, value string) error {
	mgr, err := getSessionManager()
	if err != nil {
		return err
	}
	return mgr.AddMetadataToRecord(recordName, key, value)
}

func GetCurrentRecordMetadata() (map[string]interface{}, error) {
	// Load current session first to ensure we have the latest state
	mgr, err := getSessionManager()
	if err != nil {
		return nil, err
	}
	_, err = mgr.loadCurrentSession()
	if err != nil {
		return nil, err
	}
	return mgr.GetCurrentRecordMetadata()
}

func GetRecordMetadata(recordName string) (map[string]interface{}, error) {
	mgr, err := getSessionManager()
	if err != nil {
		return nil, err
	}
	return mgr.GetRecordMetadata(recordName)
}
