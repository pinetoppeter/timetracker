package record

import (
	"fmt"
	"time"
)

type Record struct {
	ID      string
	Name    string
	StartTime time.Time
	EndTime *time.Time
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func NewRecord(name string, startTime time.Time) *Record {
	return &Record{
		ID:        fmt.Sprintf("record-%d", startTime.UnixNano()),
		Name:      name,
		StartTime: startTime,
		EndTime:   nil,
		Metadata:  make(map[string]interface{}),
	}
}

func (r *Record) Duration() time.Duration {
	endTime := time.Now()
	if r.EndTime != nil {
		endTime = *r.EndTime
	}
	return endTime.Sub(r.StartTime)
}

func (r *Record) IsRunning() bool {
	return r.EndTime == nil
}